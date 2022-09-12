package main

type NeuralNet struct {
	Version     int   `json:"version"`
	SampleSize  int   `json:"sample_size"`
	LastUpdated int64 `json:"last_updated"`
}
