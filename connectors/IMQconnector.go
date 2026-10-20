package connectors

import (
	mqutil "stresstester/mqutil"
)

func (imc IBMMQconnector) Name() string {
	return "Ibmmqconnector"
}

func (imc IBMMQconnector) Connect() error {

	return mqutil.Init(imc.Connectionstring, imc.Queuemanager, imc.Queuename)
}

func (imc IBMMQconnector) Close() error {
	mqutil.CloseConn()
	return nil
}

func (imc IBMMQconnector) Sendmessage(message string, headers map[string]string) error {
	err := mqutil.Putmessage(message)
	if err != nil {
		return err
	}
	return nil
}
