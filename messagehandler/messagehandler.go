package messagehandler

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"uuid"
)

func Getmessagetemplate(messagetemplatepath string) (string, error) {

	// 1. Read the template file
	content, err := os.ReadFile(messagetemplatepath)
	if err != nil {
		return "", err
	}
	return string(content), nil

}
func Generatefinalmessage(messagetemplate string, keyvalues map[string]string) (string, error) {

	// 2. Process placeholders
	processedText := replacePlaceholders(messagetemplate, keyvalues)

	return processedText, nil
}

// Core parsing logic
func replacePlaceholders(input string, keyvalues map[string]string) string {
	// Matches patterns like [string:10], [float:10.5], or [date:YYYY-MM-DD]
	re := regexp.MustCompile(`\[([a-zA-Z0-9\-\_]+):(string|int|float|uuid|nduuid|date|compound):([a-zA-Z0-9\-\.\:\<\>]+)\]`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Strip brackets: "[string:10]" -> "string:10"
		cleanMatch := strings.NewReplacer("[", "", "]", "").Replace(match)
		parts := strings.SplitN(cleanMatch, ":", 3)

		name := parts[0]
		dataType := parts[1]
		format := parts[2]

		val, ok := keyvalues[name]
		if ok {
			return val
		}

		switch dataType {
		case "string":
			length, _ := strconv.Atoi(format)
			val := generateRandomString(length)
			keyvalues[name] = val
			return val
		case "int":
			length, _ := strconv.Atoi(format)
			val := generateRandomInt(length)
			keyvalues[name] = val
			return val
		case "float":
			floatParts := strings.Split(format, ".")
			totalDigits, _ := strconv.Atoi(floatParts[0])
			decimalDigits, _ := strconv.Atoi(floatParts[1])
			val := generateRandomFloat(totalDigits, decimalDigits)
			keyvalues[name] = val
			return val
		case "uuid":
			val := uuid.New().String()
			keyvalues[name] = val
			return val
		case "nduuid":
			val = strings.ReplaceAll(uuid.New().String(), "-", "")
			keyvalues[name] = val
			return val
		case "compound":
			val := generateCompound(format)
			keyvalues[name] = val
			return val
		case "date":
			val := formatCurrentDate(format)
			keyvalues[name] = val
			return val
		}

		return match
	})

}

// Generates a random alphanumeric string of a fixed length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// Generates a random integer of a fixed length
func generateRandomInt(length int) string {
	if length <= 0 || length > 18 {
		return "0" // int64 can reliably hold up to 18 digits safely without overflow
	}

	// Calculate range boundaries
	// e.g., for length 4: min = 1000, max = 9999
	min := int(math.Pow10(length - 1))
	max := int(math.Pow10(length)) - 1

	// rand.Int64N(code) returns a value in [0, code).
	// We offset it by min to fit our specific bounds.
	return strconv.Itoa(rand.Intn(max-min+1) + min)
}

// Generates a float matching a specific total layout and exact decimals
func generateRandomFloat(totalDigits, decimalDigits int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var intPart int
	intDigits := totalDigits - decimalDigits
	if intDigits <= 0 {
		intPart = 0
	} else {
		maxInt := int(math.Pow10(intDigits)) - 1
		minInt := int(math.Pow10(intDigits - 1))
		intPart = r.Intn(maxInt-minInt+1) + minInt
	}
	maxDec := int(math.Pow10(decimalDigits)) - 1
	decPart := r.Intn(maxDec + 1)

	formatStr := fmt.Sprintf("%%d.%%0%dd", decimalDigits)
	return fmt.Sprintf(formatStr, intPart, decPart)
}

func generateCompound(pattern string) string {
	re1 := regexp.MustCompile(`\<(int|date):([a-zA-Z0-9\:\.\-]+)\>`)

	return re1.ReplaceAllStringFunc(pattern, func(match1 string) string {
		cleanMatch := strings.NewReplacer("<", "", ">", "").Replace(match1)
		parts := strings.SplitN(cleanMatch, ":", 2)
		dataType := parts[0]
		format := parts[1]
		switch dataType {
		case "int":
			length, _ := strconv.Atoi(format)
			return generateRandomInt(length)
		case "date":
			return formatCurrentDate(format)
		}
		return match1
	})
}

// Converts standard human layout tokens into Go reference layouts
func formatCurrentDate(layout string) string {
	now := time.Now()

	// Map standard user formats to Go's layout standards
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"MM", "01",
		"DD", "02",
		"hh", "15",
		"mm", "04",
		"ss", "05",
		"ooo", "999",
	)

	goLayout := replacer.Replace(layout)
	return now.Format(goLayout)
}
