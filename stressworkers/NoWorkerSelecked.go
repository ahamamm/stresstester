package stressworkers

import (
	"log"
	"os"
	"stresstester/connectors"
	"stresstester/flowcontextloader"
)

func (NWS NoWorkerSelected) Run(workers *int, messages *int, frequency *int, duration *int, per *string, messageTemp string, connectionParam *flowcontextloader.Contextmapper, connector connectors.Connectorinterface) {
	log.Fatalf("Critical error: No worker selected, check your parameters.")
	os.Exit(1)
}
func (NWS NoWorkerSelected) Name() string {
	return "NoWorkerSelected"
}
