package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	// Channels for keeping track of what has been published, and what failed publishing.
	// The channels are made "big enough" to make sure writing to them will not block for this PoC.
	published_objects        = make(chan IngestObject, 100000)
	published_objects_failed = make(chan IngestObject, 100000)
	prewarmed_publishers     = 10
)

// Manages publishers. Tries to keep perwarmed_publishers ready.
func publisher_manager(wg *sync.WaitGroup, publishing_queue chan PublishingObject) {
	defer wg.Done()

	// Use a dedicated WaitGroup for publishers
	var wg_publishers sync.WaitGroup

	new_publisher_queue := make(chan int, 10)

	// Make prewarmed_publishers publishers ready in the background
	for range prewarmed_publishers {
		wg_publishers.Add(1)
		go publisher(&wg_publishers, publishing_queue, new_publisher_queue)
	}

	go func() {
		// Publishers write to new_publisher_queue when they get an encoding object
		// to publish. This signifies that it's time to make a new publisher ready
		// in the background.
		for range new_publisher_queue {
			wg_publishers.Add(1)
			go publisher(&wg_publishers, publishing_queue, new_publisher_queue)
		}
	}()

	// Wait for all publishers to shut down
	wg_publishers.Wait()
	close(new_publisher_queue)
}

// A publisher takes an encoded object and publishes it to TV2 Play. Each publisher
// can only publish a single object, and the process takes 2-6 seconds.
func publisher(wg *sync.WaitGroup, publishing_queue chan PublishingObject, new_publisher_queue chan int) {
	metrics_publishers_running.Inc()
	defer metrics_publishers_running.Dec()
	defer wg.Done()

	name := uuid.NewString()

	fmt.Printf("publisher %s: Starting\n", name)

	// Starting a publisher takes 2-6 seconds before it's ready to publish
	startup_delay := 4*rand.Float64() + 2
	time.Sleep(time.Duration(startup_delay * float64(time.Second)))

	// Wait for object to be available to the publisher
	publishing_object, ok := <-publishing_queue
	if !ok {
		// channel was closed without receiving an object
		return
	}

	// announce that this publisher is now in use and that another should be started up
	new_publisher_queue <- 1

	ingest_object := publishing_object.IngestObject

	fmt.Printf("publisher %s: Publishing id=%s title=\"%s\": Started (took %.2fs)\n", name, ingest_object.Id, ingest_object.Title, startup_delay)

	// Publishing process goes here (intentinally left blank)

	// Calculate how long time it took from encoding started to publishing
	// is possible, and fail publishing if it took longer than publicationTimeout.
	time_elapsed := time.Since(publishing_object.EncodingStarted)

	if time_elapsed <= time.Duration(float64(ingest_object.PublicationTimeout)*float64(time.Second)) {
		// Publication success
		fmt.Printf("publisher %s: Publishing id=%s title=\"%s\": Finished after %v (timeout=%d)\n", name, ingest_object.Id, ingest_object.Title, time_elapsed, ingest_object.PublicationTimeout)
		metrics_publication_successes.Inc()
		published_objects <- ingest_object
		log_stats()
	} else {
		// Publication failure
		fmt.Printf("publisher %s: Publishing id=%s title=\"%s\": Failed after %v (timeout=%d)\n", name, ingest_object.Id, ingest_object.Title, time_elapsed, ingest_object.PublicationTimeout)
		metrics_publication_failures.Inc()
		published_objects_failed <- ingest_object
		log_stats()
	}

	// Publisher terminates
}
