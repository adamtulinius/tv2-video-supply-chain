package main

import (
	"sync"
	"testing"
)

// End-to-end test of the server, testing that encoding jobs will be published
func TestServer(t *testing.T) {
	t.Cleanup(test_reset_data)

	// setup
	var wg sync.WaitGroup

	encoding_queue := make(chan IngestObject, 1)
	encoding_queue <- ingest_object
	close(encoding_queue)

	wg.Add(1)
	go start_server(&wg, encoding_queue, 8080, 1)
	wg.Wait()

	// test validation

	if len(published_objects) != 1 {
		t.Errorf("number of published objects = %d; expected 1", len(published_objects))
	}

	if len(published_objects_failed) != 0 {
		t.Errorf("number of failed published objects = %d; expected 0", len(published_objects_failed))
	}
}
