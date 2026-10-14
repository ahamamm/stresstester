package stressworkers

import (
	"log"
	"os"
	"sync"

	//import stresstester/jsonreader
	"stresstester/flowcontextloader"
	"stresstester/stressworkers"
)

func (sw IBMMQSimpleworker) Name() string {
	return "IBMMQSimpleworker"

}

func (sw IBMMQSimpleworker) Run(workers *int, messages *int, frequency *int, duration *int, per *string, messageTemp string, connectionParam *flowcontextloader.Jsonmapper) {

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
	errs := make(chan stressworkers.WorkerErr, *workers)

	// Start workers
	for w := 1; w <= *workers; w++ {
		wg.Add(1)
		go simpleWorkerSender(w, *messages, messageTemp, connectionParam, resultsCh, errs, &wg)
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

func simpleWorkerSender(workerNumber int, messageumber int, messageTemp string, connectionParam *flowcontextloader.Jsonmapper /*jobs <-chan int,*/, results chan<- string, errs chan<- stressworkers.WorkerErr, wg *sync.WaitGroup) {
	defer wg.Done()
	rtnMessage, err, outworkerNumber, count := stressworkers.SendNumberOfMessages(workerNumber, messageumber, messageTemp, connectionParam)
	if err != nil {
		workerErr := stressworkers.WorkerErr{WorkerNumber: outworkerNumber, Count: count, Err: err}
		errs <- workerErr
	} else {
		results <- rtnMessage
	}

}
