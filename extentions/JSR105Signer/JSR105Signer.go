package JSR105Signer

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/beevik/etree"
	keystore "github.com/pavlo-v-chernykh/keystore-go/v4"
	dsig "github.com/russellhaering/goxmldsig"
)

const (
	xmlDSigNS       = "http://www.w3.org/2000/09/xmldsig#"
	algC14N11       = "http://www.w3.org/2006/12/xml-c14n11"
	algRSASHA256    = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
	algEnvelopedSig = "http://www.w3.org/2000/09/xmldsig#enveloped-signature"
	algSHA256Digest = "http://www.w3.org/2001/04/xmlenc#sha256"

	// WSS4J XMLUtils.outputDOM(doc, os, true) writes exactly this + '\n'
	javaPreamble = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
)

// ---------------------------------------------------------------------------
// Main flow — mirrors the Java testSignatureUsingJSR105
// ---------------------------------------------------------------------------
type JSR105SignerExtention struct{}

func (JSE JSR105SignerExtention) Name() string {
	return "JSR105SignerExtention"
}

func (JSE JSR105SignerExtention) Process(data string, param map[string]string) (string, error) {
	if param == nil {
		param = make(map[string]string)
	}
	reader := strings.NewReader(data)
	var signed bytes.Buffer
	err := signatureUsingJSR105(reader, &signed, param)
	if err != nil {
		return "", err
	}
	return signed.String(), nil
}

