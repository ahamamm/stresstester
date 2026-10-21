package utils

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

func StringFromPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func IntFromPtr(ptr *int) int {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func GetRealPath(relativepath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return relativepath, err
	}
	absPath := filepath.Join(wd, relativepath)
	return absPath, nil
}

func Decodebase64(encodedstring string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedstring)
	if err != nil {
		fmt.Println("Error decoding string:", err)
		return encodedstring, err
	}

	// Convert []byte back to string
	decodedStr := string(decodedBytes)
	return decodedStr, nil
}
