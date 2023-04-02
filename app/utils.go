package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func WriteFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

func MkDir(path string) error { // TODO remove
	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		return err
	}
	return nil
}

// checkFileExists checks if a file exists and returns an error if it does not.
func checkFileExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("error checking file %s: %v", path, err)
	}
	return nil
}
