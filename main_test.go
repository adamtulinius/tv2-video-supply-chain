package main

import (
	"strconv"
	"testing"
)

// All these tests test that options set in environment variables are parsed correctly

// Test that keep running defaults to false
func TestGetKeepRunningDefault(t *testing.T) {
	keep_running := get_keep_running()

	if keep_running {
		t.Errorf("keep running = %t; expected %t", keep_running, false)
	}
}

// Test that keep running can be enabled (1/2)
func TestGetKeepRunningYes(t *testing.T) {
	t.Setenv("KEEP_RUNNING", "yEs")

	keep_running := get_keep_running()

	if !keep_running {
		t.Errorf("keep running = %t; expected %t", keep_running, true)
	}
}

// Test that keep running can be enabled (2/2)
func TestGetKeepRunningTrue(t *testing.T) {
	t.Setenv("KEEP_RUNNING", "TrUe")

	keep_running := get_keep_running()

	if !keep_running {
		t.Errorf("keep running = %t; expected %t", keep_running, true)
	}
}

// Test default listening port
func TestGetListeningPortDefault(t *testing.T) {
	port := get_listening_port()

	if port != default_listening_port {
		t.Errorf("prometheus listening port = %d; expected %d", port, default_listening_port)
	}
}

// Test that listening port can be changed
func TestGetListeningPortNonDefault(t *testing.T) {
	configured_port := 8000

	t.Setenv("PORT", strconv.Itoa(configured_port))
	port := get_listening_port()

	if port != configured_port {
		t.Errorf("prometheus listening port = %d; expected %d", port, configured_port)
	}
}

// Test default number of encoders
func TestGetNumberOfEncodersDefault(t *testing.T) {
	encoders := get_number_of_encoders()

	if encoders != default_number_of_encoders {
		t.Errorf("number of encoders = %d; expected %d", encoders, default_number_of_encoders)
	}
}

// Test that number of encoders can be changed
func TestGetNumberOfEncodersNonDefault(t *testing.T) {
	configured_number_of_encoders := 5

	t.Setenv("ENCODERS", strconv.Itoa(configured_number_of_encoders))
	encoders := get_number_of_encoders()

	if encoders != configured_number_of_encoders {
		t.Errorf("number of encoders = %d; expected %d", encoders, configured_number_of_encoders)
	}
}
