package main

import (
	"fmt"
	"path/filepath"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/bbolt"
	"github.com/philippgille/gokv/encoding"
)

const DB_DIR = "peers/databases/"
const PEER_MODELS_DIR = "peers/models/"

func InitStore(hostID string) (gokv.Store, error) {
	// create db and weights directory
	dbDir := filepath.Join(".", DB_DIR)
	peerModelsDir := filepath.Join(".", PEER_MODELS_DIR)

	err := MkDir(dbDir)
	if err != nil {
		return nil, err
	}
	err = MkDir(peerModelsDir)
	if err != nil {
		return nil, err
	}

	dbPath := fmt.Sprintf("%s/%s.db", dbDir, hostID)
	dbOptions := bbolt.Options{
		BucketName: "default",
		Path:       dbPath,
		Codec:      encoding.JSON,
	}

	db, err := bbolt.NewStore(dbOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db: %s", err)
	}

	// db has been previously initialized if host's ID is present
	hostNeuralNet := new(NeuralNet)
	found, err := db.Get(hostID, hostNeuralNet)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup host ID %s: %s", hostID, err)
	}
	if !found {
		fmt.Printf("No previous store found for host %s. Initializing...\n", hostID)
		err := db.Set(hostID, NeuralNet{Version: 0, SampleSize: 0})
		if err != nil {
			return nil, fmt.Errorf("failed to update version for host %s: %s", hostID, err)
		}

	} else {
		fmt.Printf("hostnnet: %+v\n", *hostNeuralNet)
	}

	return db, nil
}
