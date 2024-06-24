package main

import (
	"sync"
	"testing"
	"time"
)

var (
	ingest_object = IngestObject{
		"b385fcf5-bacf-4242-892c-3ed08799a775",
		"Det sidste ord: Peter Belli",
		"TV 2",
		6,
		7,
	}
)

func test_reset_data() {
	published_objects = make(chan IngestObject, 10)
	published_objects_failed = make(chan IngestObject, 10)
}

// Test encoding time, number of encoded objects, and that encoded objects are passed on to a publisher
func TestEncoder(t *testing.T) {
	t.Cleanup(test_reset_data)

	// setup
	var wg sync.WaitGroup
	wg.Add(1)

	encoding_done_chan := make(chan int, 10)
	encoding_queue := make(chan IngestObject, 1)
	encoding_queue <- ingest_object
	close(encoding_queue)

	time_start := time.Now()
	encoder(&wg, 0, encoding_queue, encoding_done_chan)

	encoding_time := time.Since(time_start)

	wg.Wait()

	// Validate correct number of encoded objects
	if len(encoding_done_chan) != 1 {
		t.Errorf("number of encoded objects = %d; expected 1", len(encoding_done_chan))
	}

	// Validate correct number of published objects
	if len(published_objects) != 1 {
		t.Errorf("number of published objects = %d; expected 1", len(published_objects))
	}

	// Validate correct number of failed publishes
	if len(published_objects_failed) != 0 {
		t.Errorf("number of failed published objects = %d; expected 0", len(published_objects_failed))
	}

	// Test that runtime of encoding is between ingest_object.EncodingTime and ingest_object.EncodingTime+0.1s
	duration_lower_limit := time.Duration(ingest_object.EncodingTime) * time.Second
	duration_upper_limit := time.Duration((float64(ingest_object.EncodingTime) + 0.1) * float64(time.Second))

	if encoding_time < duration_lower_limit {
		t.Errorf("encoding finished in %v; expected at least %ds", encoding_time, ingest_object.EncodingTime)
	}

	if encoding_time > duration_upper_limit {
		t.Errorf("encoding finished in %v; expected %ds", encoding_time, ingest_object.EncodingTime)
	}
}
