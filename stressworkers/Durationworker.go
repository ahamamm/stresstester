package stressworkers

import (
	"log"
	"os"
	"sync"
	"time"

	//import stresstester/jsonreader
	"stresstester/connectors"
	"stresstester/flowcontextloader"
)

func (sw Durationworker) Name() string {
	return "Durationworker"

}

func (sw Durationworker) Run(workers *int, messages *int, frequency *int, duration *int, per *string, messageTemp string, connectionParam *flowcontextloader.Contextmapper, connector connectors.Connectorinterface) {

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

	var wg sync.WaitGroup
	resultsCh := make(chan WorkerReturnType, *workers)

	// Start workers
	for w := 1; w <= *workers; w++ {
		wg.Add(1)
		go durationWorkerSender(w, *messages, *frequency, *duration, messageTemp, connectionParam, connector, resultsCh, &wg)
	}

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

func durationWorkerSender(workerNumber int, messageumber int, frequency int, duration int, messageTemp string, connectionParam *flowcontextloader.Contextmapper, connector connectors.Connectorinterface, results chan<- WorkerReturnType, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(frequency) * time.Second)
	stopTime := time.After(time.Duration(duration) * time.Second)

	defer ticker.Stop()

	var count int
	var errorCount int
	var round int

	count = 0
	errorCount = 0
	round = 0

	wrt := WorkerReturnType{WorkerNumber: workerNumber, Count: 0, Round: 0, ErrorCount: 0, Errs: nil}
	err := connector.Connect()
	if err != nil {
		wrt.Errs = append(wrt.Errs, err)
		results <- wrt
		return
	}
	defer connector.Close()

	round++
	log.Printf("Worker #%d Sending Round #%d.................................\n", workerNumber, round)
	errs, count, errorCount := SendNumberOfMessageswithoutConnect(messageumber, messageTemp, connectionParam, connector)
	wrt.ErrorCount = errorCount + wrt.ErrorCount
	wrt.Count = count + wrt.Count
	wrt.Errs = append(wrt.Errs, errs...)
	wrt.Round = round

	for {
		select {
		case <-ticker.C:
			round++
			log.Printf("Worker #%d Sending Round #%d.................................\n", workerNumber, round)
			errs, count, errorCount := SendNumberOfMessageswithoutConnect(messageumber, messageTemp, connectionParam, connector)
			wrt.ErrorCount = errorCount + wrt.ErrorCount
			wrt.Count = count + wrt.Count
			wrt.Errs = append(wrt.Errs, errs...)
			wrt.Round = round

		case <-stopTime:

			results <- wrt
			return
		}
	}

}
