package main

import "os"

func GetOrDefaultEnv(key string, fallback string) string {
	v := os.Getenv(key)
	if len(v) == 0 {
		return fallback
	}
	return v
}

const WEIGHTS_FILENAME = "weights.h5"     // must be in sync with model manager weights filename constant!
const METADATA_FILENAME = "metadata.json" // must be in sync with model manager metadata filename constant!

var HOST_MODEL_WEIGHTS_PATH = GetOrDefaultEnv("HOST_MODEL_WEIGHTS_PATH", "shared/model/"+WEIGHTS_FILENAME)
var HOST_MODEL_METADATA_PATH = GetOrDefaultEnv("HOST_MODEL_METADATA_PATH", "shared/model/"+METADATA_FILENAME)

var PEERS_MODELS_DIR = GetOrDefaultEnv("PEERS_MODELS_DIR", "shared/peers/")

var DB_DIR = GetOrDefaultEnv("DB_DIR", "database/")
