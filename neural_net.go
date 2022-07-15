package main

import "time"

type NeuralNet struct {
	Version    int `json:"version"`
	SampleSize int `json:"sample_size"`
}

type NeuralNetMetadata struct {
	Version          int       `json:"version"`
	SampleSize       int       `json:"sample_size"`
	UpdatedTimestamp time.Time `json:"updated_timestamp"`
}