func signatureUsingJSR105(sourceDocument io.Reader, out io.Writer, param map[string]string) error {
	// 1) Parse — keep the tree EXACTLY as parsed. Java's DOM keeps
	// whitespace-only text nodes and Santuario's C14N hashes them, so
	// preserving input whitespace is required for byte-identical digests.
	doc := etree.NewDocument()
	if _, err := doc.ReadFrom(sourceDocument); err != nil {
		return fmt.Errorf("failed to read XML document: %w", err)
	}

	// 2) Load JKS keystore
	ksPath, ok := param["KEYSTORE_LOCATION"]
	if !ok {
		ksPath = "files\\ips.jks"
	}

	storePassword, ok := param["KEYSTORE_PASSWORD"]
	if !ok {
		storePassword = "password2" // demo only — use a secret manager in production
	}

	keyPassword, ok := param["KEY_PASSWORD"]
	if !ok {
		keyPassword = "password2"
	}

	alias, ok := param["KEY_ALIAS"]
	if !ok {
		keyPassword = "ips"
	}

	abskspath, err := getRealPath(ksPath)
	if err != nil {
		log.Printf("Error getting Message template abs path: %v, relative pathe used.\n", err)
	}

	privKey, certs, err := loadKeyAndCertFromJKS(abskspath, storePassword, keyPassword, alias)
	if err != nil {
		return fmt.Errorf("cannot load key/cert from JKS: %w", err)
	}

	// 3) Sign
	localPart, ok := param["LOCAL_PART"]
	if !ok {
		keyPassword = "Message"
	}

	if err := signUsingJSR105(doc, privKey, certs[0], localPart); err != nil {
		return fmt.Errorf("signing failed: %w", err)
	}

	// 4) Verify FIRST — on the same tree state that was digested.
	// Anything that mutates the tree (attribute sorting, etc.) must come after.
	if err := verifyUsingJSR105(doc, certs); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// 5) Output cosmetics — attribute order never affects C14N, so this is
	// digest-safe (and runs after verification anyway).
	sortXMLNSAttributes(doc.Root())

	// 6) Write with Java's preamble; input whitespace is preserved by etree,
	// matching Java's DOM identity-transform serialization.

	if err := writeJavaStyle(doc, out); err != nil {
		return fmt.Errorf("cannot write signed XML: %w", err)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Signing — mirrors SignatureUtils.signUsingJSR105
// ---------------------------------------------------------------------------

func signUsingJSR105(doc *etree.Document, privKey *rsa.PrivateKey, signingCert *x509.Certificate, localPart string) error {
	c14n := dsig.MakeC14N11Canonicalizer()

	sigParent := findFirstElementByTag(doc.Root(), "Sgntr")
	if sigParent == nil {
		return errors.New("cannot locate <Sgntr> element to attach the signature")
	}
	elementsToSign := findAllElementsByTag(doc.Root(), localPart)
	if len(elementsToSign) == 0 {
		return fmt.Errorf("no elements matching local-name '%s' were found", localPart)
	}

	// DigestValue: SHA-256 over C14N-11 of the whole document (URI=""),
	// computed BEFORE the Signature exists (the enveloped-signature transform
	// removes it again at verification time).
	docCanonical, err := c14n.Canonicalize(doc.Root())
	if err != nil {
		return fmt.Errorf("canonicalization of document failed: %w", err)
	}
	docDigest := sha256.Sum256(docCanonical)
	digestB64 := base64.StdEncoding.EncodeToString(docDigest[:])

	// <Signature xmlns="http://www.w3.org/2000/09/xmldsig#"> — DEFAULT
	// namespace, exactly like Santuario marshals it. A "ds:" prefix would
	// change the canonicalized SignedInfo bytes under inclusive C14N-11 and
	// therefore the SignatureValue.
	sigEl := etree.NewElement("Signature")
	sigEl.CreateAttr("xmlns", xmlDSigNS)

	signedInfo := dsChild(sigEl, "SignedInfo")
	dsChild(signedInfo, "CanonicalizationMethod").CreateAttr("Algorithm", algC14N11)
	dsChild(signedInfo, "SignatureMethod").CreateAttr("Algorithm", algRSASHA256)

	for range elementsToSign {
		ref := dsChild(signedInfo, "Reference")
		ref.CreateAttr("URI", "")
		transforms := dsChild(ref, "Transforms")
		dsChild(transforms, "Transform").CreateAttr("Algorithm", algEnvelopedSig)
		dsChild(transforms, "Transform").CreateAttr("Algorithm", algC14N11)
		dsChild(ref, "DigestMethod").CreateAttr("Algorithm", algSHA256Digest)
		dsChild(ref, "DigestValue").SetText(digestB64)
	}

	sigValueEl := dsChild(sigEl, "SignatureValue")

	keyInfo := dsChild(sigEl, "KeyInfo")
	x509Data := dsChild(keyInfo, "X509Data")
	dsChild(x509Data, "X509Certificate").SetText(santuarioBase64(signingCert.Raw))
	issuerSerial := dsChild(x509Data, "X509IssuerSerial")
	dsChild(issuerSerial, "X509IssuerName").SetText(signingCert.Issuer.String())
	dsChild(issuerSerial, "X509SerialNumber").SetText(signingCert.SerialNumber.String())

	// Attach at final location (same placement as DOMSignContext)
	sigParent.AddChild(sigEl)

	// SignedInfo: canonicalize the subtree with ALL in-scope namespaces on its
	// apex element — exactly what JSR-105 does before signing.
	siCanonical, err := canonicalizeSubtree(c14n, signedInfo)
	if err != nil {
		return fmt.Errorf("canonicalization of SignedInfo failed: %w", err)
	}
	if os.Getenv("DSIG_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "SIGNEDINFO_C14N:\n%s\n", siCanonical)
		fmt.Fprintf(os.Stderr, "SIGNEDINFO_C14N_B64=%s\n",
			base64.StdEncoding.EncodeToString(siCanonical))
	}
	siDigest := sha256.Sum256(siCanonical)
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, siDigest[:])
	if err != nil {
		return fmt.Errorf("RSA signing failed: %w", err)
	}
	sigValueEl.SetText(santuarioBase64(sigBytes))
	return nil
}

// canonicalizeSubtree reproduces C14N subtree semantics: every namespace
// binding in scope at the element's position is emitted on its apex element.
// We copy the element, pin the in-scope bindings onto the copy, and
// canonicalize the detached copy — independent of how the library resolves
// namespaces above the canonicalized node. Used symmetrically by sign+verify.
func canonicalizeSubtree(c14n dsig.Canonicalizer, el *etree.Element) ([]byte, error) {
	cp := el.Copy()
	for prefix, uri := range inScopeNamespaces(el) {
		if prefix == "" {
			cp.CreateAttr("xmlns", uri)
		} else {
			cp.CreateAttr("xmlns:"+prefix, uri)
		}
	}
	return c14n.Canonicalize(cp)
}

// inScopeNamespaces walks root→el collecting xmlns/xmlns:* declarations with
// proper shadowing (deeper declarations override shallower ones).
func inScopeNamespaces(el *etree.Element) map[string]string {
	var path []*etree.Element
	for e := el; e != nil; e = e.Parent() {
		path = append([]*etree.Element{e}, path...)
	}
	m := make(map[string]string)
	for _, e := range path {
		for _, a := range e.Attr {
			switch {
			case a.Space == "xmlns":
				m[a.Key] = a.Value
			case a.Space == "" && a.Key == "xmlns":
				m[""] = a.Value
			}
		}
	}
	delete(m, "xml") // implicit binding, C14N handles it specially
	return m
}

// ---------------------------------------------------------------------------
// Verification — mirrors SignatureUtils.verifyUsingJSR105
// ---------------------------------------------------------------------------

func verifyUsingJSR105(doc *etree.Document, chain []*x509.Certificate) error {
	c14n := dsig.MakeC14N11Canonicalizer()

	sigEl := findFirstElementByTag(doc.Root(), "Signature")
	if sigEl == nil {
		return errors.New("no <Signature> element found in document")
	}
	signedInfo := findFirstElementByTag(sigEl, "SignedInfo")
	if signedInfo == nil {
		return errors.New("signature is missing <SignedInfo>")
	}
	ref := findFirstElementByTag(signedInfo, "Reference")
	if ref == nil {
		return errors.New("signature is missing <Reference>")
	}
	digestValueEl := findFirstElementByTag(ref, "DigestValue")
	sigValueEl := findFirstElementByTag(sigEl, "SignatureValue")
	if digestValueEl == nil || sigValueEl == nil {
		return errors.New("signature is missing required elements")
	}

	// (1) Reference digest: URI="" + enveloped-signature transform
	raw, err := doc.WriteToString()
	if err != nil {
		return err
	}
	docCopy := etree.NewDocument()
	if err := docCopy.ReadFromString(raw); err != nil {
		return err
	}
	removeSignatureElements(docCopy.Root()) // the enveloped transform

	docCanonical, err := c14n.Canonicalize(docCopy.Root())
	if err != nil {
		return err
	}
	docDigest := sha256.Sum256(docCanonical)
	want, err := base64.StdEncoding.DecodeString(stripBase64Whitespace(digestValueEl.Text()))
	if err != nil {
		return fmt.Errorf("DigestValue is not valid base64: %w", err)
	}
	if subtle.ConstantTimeCompare(docDigest[:], want) != 1 {
		return errors.New("reference digest mismatch (document changed after signing?)")
	}

	// (2) SignatureValue: RSA-SHA256 over canonicalized SignedInfo
	siCanonical, err := canonicalizeSubtree(c14n, signedInfo)
	if err != nil {
		return err
	}
	siDigest := sha256.Sum256(siCanonical)
	// The wrapped base64 contains line breaks (\r\n, normalized to \n when
	// re-parsed); Go's decoder rejects them, so strip whitespace first.
	sigBytes, err := base64.StdEncoding.DecodeString(stripBase64Whitespace(sigValueEl.Text()))
	if err != nil {
		return fmt.Errorf("SignatureValue is not valid base64: %w", err)
	}
	pub, ok := chain[0].PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("certificate does not carry an RSA public key")
	}
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, siDigest[:], sigBytes); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// JKS loading — keystore-go/v4: New() + Load() + GetPrivateKeyEntry()
// ---------------------------------------------------------------------------

