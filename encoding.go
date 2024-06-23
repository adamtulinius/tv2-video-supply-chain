package main

import (
	"fmt"
	"sync"
	"time"
)

// An encoder reads encoding jobs from encoding_queue. A publisher is started, and the
// encoded data is sent to the publisher over a channel.
func encoder(wg *sync.WaitGroup, encoder_id int, encoding_queue chan IngestObject, encoding_done_chan chan int) {
	defer wg.Done()

	fmt.Printf("encoder %d: Started\n", encoder_id)
	metrics_encoders_available.Inc()
	defer metrics_encoders_available.Dec()

	// Continiously read encoding jobs from the queue
	for ingest_object := range encoding_queue {
		metrics_encoders_busy.Inc()
		encoding_started := time.Now()

		// Create a communication channel to the publisher
		publishing_queue := make(chan PublishingObject, 1)

		fmt.Printf("encoder %d: Encoding: id=%s title=\"%s\"\n", encoder_id, ingest_object.Id, ingest_object.Title)

		// Fake a heavy encoding job
		time.Sleep(time.Duration(ingest_object.EncodingTime) * time.Second)
		metrics_encoded.Inc()
		encoding_done_chan <- 1
		log_stats()

		// Create the publisher
		fmt.Printf("encoder %d: Starting publisher: id=%s title=\"%s\"\n", encoder_id, ingest_object.Id, ingest_object.Title)
		wg.Add(1)
		// Passing some extra parameters just to get some consistency in log output
		// to keep track of what object the publisher belongs to (not strictly
		// necessary).
		go publisher(wg, ingest_object.Id, ingest_object.Title, publishing_queue)

		// Wrap the IngestObject in a PublishingObject so that the publisher knows
		// when encoding started
		publishing_object := PublishingObject{
			ingest_object,
			encoding_started,
		}

		// Send the object to the queue belonging to the publisher
		publishing_queue <- publishing_object
		close(publishing_queue)
		log_stats()

		metrics_encoders_busy.Dec()
	}
}
