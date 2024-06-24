package main

import "sync"

// Start the server half of the system, including promethings endpoints and encoders
func start_server(wg *sync.WaitGroup, encoding_queue chan IngestObject, listening_port int, encoder_count int) {
	defer wg.Done()

	// Start prometheus endpoint
	go prometheus_endpoint(listening_port)
	go update_prometheus_metrics(encoding_queue)

	// Channel for indicating encoding successes, so they can be shown as stats in the console
	encoding_done_chan := make(chan int)

	// Buffered channel encoders use to hand over work to publishers on
	publishing_queue := make(chan PublishingObject, 100)

	// Goroutine that does the actual counting of encoding successes
	wg.Add(1)
	go encoder_stats(wg, encoding_done_chan)

	// Use a dedicated WaitGroup for encoders, so that the start_server goroutine can wait
	// for all encoders to finish, and close the encoding_done_chan channel.
	var wg_encoders sync.WaitGroup

	// Start encoders
	for encoder_id := range encoder_count {
		wg_encoders.Add(1)
		go encoder(&wg_encoders, encoder_id, encoding_queue, encoding_done_chan, publishing_queue)
	}

	// Start a new publisher manager
	wg.Add(1)
	go publisher_manager(wg, publishing_queue)

	wg_encoders.Wait()
	// This program will not terminate unless this channel is closed o/
	close(encoding_done_chan)

	// Since all encoders have stopped at this point, it's now safe to close the publishing_queue.
	// Publishers might still be running, but no new publishing job will be created.
	close(publishing_queue)
}
