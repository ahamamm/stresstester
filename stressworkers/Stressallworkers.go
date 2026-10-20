package stressworkers

import (
	"log"
	"os"
	"sync"

	//import stresstester/jsonreader
	"stresstester/connectors"
	"stresstester/flowcontextloader"
	"stresstester/request"
)

func (sw Stressallworkers) Name() string {
	return "IBMMQSimpleAllworkers"
}

func (sw Stressallworkers) Run(workers *int, messages *int, frequency *int, duration *int, per *string, messageTemp string, connectionParam *flowcontextloader.Contextmapper, connector connectors.Connectorinterface) {

	if *workers > 10 && *workers < 1 {
		log.Fatalf("Critical error: number of simultanous workers max 10 min 1")
		os.Exit(1)
	}

	if *messages <= 0 {
		log.Fatalf("Critical error: mpw number of messages per worker must be >= 1")
		os.Exit(1)
	}

	var wg sync.WaitGroup

	inputCh := make(chan request.RequestImpl, *workers**messages) // inputchannel

	resultsCh := make(chan WorkerReturnType, *workers)

	// Start workers
	for w := 1; w <= *workers; w++ {
		wg.Add(1)
		go simpleAllWorkerSender(w, connector, inputCh, resultsCh, &wg)
	}

	for i := 0; i < *messages; i++ {
		req := request.RequestImpl{
			MessageTemp:             messageTemp,
			PrecallList:             connectionParam.Precalls,
			WholemessageprocessList: connectionParam.Wholemessageprocesses,
			RequestContext:          make(map[string]string),
			Message:                 "",
			ProcessedMessage:        "",
		}

		inputCh <- req
	}

	close(inputCh)

	// Wait for all workers to finish in a background goroutine
	// so we don't block the main thread from reading results
	go func() {
		wg.Wait()
		close(resultsCh) // Always close channels when done sending
	}()

	// Collect results and errors from the main thread
	// We read until both channels are completely drained
	for resultsCh != nil {
		res, ok := <-resultsCh
		if !ok {
			resultsCh = nil // channel is closed
			continue
		}
		log.Printf("Worker %d, Sent #messages: %d Successfully, Failed to send #messsages: %d\n ", res.WorkerNumber, res.Count, res.ErrorCount)
	}
}

func simpleAllWorkerSender(workerNumber int, connector connectors.Connectorinterface, inputCh <-chan request.RequestImpl, results chan<- WorkerReturnType, wg *sync.WaitGroup) {
	defer wg.Done()
	var count int
	var errorCount int
	count = 0
	errorCount = 0
	wrt := WorkerReturnType{WorkerNumber: workerNumber, Count: 0, Round: 0, ErrorCount: 0, Errs: nil}
	err := connector.Connect()
	if err != nil {
		wrt.Errs = append(wrt.Errs, err)
		log.Printf("Pre-Call error: %v\n", err)
		results <- wrt
		return
	}
	defer connector.Close()
	for msg := range inputCh {

		err := msg.Precalls()
		if err != nil {
			errorCount++
			wrt.Errs = append(wrt.Errs, err)
			log.Printf("Pre-Call error: %v\n", err)
			continue
		}

		err = msg.Messagegeneration()
		if err != nil {
			errorCount++
			wrt.Errs = append(wrt.Errs, err)
			log.Printf("Message generation error: %v\n", err)
			continue
		}

		var finalmessage string
		err = msg.Wholemessageprocess()
		if err != nil {
			errorCount++
			finalmessage = msg.Message
			wrt.Errs = append(wrt.Errs, err)
			log.Printf("Critical error: %v\n", err)

		} else {
			finalmessage = msg.ProcessedMessage
		}

		err = connector.Sendmessage(finalmessage, nil)
		if err != nil {
			errorCount++
			wrt.Errs = append(wrt.Errs, err)
		} else {
			count++
		}
	}
	//rtnMessage := fmt.Sprintf("Worker %d finished job and %d %s", workerNumber, count, " Message(s) sent ��� successfully!")
	wrt.ErrorCount = errorCount
	wrt.Count = count
	results <- wrt
}
