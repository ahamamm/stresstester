package utils

import (
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
