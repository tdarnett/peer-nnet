package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/bbolt"
	"github.com/philippgille/gokv/encoding"
)

const DB_DIR = "peers/databases/"
const WEIGHTS_DIR = "peers/weights/"

func InitStore(hostID string) gokv.Store {
	// create db and weights directory
	dbDir := filepath.Join(".", DB_DIR)
	weightsDir := filepath.Join(".", WEIGHTS_DIR)

	MkDir(dbDir)
	MkDir(weightsDir)

	dbPath := fmt.Sprintf("%s/%s.db", dbDir, hostID)
	dbOptions := bbolt.Options{
		BucketName: "default",
		Path:       dbPath,
		Codec:      encoding.JSON,
	}

	db, err := bbolt.NewStore(dbOptions)
	if err != nil {
		log.Fatal(err)
	}

	// db has been previously initialized if host's ID is present
	hostNeuralNet := new(NeuralNet)
	found, err := db.Get(hostID, hostNeuralNet)
	if err != nil {
		panic(err)
	}
	if !found {
		fmt.Printf("No previous store found for host %s. Initializing...", hostID)
		err := db.Set(hostID, NeuralNet{version: 0})
		if err != nil {
			panic(err)
		}

	} else {
		fmt.Printf("hostnnet: %+v", *hostNeuralNet)
	}

	return db
}