func loadKeyAndCertFromJKS(path, storePassword, keyPassword, alias string) (*rsa.PrivateKey, []*x509.Certificate, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open keystore file %s: %w", path, err)
	}
	defer f.Close()

	ks := keystore.New()
	if err := ks.Load(f, []byte(storePassword)); err != nil {
		return nil, nil, fmt.Errorf("keystore load error: %w", err)
	}

	entry, err := ks.GetPrivateKeyEntry(alias, []byte(keyPassword))
	if err != nil {
		return nil, nil, fmt.Errorf("private key entry %q: %w", alias, err)
	}

	privAny, err := x509.ParsePKCS8PrivateKey(entry.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse private key: %w", err)
	}
	priv, ok := privAny.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, errors.New("private key is not RSA")
	}

	if len(entry.CertificateChain) == 0 {
		return nil, nil, errors.New("no certificate chain in keystore entry")
	}
	certs := make([]*x509.Certificate, 0, len(entry.CertificateChain))
	for i, c := range entry.CertificateChain {
		cert, err := x509.ParseCertificate(c.Content)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot parse certificate %d: %w", i, err)
		}
		certs = append(certs, cert)
	}
	return priv, certs, nil
}

// ---------------------------------------------------------------------------
// Java-serializer emulation helpers
// ---------------------------------------------------------------------------

