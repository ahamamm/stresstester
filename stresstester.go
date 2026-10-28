package main

import (
	"flag"
	"log"
	"os"

	//import stresstester/jsonreader
	"stresstester/connectors"
	"stresstester/flowcontextloader"
	"stresstester/pickers"
	"stresstester/utils"

	//import stresstester/mqutil

	//Import "stresstester/messagehandler"
	messagehandler "stresstester/messagehandler"
	//Import "stresstester/utils"

	//import stresstester/workers
	stressworkers "stresstester/stressworkers"
)

func main() {
	// Arguments: (flag name, default value, description)
	tpath := flag.String("tpath", "jsontemplate/temp1.json", "a path to a JSON template")
	workers := flag.Int("workers", 1, "Number of simultanous workers max 10 min 1")
	per := flag.String("per", "one-worker", "messages per all-workers / one-worker")
	strategy := flag.String("strategy", "stress", "stress / duration/")
	frqnc := flag.Int("frqnc", 1, "send message(s) each number of seconds")
	//is not working with stress and mps stategeies
	duration := flag.Int("duration", 0, "test duration")
	messages := flag.Int("messages", 1, "Number of messages to be sent")

	// Parse the flags. This must be called before using the variables.
	flag.Parse()

	// Access the values using pointers (*variableName)
	log.Println("Template Path value:", *tpath)
	log.Println("Workers:", *workers)
	log.Println("messages will be sent by:", *per)
	log.Println("test strategy: -> (stress / duration)", *strategy)
	log.Println("send message(s) each number of seconds:", *frqnc)
	log.Println("duration:", *duration)
	log.Println("Number of Messages per each worker:", *messages)

	if *workers > 10 && *workers < 1 {
		log.Fatalf("Critical error: Number of simultanous workers max 10 min 1")
		os.Exit(1)
	}

	if *messages <= 0 {
		log.Fatalf("Critical error: Number of messages must be >= 1")
		os.Exit(1)
	}

	connectionParam, err := flowcontextloader.Mappjson(*tpath)
	if err != nil {
		log.Fatalf("Critical error: %v", err)
		os.Exit(1)
	}

	absPath, err := utils.GetRealPath(connectionParam.Messagefilepath)
	if err != nil {
		log.Printf("Error getting Message template abs path: %v, relative pathe used.\n", err)
	}

	messageTemp, err := messagehandler.Getmessagetemplate(absPath)
	if err != nil {
		log.Fatalf("Critical error: %v", err)
		os.Exit(1)
	}

	//fmt.Printf("%#v\n", connectionParam )

	var pickedworker stressworkers.Workerinterface
	err, pickedworker = pickers.Workerpicker(strategy, per, connectionParam.Endpointtype, frqnc, duration)
	if err != nil {
		log.Fatalf("Critical error: %v", err)
		os.Exit(1)
	}
	log.Println("Selected Worker is:", pickedworker.Name())

	var pickedconnector connectors.Connectorinterface
	err, pickedconnector = pickers.Connectorpicker(connectionParam)
	if err != nil {
		log.Fatalf("Critical error: %v", err)
		os.Exit(1)
	}
	log.Println("Selected Connector is:", pickedconnector.Name())

	pickedworker.Run(workers, messages, frqnc, duration, per, messageTemp, &connectionParam, pickedconnector)
}
