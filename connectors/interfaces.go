package connectors

type Connectorinterface interface {
	Name() string
	Connect() error
	Close() error
	Sendmessage(message string, headers map[string]string) error
}

type IBMMQconnector struct {
	Connectionstring string
	Queuemanager     string
	Queuename        string
}

type NoConnectorSelected struct{}
