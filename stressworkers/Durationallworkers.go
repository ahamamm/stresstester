package stressworkers

import (
	"log"
	"os"
	"sync"
	"time"

	//import stresstester/jsonreader
	"stresstester/connectors"
	"stresstester/flowcontextloader"
	"stresstester/request"
)

type DurationAllWorkersParam struct {
	Round   int
	Request request.RequestImpl
}

func (sw Durationallworkers) Name() string {
	return "Durationallworkers"

}

func (sw Durationallworkers) Run(workers *int, messages *int, frequency *int, duration *int, per *string, messageTemp string, connectionParam *flowcontextloader.Contextmapper, connector connectors.Connectorinterface) {

	if *workers > 10 || *workers < 1 {
		log.Fatalf("Critical error: number of simultanous workers max 10 min 1")
		os.Exit(1)
	}

	if *messages <= 0 {
		log.Fatalf("Critical error: mpw number of messages per worker must be >= 1")
		os.Exit(1)
	}

	if *duration <= 0 {
		log.Fatalf("Critical error: duration must be >= 1")
		os.Exit(1)
	}

	if *frequency <= 0 {
		log.Fatalf("Critical error: frequency must be >= 1")
		os.Exit(1)
	}

	ticker := time.NewTicker(time.Duration(*frequency) * time.Second)
	stopTime := time.After(time.Duration(*duration) * time.Second)

	defer ticker.Stop()

	var round int

	round = 0

	var wg sync.WaitGroup

	resultsCh := make(chan WorkerReturnType, *workers)
	inputCh := make(chan DurationAllWorkersParam, (*workers)*(*messages))

	// Start workers
	for w := 1; w <= *workers; w++ {
		wg.Add(1)
		go durationAllWorkersSender(w, connector, inputCh, resultsCh, &wg)
	}

	sendBurst := func() {
		round++

		for i := 0; i < *messages; i++ {

			req := request.RequestImpl{
				MessageTemp:             messageTemp,
				PrecallList:             connectionParam.Precalls,
				WholemessageprocessList: connectionParam.Wholemessageprocesses,
				RequestContext:          make(map[string]string),
				Message:                 "",
				ProcessedMessage:        "",
			}

			inputCh <- DurationAllWorkersParam{round, req}

		}
	}

	sendBurst()

	keepRunning := true
	for keepRunning {
		select {
		case <-ticker.C:
			sendBurst()

		case <-stopTime:
			keepRunning = false
		}
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
		log.Printf("Worker %d, Ran for #rounds: %d, Sent #messages: %d Successfully, Failed to send #messsages: %d\n", res.WorkerNumber, res.Round, res.Count, res.ErrorCount)
	}

}

func durationAllWorkersSender(workerNumber int, connector connectors.Connectorinterface, inputCh <-chan DurationAllWorkersParam, results chan<- WorkerReturnType, wg *sync.WaitGroup) {
	defer wg.Done()

	var count int
	var errorCount int
	var round int

	round = 0
	count = 0
	errorCount = 0
	var errs []error

	rtn := func() {
		wrt := WorkerReturnType{WorkerNumber: workerNumber, Count: count, Round: round, ErrorCount: errorCount, Errs: errs}
		wrt.ErrorCount = errorCount
		wrt.Count = count
		wrt.Round = round
		wrt.Errs = append(wrt.Errs, errs...)
		results <- wrt
	}

	err := connector.Connect()
	if err != nil {
		errs = append(errs, err)
		rtn()
		return
	}
	defer connector.Close()

	for msg := range inputCh {

		if round != msg.Round {
			log.Printf("Worker #%d Sending Round #%d.................................\n", workerNumber, msg.Round)
		}
		round = msg.Round

		err := msg.Request.Precalls()
		if err != nil {
			log.Printf("Error sending PreCall message: %v\n", err)
			errorCount++
			errs = append(errs, err)
			continue
		}

		err = msg.Request.Messagegeneration()
		if err != nil {
			log.Printf("Error generating message: %v\n", err)
			errorCount++
			errs = append(errs, err)
			continue
		}

		var fmsg string
		//apply extention
		err = msg.Request.Wholemessageprocess()
		if err != nil {
			//inputCh <- DurationAllWorkersParam{round, req.Message, req}
			fmsg = msg.Request.Message
			log.Printf("Error Apply Extentions on message: %v", err)
			errorCount++
			errs = append(errs, err)
			continue

		} else {
			fmsg = msg.Request.ProcessedMessage
		}

		err = connector.Sendmessage(fmsg, nil)

		if err != nil {
			errorCount++
			errs = append(errs, err)
			log.Printf("Error sending message: %v", err)
		} else {
			count++
		}
	}

	rtn()
}
