package flowcontextloader

import (
	"encoding/json"
	"os"
)

type Contextmapper struct {
	Name                  string                `json:"name"`
	Endpointtype          string                `json:"endpointtype"`
	Connectionstring      string                `json:"connectionstring"`
	Queuemanager          string                `json:"queuemanager"`
	Queuename             string                `json:"queuename"`
	Messagefilepath       string                `json:"messagefilepath"`
	Precalls              []Precall             `json:"precalls,omitempty"`
	Wholemessageprocesses []Wholemessageprocess `json:"wholemessageprocesses,omitempty"`
}

type Precall struct {
	Order              int               `json:"order,omitempty"`
	Name               string            `json:"name,omitempty"`
	Endpointtype       string            `json:"endpointtype,omitempty"`
	Endpointurl        string            `json:"endpointurl,omitempty"`
	Httpmethod         string            `json:"httpmethod,omitempty"`
	Messagefilepath    string            `json:"messagefilepath,omitempty"`
	Statusxpath        string            `json:"statusxpath,omitempty"`
	Statussuccessvalue string            `json:"statussuccessvalue,omitempty"`
	Timeout            int               `json:"timwout,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	Variables          []Variable        `json:"variables,omitempty"`
}
type Wholemessageprocess struct {
	Process string            `json:"process,omitempty"`
	Order   int               `json:"order,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
}

/*
sample JSON #1 IPS outbound flow using pain1

{
  "name": "Test Template 2",
  "endpointtype": "IBMQM",
  "connectionstring": "APP1.SVRCONN/TCP/10.11.34.132(1425)",
  "queuemanager": "QPFQMS01",
  "queuename": "TESTQUEUE",
  "messagefilepath": "messagetemplate\\temp1.txt",
  "precalls": [
    {
      "order": 1,
      "name": "bookentry-create",
      "endpointtype": "SOAPWebService",
      "endpointurl": "http://10.11.34.16:7837/AlRajhiCustAccount/BookedEntrySvcTrxn?",
      "httpmethod": "POST",
      "messagefilepath": "messagetemplate\\bookentry-create-precall.txt",
      "statusxpath": "/soapenv:Envelope/soapenv:Body/rca:BookedEntrySvcTrxnRs/Hdr/arb:Status/arb:StatusCd",
      "statussuccessvalue": "I000000",
      "returnxpath": "/soapenv:Envelope/soapenv:Body/rca:BookedEntrySvcTrxnRs/Body/BookingInfo/BookedEntryRefNum",
      "variables": [
        {
          "variablename": "bookentryref",
          "variablexpath": "/soapenv:Envelope/soapenv:Body/rca:BookedEntrySvcTrxnRs/Body/BookingInfo/BookedEntryRefNum"
        }
      ]
    }
  ]
}

sample JSON #2 IPS inbound flow using pacs8

{
  "name": "Test Template 2",
  "endpointtype": "IBMQM",
  "connectionstring": "APP1.SVRCONN/TCP/10.11.34.132(1425)",
  "queuemanager": "QPFQMS01",
  "queuename": "TESTQUEUE",
  "messagefilepath": "messagetemplate\\pacs8_template.txt",
  "Wholemessageprocesses": [
    {
      "order": 1,
      "process": "JSR105SignerExtention",
      "params": {
        "KEYSTORE_PASSWORD": "password2",
        "KEY_PASSWORD": "password2",
        "LOCAL_PART": "Message",
        "KEYSTORE_LOCATION":"D:\\code\\GoCode\\stresstester\\files\\ips.jks",
        "KEY_ALIAS":"ips"
      }
    }
  ]
}
*/

type Variable struct {
	Variablename  string `json:"variablename,omitempty"`
	Variablexpath string `json:"variablexpath,omitempty"`
}

func Mappjson(templatepath string) (Contextmapper, error) {
	jsonmapper := Contextmapper{}
	fileBytes, err := os.ReadFile(templatepath)
	if err != nil {
		return jsonmapper, err
	}

	err = json.Unmarshal(fileBytes, &jsonmapper)

	if err != nil {
		return jsonmapper, err
	}

	return jsonmapper, nil
}
