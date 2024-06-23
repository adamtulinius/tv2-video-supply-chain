package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	metadata_input_file        = os.Stdin
	default_number_of_encoders = 10
	default_listening_port     = 8080
)

// main application entry point
func main() {
	// A single WaitGroup is used to track running goroutines, so that the
	// application doesn't shut down before all work has finished. A pointer
	// to this WaitGroup is passed along to all client and server components.
	var wg sync.WaitGroup

	// Let the application run forever if environment var `KEEP_RUNNING=yes`
	if get_keep_running() {
		fmt.Println("Keep running: true")
		// Just add to the wait group without ever decrementing it. It's a bit of a hack,
		// but I think that's ok since this is a bonus feature anyways.
		// A more proper solution would be not to close the channels raw_metadata_queue and encoding_queue.
		wg.Add(1)
	}

	// Encoding jobs are sent from the client to the server using the encoding_queue channel
	encoding_queue := make(chan IngestObject)

	// start the client and server
	wg.Add(1)
	go start_client(&wg, encoding_queue, metadata_input_file)
	wg.Add(1)
	go start_server(&wg, encoding_queue, get_listening_port(), get_number_of_encoders())

	// Wait for all goroutines to terminate before terminating the application
	wg.Wait()
	log_stats()
}

// Check if the environment variable `KEEP_RUNNING` is set
func get_keep_running() bool {
	option_keep_running := strings.ToLower(os.Getenv("KEEP_RUNNING"))
	return (option_keep_running == "yes" || option_keep_running == "true")
}

// Get number of encoders from the environment variable `ENCODERS`,
// or return default value
func get_number_of_encoders() int {
	if value, ok := os.LookupEnv("ENCODERS"); ok {
		number_of_encoders, err := strconv.Atoi(value)
		if err != nil {
			panic("Could not parse ENCODERS as a number")
		}

		return number_of_encoders
	} else {
		return default_number_of_encoders
	}
}

// Get listening port for prometheus metrics from the environment variable `PORT`,
// or return default value
func get_listening_port() int {
	if value, ok := os.LookupEnv("PORT"); ok {
		port, err := strconv.Atoi(value)
		if err != nil {
			panic("Could not parse PORT as a number")
		}

		return port
	} else {
		return default_listening_port
	}
}
