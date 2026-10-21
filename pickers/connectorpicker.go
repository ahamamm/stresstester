package pickers

import (
	"errors"
	"stresstester/connectors"
	"stresstester/flowcontextloader"
)

func Connectorpicker(contextmapper flowcontextloader.Contextmapper) (error, connectors.Connectorinterface) {
	var ci connectors.Connectorinterface
	var err error

	if contextmapper.Endpointtype == "IBMQM" {
		ci = connectors.IBMMQconnector{Ibmmqconninfo: contextmapper.Ibmmqconninfo}
	} else {
		ci = connectors.NoConnectorSelected{}
		err = errors.New("Wrong Endpoint type, No Connector selected, check your parameters.")

	}
	return err, ci
}
