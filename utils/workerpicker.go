package utils

import (
	"errors"
	"stresstester/stressworkers"
	ibmmq "stresstester/stressworkers/ibmmq"
)

func Workerpicker(strategy *string, per *string, endpointtype string, frequency *int, duration *int) (error, stressworkers.Workerinterface) {
	var sw stressworkers.Workerinterface
	var err error
	switch endpointtype {
	case "IBMQM":
		if (strategy != nil && *strategy == "stress") && (per != nil && *per == "one-worker") {
			sw = ibmmq.IBMMQSimpleworker{}
		} else if (strategy != nil && *strategy == "stress") && (per != nil && *per == "all-workers") {
		} else if (strategy != nil && *strategy == "duration") && (per != nil && *per == "one-worker") {
			sw = ibmmq.IBMMQDurationWorker{}
		} else if (strategy != nil && *strategy == "duration") && (per != nil && *per == "all-workers") {
			sw = ibmmq.IBMMQDurationAllWorkers{}
		} else {
			sw = ibmmq.IBMMQNoWorkerSelected{}
			err = errors.New("Wrong Startegy and Per compination, No worker selected, check your parameters.")
		}
	default:
		sw = stressworkers.NoWorkerSelected{}
		err = errors.New("Wrong endpoint type, no worker selected, check your parameters.")
	}
	return err, sw
}
