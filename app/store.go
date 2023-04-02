package main

import (
	"fmt"
	"os"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/bbolt"
	"github.com/philippgille/gokv/encoding"
)

func InitStore(hostID string) (gokv.Store, error) {
	if err := os.MkdirAll(DB_DIR, os.ModePerm); err != nil {
		return nil, err
	}

	dbPath := fmt.Sprintf("%s/peers.db", DB_DIR)
	dbOptions := bbolt.Options{
		BucketName: "default",
		Path:       dbPath,
		Codec:      encoding.JSON,
	}

	db, err := bbolt.NewStore(dbOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db: %s", err)
	}

	return db, nil
}
