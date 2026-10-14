package stressworkers

import (
	"fmt"
	"log"

	//import stresstester/jsonreader
	"stresstester/flowcontextloader"
	"stresstester/request"

	//import stresstester/mqutil
	mqutil "stresstester/mqutil"
)

func SendNumberOfMessages(workerNumber int, messageumber int, messageTemp string, connectionParam *flowcontextloader.Jsonmapper) (string, error, int, int) {
	var count int
	count = 0

	err := mqutil.Init(connectionParam.Connectionstring, connectionParam.Queuemanager, connectionParam.Queuename)
	if err != nil {
		return "", err, workerNumber, count
	}

	for i := 0; i < messageumber; i++ {

		req := request.RequestImpl{
			MessageTemp:             messageTemp,
			PrecallList:             connectionParam.Precalls,
			WholemessageprocessList: connectionParam.Wholemessageprocesses,
			RequestContext:          make(map[string]string),
			Message:                 "",
			ProcessedMessage:        "",
		}

		err = req.Precalls()
		if err != nil {
			return "", err, workerNumber, count
		}

		err = req.Messagegeneration()
		if err != nil {
			return "", err, workerNumber, count
		}

		var finalmessage string
		//apply extention
		err = req.Wholemessageprocess()
		if err != nil {
			finalmessage = req.Message
			log.Printf("Critical error: %v, message \n %s", err, req.Message)

		} else {
			finalmessage = req.ProcessedMessage
		}
		err = mqutil.Putmessage(finalmessage)
		if err != nil {
			return "", err, workerNumber, count
		}
		count++
	}
	mqutil.CloseConn()

	return fmt.Sprintf("Worker %d finished job and %d %s", workerNumber, count, " Message(s) sent 🎉 successfully!"), nil, workerNumber, count
}

func SendNumberOfMessageswithoutConnect(messageumber int, messageTemp string, connectionParam *flowcontextloader.Jsonmapper) ([]error, int, int) {
	var errs []error
	var count int
	var errorCount int
	count = 0
	errorCount = 0

	for i := 0; i < messageumber; i++ {

		req := request.RequestImpl{
			MessageTemp:             messageTemp,
			PrecallList:             connectionParam.Precalls,
			WholemessageprocessList: connectionParam.Wholemessageprocesses,
			RequestContext:          make(map[string]string),
			Message:                 "",
			ProcessedMessage:        "",
		}

		err := req.Precalls()
		if err != nil {
			errs = append(errs, err)
			errorCount++
			continue
		}

		req.Messagegeneration()

		err = req.Messagegeneration()
		if err != nil {
			errs = append(errs, err)
			errorCount++
			continue
		}

		//apply extention
		var finalmessage string
		err = req.Wholemessageprocess()
		if err != nil {
			errs = append(errs, err)
			finalmessage = req.Message

		} else {
			finalmessage = req.ProcessedMessage
		}

		err = mqutil.Putmessage(finalmessage)
		if err != nil {
			errs = append(errs, err)
			errorCount++
		} else {
			count++
		}
	}
	return errs, count, errorCount
}