// dsChild creates an unprefixed child — it inherits the Signature element's
// default xmlns, exactly like Santuario's output.
func dsChild(parent *etree.Element, tag string) *etree.Element {
	return parent.CreateElement(tag)
}

// santuarioBase64 replicates Santuario's base64 encoder: 76 chars per line,
// joined with CRLF. The serializer then escapes the CR as &#xD;.
func santuarioBase64(data []byte) string {
	enc := base64.StdEncoding.EncodeToString(data)
	var parts []string
	for len(enc) > 76 {
		parts = append(parts, enc[:76])
		enc = enc[76:]
	}
	return strings.Join(append(parts, enc), "\r\n")
}

// stripBase64Whitespace removes line breaks/spaces that XML Schema
// base64Binary allows inside encoded values.
func stripBase64Whitespace(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\r', '\n', ' ', '\t':
			return -1
		}
		return r
	}, s)
}

// sortXMLNSAttributes moves xmlns declarations to the front of each element,
// sorted lexicographically by prefix — the order Xalan emits them in.
func sortXMLNSAttributes(el *etree.Element) {
	var ns, other []etree.Attr
	for _, a := range el.Attr {
		if isXMLNS(a) {
			ns = append(ns, a)
		} else {
			other = append(other, a)
		}
	}
	sort.Slice(ns, func(i, j int) bool { return xmlnsKey(ns[i]) < xmlnsKey(ns[j]) })
	el.Attr = append(ns, other...)
	for _, c := range el.ChildElements() {
		sortXMLNSAttributes(c)
	}
}

func isXMLNS(a etree.Attr) bool {
	return a.Space == "xmlns" || (a.Space == "" && a.Key == "xmlns")
}

func xmlnsKey(a etree.Attr) string {
	if a.Space == "xmlns" {
		return "xmlns:" + a.Key
	}
	return a.Key
}

// writeJavaStyle drops the input's XML declaration and writes Java's preamble
// plus the document. The &#xD; replacement guarantees Santuario-style CR
// escaping in the wrapped base64 regardless of etree's escaping behavior.
func writeJavaStyle(doc *etree.Document, out io.Writer) error {
	kept := make([]etree.Token, 0, len(doc.Child))
	for _, t := range doc.Child {
		if pi, ok := t.(*etree.ProcInst); ok && pi.Target == "xml" {
			continue
		}
		kept = append(kept, t)
	}
	doc.Child = kept

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		return err
	}
	s := strings.ReplaceAll(buf.String(), "\r\n", "&#xD;\n")
	if _, err := io.WriteString(out, javaPreamble); err != nil {
		return err
	}
	_, err := io.WriteString(out, s)
	return err
}

// ---------------------------------------------------------------------------
// Element search by local tag name (namespace-agnostic)
// ---------------------------------------------------------------------------

func findAllElementsByTag(root *etree.Element, tag string) []*etree.Element {
	var out []*etree.Element
	var walk func(*etree.Element)
	walk = func(e *etree.Element) {
		if e.Tag == tag {
			out = append(out, e)
		}
		for _, c := range e.ChildElements() {
			walk(c)
		}
	}
	walk(root)
	return out
}

func findFirstElementByTag(root *etree.Element, tag string) *etree.Element {
	if list := findAllElementsByTag(root, tag); len(list) > 0 {
		return list[0]
	}
	return nil
}

// removeSignatureElements implements the enveloped-signature transform:
// every <Signature> element is removed from the tree.
func removeSignatureElements(el *etree.Element) {
	var toRemove []*etree.Element
	for _, c := range el.ChildElements() {
		if c.Tag == "Signature" {
			toRemove = append(toRemove, c)
		} else {
			removeSignatureElements(c)
		}
	}
	for _, c := range toRemove {
		el.RemoveChild(c)
	}
}

func getRealPath(relativepath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return relativepath, err
	}
	absPath := filepath.Join(wd, relativepath)
	return absPath, nil
}
