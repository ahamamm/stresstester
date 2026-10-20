package stressworkers

import (
	//import stresstester/jsonreader

	"stresstester/connectors"
	"stresstester/flowcontextloader"
)

type WorkerErr struct {
	WorkerNumber int
	Count        int
	Err          error
}

type WorkerReturnType struct {
	WorkerNumber int
	Count        int
	Round        int
	ErrorCount   int
	Errs         []error
}

type NoWorkerSelected struct{}
type Stressworker struct{}
type Stressallworkers struct{}
type Durationworker struct{}
type Durationallworkers struct{}

type Workerinterface interface {
	Name() string
	Run(workers *int, messages *int, frequency *int, duration *int, per *string, messageTemp string, connectionParam *flowcontextloader.Contextmapper, connector connectors.Connectorinterface)
}
