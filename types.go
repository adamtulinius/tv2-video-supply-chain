package main

import "time"

// Struct representing the metadata as described in the assignment
type IngestObject struct {
	Id                 string `json:"id"`
	Title              string `json:"title"`
	Provider           string `json:"provider"`
	EncodingTime       int    `json:"encodingTime"`
	PublicationTimeout int    `json:"publicationTimeout"`
}

// IngestObject wrapped to include a field for tracking when encoding started,
// so that publishingTimeout can be honered correctly
type PublishingObject struct {
	IngestObject    IngestObject
	EncodingStarted time.Time
}
