package main

import (
	"sync"
	"testing"
	"time"
)

// Test that object is published
func TestPublisherWillPublish(t *testing.T) {
	t.Cleanup(test_reset_data)

	// setup

	var wg sync.WaitGroup
	new_publisher_queue := make(chan int, 10)
	publishing_queue := make(chan PublishingObject, 1)

	publishing_object := PublishingObject{
		ingest_object,
		time.Now(),
	}
	publishing_queue <- publishing_object

	wg.Add(1)
	publisher(&wg, publishing_queue, new_publisher_queue)

	wg.Wait()

	// test validation

	if len(published_objects) == 0 {
		t.Errorf("Nothing was published")
	} else {
		published_object := <-published_objects

		if len(published_objects) == 1 && published_object.Id != ingest_object.Id {
			t.Errorf("published object id = %s; expected %s", published_object.Id, ingest_object.Id)
		}

		if len(published_objects) > 1 {
			t.Errorf("number of published objects = %d; expected 1", len(published_objects))
		}
	}

	if len(published_objects_failed) > 0 {
		t.Errorf("number of failed published objects = %d; expected 0", len(published_objects))
	}
}

// Test that object fails publication
func TestPublisherWillFail(t *testing.T) {
	t.Cleanup(test_reset_data)

	// setup

	var wg sync.WaitGroup
	new_publisher_queue := make(chan int, 10)
	publishing_queue := make(chan PublishingObject, 1)

	publishing_object := PublishingObject{
		ingest_object_1s_timeout,
		time.Now(),
	}
	publishing_queue <- publishing_object

	wg.Add(1)
	publisher(&wg, publishing_queue, new_publisher_queue)

	wg.Wait()

	// test validation

	if len(published_objects) > 0 {
		t.Errorf("number of published objects = %d; expected 0", len(published_objects))
	}

	if len(published_objects_failed) != 1 {
		t.Errorf("number of failed published objects = %d; expected 1", len(published_objects))
	}
}
