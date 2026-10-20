package stressworkers

import (
	"log"
	"os"
	"sync"

	//import stresstester/jsonreader
	"stresstester/connectors"
	"stresstester/flowcontextloader"
)

func (sw Stressworker) Name() string {
	return "Stressworker"

}

func (sw Stressworker) Run(workers *int, messages *int, frequency *int, duration *int, per *string, messageTemp string, connectionParam *flowcontextloader.Contextmapper, connector connectors.Connectorinterface) {

	if *workers > 10 && *workers < 1 {
		log.Fatalf("Critical error: number of simultanous workers max 10 min 1")
		os.Exit(1)
	}

	if *messages <= 0 {
		log.Fatalf("Critical error: mpw number of messages per worker must be >= 1")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	resultsCh := make(chan string, *workers) // Buffered channel
	errs := make(chan WorkerErr, *workers)

	// Start workers
	for w := 1; w <= *workers; w++ {
		wg.Add(1)
		go simpleWorkerSender(w, *messages, messageTemp, connectionParam, connector, resultsCh, errs, &wg)
	}

	// Wait for all workers to finish in a background goroutine
	// so we don't block the main thread from reading results
	go func() {
		wg.Wait()
		close(resultsCh) // Always close channels when done sending
		close(errs)
	}()

	// Collect results and errors from the main thread
	// We read until both channels are completely drained
	for resultsCh != nil || errs != nil {
		select {
		case res, ok := <-resultsCh:
			if !ok {
				resultsCh = nil // channel is closed
				continue
			}
			log.Println(res)
		case err, ok := <-errs:
			if !ok {
				errs = nil // channel is closed
				continue
			}
			log.Printf("Worker %d Error caught: %s, Messages Sent: %d\n", err.WorkerNumber, err.Err, err.Count)
		}
	}
}

func simpleWorkerSender(workerNumber int, messageumber int, messageTemp string, connectionParam *flowcontextloader.Contextmapper, connector connectors.Connectorinterface /*jobs <-chan int,*/, results chan<- string, errs chan<- WorkerErr, wg *sync.WaitGroup) {
	defer wg.Done()
	rtnMessage, err, outworkerNumber, count := SendNumberOfMessages(workerNumber, messageumber, messageTemp, connectionParam, connector)
	if err != nil {
		workerErr := WorkerErr{WorkerNumber: outworkerNumber, Count: count, Err: err}
		errs <- workerErr
	} else {
		results <- rtnMessage
	}

}
