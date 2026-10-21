package connectors

import "stresstester/flowcontextloader"

type Connectorinterface interface {
	Name() string
	Connect() error
	Close() error
	Sendmessage(message string, headers map[string]string) error
}

type IBMMQconnector struct {
	Ibmmqconninfo flowcontextloader.Ibmmqconnectioninfo
}

type NoConnectorSelected struct{}
