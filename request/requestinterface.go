package request

import "stresstester/flowcontextloader"

type RequestImpl struct {
	MessageTemp             string
	PrecallList             []flowcontextloader.Precall
	WholemessageprocessList []flowcontextloader.Wholemessageprocess
	RequestContext          map[string]string
	Message                 string
	ProcessedMessage        string
}

/*type RequestImpl struct {
	MessageTemp      string
	ConnectionParam  *jsonreader.Jsonmapper
	RequestContext   map[string]string
	Message          string
	ProcessedMessage string
}*/

type Requestinterface interface {
	Precalls() error
	Messagegeneration() error
	Wholemessageprocess() error
}
