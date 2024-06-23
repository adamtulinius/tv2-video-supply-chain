package main

import (
	"os"
	"sync"
	"testing"
)

// Test that the client will start and add encoding jobs to a queue
func TestClient(t *testing.T) {
	// setup
	var wg sync.WaitGroup

	encoding_queue := make(chan IngestObject, 100000)
	input_file, _ := os.Open("data/tv2-video-metadata-ingest-test.txt")

	wg.Add(1)
	go start_client(&wg, encoding_queue, input_file)

	wg.Wait()

	// test validation

	if len(encoding_queue) != 10 {
		t.Errorf("number of objects in encoding queue = %d; expected 10", len(encoding_queue))
	}
}
