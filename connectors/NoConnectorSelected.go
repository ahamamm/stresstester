package connectors

import (
	"errors"
)

func (imc NoConnectorSelected) Name() string {
	return "NoConnectorSelected"
}

func (imc NoConnectorSelected) Connect() error {
	return errors.New("Wrong Endpoint type, No Connector selected, check your parameters.")
}

func (imc NoConnectorSelected) Close() error {
	return errors.New("Wrong Endpoint type, No Connector selected, check your parameters.")
}

func (imc NoConnectorSelected) Sendmessage(message string, headers map[string]string) error {
	return errors.New("Wrong Endpoint type, No Connector selected, check your parameters.")
}
