package request

import (
	"cmp"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
	ext "stresstester/extentions"
	"stresstester/flowcontextloader"
	"stresstester/messagehandler"
	"stresstester/webservicehandler"
)

func Doprecalls(Precalls *[]flowcontextloader.Precall, context map[string]string) error {
	for _, pc := range *Precalls {

		absPath, err := GetRealPath(pc.Messagefilepath)
		if err != nil {
			log.Printf("Error getting Message template abs path: %v, relative pathe used.\n", err)
		}

		messageTemp, err := messagehandler.Getmessagetemplate(absPath)
		if err != nil {
			return err
		}

		finalmessage, err := messagehandler.Generatefinalmessage(messageTemp, context)
		if err != nil {
			return err
		}

		if pc.Endpointtype == "SOAPWebService" {
			err, result := webservicehandler.SendSoapRequest(pc.Endpointurl, pc.Httpmethod, pc.Timeout, finalmessage, pc.Name, pc.Statussuccessvalue, pc.Statusxpath, pc.Variables)
			if err != nil {
				return err
			}
			maps.Copy(context, result)
		}
	}

	return nil

}

func ApplyExtentions(message string, Wholemessageprocesses *[]flowcontextloader.Wholemessageprocess) (string, error) {

	slices.SortFunc[[]flowcontextloader.Wholemessageprocess](*Wholemessageprocesses, func(a, b flowcontextloader.Wholemessageprocess) int {
		return cmp.Compare(a.Order, b.Order)
	})

	for _, extention := range *Wholemessageprocesses {
		err, extensiontype := ext.Extentionpicker(extention.Process)
		if err != nil {

			return message, err
		}
		message, err = extensiontype.Process(message, extention.Params)
		if err != nil {
			return message, err
		}
	}
	return message, nil
}
func GetRealPath(relativepath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return relativepath, err
	}
	absPath := filepath.Join(wd, relativepath)
	return absPath, nil
}
