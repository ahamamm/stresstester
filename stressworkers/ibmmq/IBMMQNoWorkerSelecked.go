package stressworkers

import (
	"log"
	"os"
	"stresstester/flowcontextloader"
)

func (NWS IBMMQNoWorkerSelected) Run(workers *int, messages *int, frequency *int, duration *int, per *string, messageTemp string, connectionParam *flowcontextloader.Jsonmapper) {
	log.Fatalf("Critical error: No worker selected, check your parameters.")
	os.Exit(1)
}
func (NWS IBMMQNoWorkerSelected) Name() string {
	return "IBMMQNoWorkerSelected"
}
