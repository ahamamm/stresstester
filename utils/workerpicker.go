package utils

import (
	"errors"
	"stresstester/stressworkers"
)

func Workerpicker(strategy *string, per *string, endpointtype string, frequency *int, duration *int) (error, stressworkers.Workerinterface) {
	var sw stressworkers.Workerinterface
	var err error

	if (strategy != nil && *strategy == "stress") && (per != nil && *per == "one-worker") {
		sw = stressworkers.Stressworker{}
	} else if (strategy != nil && *strategy == "stress") && (per != nil && *per == "all-workers") {
		sw = stressworkers.Stressallworkers{}
	} else if (strategy != nil && *strategy == "duration") && (per != nil && *per == "one-worker") {
		sw = stressworkers.Durationworker{}
	} else if (strategy != nil && *strategy == "duration") && (per != nil && *per == "all-workers") {
		sw = stressworkers.Durationallworkers{}
	} else {
		sw = stressworkers.NoWorkerSelected{}
		err = errors.New("Wrong endpoint type, no worker selected, check your parameters.")

	}
	return err, sw
}
