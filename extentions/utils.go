package extentions

import (
	"errors"
	"stresstester/extentions/JSR105Signer"
)

func Extentionpicker(name string) (error, Extentioninterface) {
	var ext Extentioninterface
	var err error

	switch name {
	case "JSR105SignerExtention":
		ext = JSR105Signer.JSR105SignerExtention{}
	default:
		err = errors.New("No extention selected")
		ext = NoExtentionSelected{}
		return err, ext
	}
	return err, ext
}
