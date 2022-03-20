package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/bbolt"
	"github.com/philippgille/gokv/encoding"
)

const DB_DIR = "databases/"

func InitStore(hostID string) gokv.Store {
	// create db directory
	dbDir := filepath.Join(".", DB_DIR)
	err := os.MkdirAll(dbDir, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}

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
