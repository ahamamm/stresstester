package extentions

import "errors"

type NoExtentionSelected struct{}

func (JSE NoExtentionSelected) Name() string {
	return "NoExtentionSelected"
}

func (JSE NoExtentionSelected) Process(data string, param map[string]string) (string, error) {
	return "", errors.New("No extention selected")
}
