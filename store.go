package main

import (
	"fmt"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/bbolt"
	"github.com/philippgille/gokv/encoding"
)

func InitStore(hostID string) (gokv.Store, error) {
	err := MkDir(DB_DIR)
	if err != nil {
		return nil, err
	}

	dbPath := fmt.Sprintf("%s/%s.db", DB_DIR, hostID)
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
