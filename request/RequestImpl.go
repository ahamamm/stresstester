package request

import (
	"stresstester/messagehandler"
)

func (RI *RequestImpl) Precalls() error {
	if RI.RequestContext == nil {
		RI.RequestContext = make(map[string]string)
	}
	err := Doprecalls(&RI.PrecallList, RI.RequestContext)
	if err != nil {
		return err
	}
	return nil
}

func (RI *RequestImpl) Messagegeneration() error {
	message, err := messagehandler.Generatefinalmessage(RI.MessageTemp, RI.RequestContext)
	if err != nil {
		return err
	}
	RI.Message = message
	return nil
}

func (RI *RequestImpl) Wholemessageprocess() error {
	//apply extention
	proceesedmsg, err := ApplyExtentions(RI.Message, &RI.WholemessageprocessList)
	if err != nil {
		return err
	}
	RI.ProcessedMessage = proceesedmsg
	return err
}
