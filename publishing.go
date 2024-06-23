package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

var (
	// Channels for keeping track of what has been published, and what failed publishing.
	// The channels are made "big enough" to make sure writing to them will not block for this PoC.
	published_objects        = make(chan IngestObject, 100000)
	published_objects_failed = make(chan IngestObject, 100000)
)

// A publisher takes an encoded object and publishes it to TV2 Play. Each publisher
// can only publish a single object, and the process takes 2-6 seconds.
func publisher(wg *sync.WaitGroup, id string, title string, publishing_queue chan PublishingObject) {
	metrics_publishers_running.Inc()
	defer metrics_publishers_running.Dec()
	defer wg.Done()

	fmt.Printf("publisher: Publishing id=%s title=\"%s\": Starting\n", id, title)

	// Starting a publisher takes 2-6 seconds before it's ready to publish
	startup_delay := 4*rand.Float64() + 2
	time.Sleep(time.Duration(startup_delay * float64(time.Second)))

	// Wait for the encoder to send the encoded date to the publisher
	publishing_object := <-publishing_queue
	ingest_object := publishing_object.IngestObject

	fmt.Printf("publisher: Publishing id=%s title=\"%s\": Started (took %.2fs)\n", ingest_object.Id, ingest_object.Title, startup_delay)

	// Publishing process goes here (intentinally left blank)

	// Calculate how long time it took from encoding started to publishing
	// is possible, and fail publishing if it took longer than publicationTimeout.
	time_elapsed := time.Since(publishing_object.EncodingStarted)

	if time_elapsed <= time.Duration(float64(ingest_object.PublicationTimeout)*float64(time.Second)) {
		// Publication success
		fmt.Printf("publisher: Publishing id=%s title=\"%s\": Finished after %v (timeout=%d)\n", ingest_object.Id, ingest_object.Title, time_elapsed, ingest_object.PublicationTimeout)
		metrics_publication_successes.Inc()
		published_objects <- ingest_object
		log_stats()
	} else {
		// Publication failure
		fmt.Printf("publisher: Publishing id=%s title=\"%s\": Failed after %v (timeout=%d)\n", ingest_object.Id, ingest_object.Title, time_elapsed, ingest_object.PublicationTimeout)
		metrics_publication_failures.Inc()
		published_objects_failed <- ingest_object
		log_stats()
	}

	// Publisher terminates
}
