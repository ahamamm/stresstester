package connectors

import (
	mqutil "stresstester/mqutil"
)

func (imc IBMMQconnector) Name() string {
	return "Ibmmqconnector"
}

func (imc IBMMQconnector) Connect() error {

	return mqutil.Init(imc.Ibmmqconninfo.Channelname, imc.Ibmmqconninfo.Connectionname, imc.Ibmmqconninfo.Queuemanager, imc.Ibmmqconninfo.Queuename, imc.Ibmmqconninfo.Userid, imc.Ibmmqconninfo.Password)
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
