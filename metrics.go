package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define prometheus metrics
var (
	encoded_total = 0

	metrics_name_prefix = "tv2_video_"

	metrics_encoding_queue_length = promauto.NewGauge(prometheus.GaugeOpts{
		Name: metrics_name_prefix + "encoding_queue",
		Help: "The total number of objects awaiting encoding.",
	})

	metrics_metadata_parse_errors = promauto.NewCounter(prometheus.CounterOpts{
		Name: metrics_name_prefix + "metadata_parse_errors_total",
		Help: "The total number of objects that couldn't be parsed.",
	})

	metrics_encoding_requests_accepted = promauto.NewCounter(prometheus.CounterOpts{
		Name: metrics_name_prefix + "encoding_requests_accepted_total",
		Help: "The total number of encoding requests that has been accepted.",
	})

	metrics_encoders_available = promauto.NewGauge(prometheus.GaugeOpts{
		Name: metrics_name_prefix + "encoders_available",
		Help: "The total number of available encoders.",
	})

	metrics_encoders_busy = promauto.NewGauge(prometheus.GaugeOpts{
		Name: metrics_name_prefix + "encoders_busy",
		Help: "The total number of busy encoders.",
	})

	metrics_publishers_running = promauto.NewGauge(prometheus.GaugeOpts{
		Name: metrics_name_prefix + "publishers_running",
		Help: "The total number of running publishers.",
	})

	metrics_publication_successes = promauto.NewCounter(prometheus.CounterOpts{
		Name: metrics_name_prefix + "publication_successes_total",
		Help: "The total number of successful publications.",
	})

	metrics_publication_failures = promauto.NewCounter(prometheus.CounterOpts{
		Name: metrics_name_prefix + "publication_failures_total",
		Help: "The total number of publications that didn't finish within the time limit.",
	})

	metrics_encoded = promauto.NewCounter(prometheus.CounterOpts{
		Name: metrics_name_prefix + "encoded_total",
		Help: "The total number of objects that has been encoded.",
	})
)

// start a web server to expose prometheus metrics
func prometheus_endpoint(listening_port int) {
	// expose metrics
	http.Handle("/metrics", promhttp.Handler())

	// bind to port and panic if in use
	err := http.ListenAndServe(fmt.Sprintf(":%d", listening_port), nil)
	if err != nil {
		panic(err)
	}
}

// Helper function to update the length of the encoding queue
//
// Since the encoding queue is shared between both client and server,
// it seemed cleaner to just have the encoding queue length metric
// updated async here.
func update_prometheus_metrics(encoding_queue chan IngestObject) {
	for {
		metrics_encoding_queue_length.Set(float64(len(encoding_queue)))
		time.Sleep(100 * time.Millisecond)
	}
}

// Helper function that prints the current processing progress
func log_stats() {
	fmt.Printf(
		"stats: client_parse_errors=%d client_submitted=%d encoded=%d published=%d publish_failed=%d\n",
		len(metadata_unparsable_objects),
		client_submitted,
		encoded_total,
		len(published_objects),
		len(published_objects_failed),
	)
}

// update the number of encoded objects
func encoder_stats(wg *sync.WaitGroup, encoding_done_chan chan int) {
	defer wg.Done()

	// increment counter everytime something is written to the channel
	for range encoding_done_chan {
		encoded_total += 1
	}
}
