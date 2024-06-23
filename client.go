package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	client_submitted            = 0
	metadata_unparsable_objects = make(chan string, 100000)
)

// The client parses encoding metadata and submits it object by object to the server
// through the `encoding_queue` channel. Metadata is read from `input_file`, which could
// be e.g. STDIN.
// The client consists on two main parts:
// 1) A goroutine that parses the input_file and adds the parsed IngestObjects to a queue
// 2) A goroutine that reads from the queue and sends them to the server at a throttled rate
func start_client(wg *sync.WaitGroup, encoding_queue chan IngestObject, input_file *os.File) {
	defer wg.Done()

	// For this PoC the objects could also just have been added to a list, but it's easier
	// to do separation of concerns with channels.
	metadata_queue := make(chan IngestObject)

	// Start input_file parsing
	wg.Add(1)
	go parse_ingest_objects(wg, metadata_queue, input_file)

	// Start goroutine that submits jobs to the server
	wg.Add(1)
	go job_submitter(wg, metadata_queue, encoding_queue)
}

// Reads input_file line by line and parses each line as JSON. The result is added to the
// queue metadata_queue.
func parse_ingest_objects(wg *sync.WaitGroup, metadata_queue chan IngestObject, input_file *os.File) {
	defer wg.Done()

	// Buffered read of input_file
	scanner := bufio.NewScanner(input_file)

	// by default .Scan() reads a single line
	for scanner.Scan() {
		// Parse read line as JSON.
		var ingest_object IngestObject
		if err := json.Unmarshal(scanner.Bytes(), &ingest_object); err != nil {
			// Parsing failed, log it
			metrics_metadata_parse_errors.Inc()
			fmt.Fprintf(os.Stderr, "client: couldn't parse line as JSON: %v\n", scanner.Text())
			metadata_unparsable_objects <- scanner.Text()
		} else {
			// Add parsed object to the next queue in line
			metadata_queue <- ingest_object
			fmt.Printf("client: Parsed %s (%s)\n", ingest_object.Id, ingest_object.Title)
			log_stats()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "client: reading standard input:", err)
	}

	close(metadata_queue)
}

// This goroutine simulates another system sending encoding requests at a rate of approx 2/s.
// Encoding jobs/requests are written to encoding_queue.
func job_submitter(wg *sync.WaitGroup, metadata_queue chan IngestObject, encoding_queue chan IngestObject) {
	defer wg.Done()

	// Assumption: Since it is specified that encoding jobs are submitted at a rate of _up to_ 2 req/s.
	// Block on channel writes until the server has capacity to deal with the request.

	for ingest_object := range metadata_queue {
		fmt.Printf("client: New encoding request: id=%s title=\"%s\"\n", ingest_object.Id, ingest_object.Title)
		log_stats()

		encoding_queue <- ingest_object
		fmt.Printf("client: New encoding request: id=%s title=\"%s\": Queued\n", ingest_object.Id, ingest_object.Title)

		metrics_encoding_requests_accepted.Inc()
		client_submitted += 1
		log_stats()

		// Wait half a second to get up to ~2 req/s
		time.Sleep(500 * time.Millisecond)
	}

	close(encoding_queue)
}
