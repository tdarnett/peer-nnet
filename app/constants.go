package main

import (
	"os"
	"path"
)

func GetEnvOrDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

const WEIGHTS_FILENAME  = "weights.h5"    // must be in sync with model manager weights filename constant!
const METADATA_FILENAME = "metadata.json" // must be in sync with model manager metadata filename constant!

const (
	hostModelDir = "model"
	peerModelDir = "peers"
	defaultDbDir = "db"
)

var HOST_MODEL_WEIGHTS_PATH = GetEnvOrDefault("HOST_MODEL_WEIGHTS_PATH", path.Join(hostModelDir, WEIGHTS_FILENAME))
var HOST_MODEL_METADATA_PATH = GetEnvOrDefault("HOST_MODEL_METADATA_PATH", path.Join(hostModelDir, METADATA_FILENAME))
var PEERS_MODELS_DIR = GetEnvOrDefault("PEERS_MODELS_DIR", peerModelDir)
var DB_DIR = GetEnvOrDefault("DB_DIR", defaultDbDir)
