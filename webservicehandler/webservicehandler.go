package webservicehandler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"stresstester/flowcontextloader"

	"github.com/antchfx/xmlquery"
)

func SendSoapRequest(url string, method string, intimeout int, payload string, callname string, statussuccessvalue string, statusxpath string, returnvariables []flowcontextloader.Variable) (error, map[string]string) {
	var ftimeout int

	if intimeout >= 1 {
		ftimeout = intimeout
	} else {
		ftimeout = 10
	}
	client := &http.Client{

		Timeout: time.Duration(ftimeout) * time.Second,
	}

	req, err := http.NewRequest(method, url, bytes.NewBufferString(payload))
	if err != nil {
		return err, nil
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close() // Ensure the body is closed to prevent memory leaks

	// 7. Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		str := fmt.Sprintf("Server returned error %d: %s payload is %s", resp.StatusCode, string(bodyBytes), payload)

		return errors.New(str), nil
	}

	// 8. Parse the XML Response stream directly via xmlquery
	doc, err := xmlquery.Parse(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse XML response: %v", err))
	}

	status := xmlquery.FindOne(doc, statusxpath)
	if status.InnerText() != statussuccessvalue {
		bodyBytes, _ := io.ReadAll(resp.Body)
		str := fmt.Sprintf("unsuccessul call to precall callname : %s \n  status: %s\n request:%s \n response: %s \n %v", callname, status.InnerText(), payload, string(bodyBytes), err)
		return errors.New(str), nil
	}

	resultmap := make(map[string]string)

	for _, vrbl := range returnvariables {
		val := xmlquery.FindOne(doc, vrbl.Variablexpath)
		resultmap[vrbl.Variablename] = val.InnerText()
	}

	return nil, resultmap

}
