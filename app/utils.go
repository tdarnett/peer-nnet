package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func WriteFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

func checkFileExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("error checking file %s: %v", path, err)
	}
	return nil
}
