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
	wg.Add(1)

	publishing_object := PublishingObject{
		ingest_object,
		time.Now(),
	}

	publisher(&wg, publishing_object)

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
	wg.Add(1)

	ingest_object.PublicationTimeout = 1

	publishing_object := PublishingObject{
		ingest_object,
		time.Now(),
	}

	publisher(&wg, publishing_object)

	wg.Wait()

	// test validation

	if len(published_objects) > 0 {
		t.Errorf("number of published objects = %d; expected 0", len(published_objects))
	}

	if len(published_objects_failed) != 1 {
		t.Errorf("number of failed published objects = %d; expected 1", len(published_objects))
	}
}
