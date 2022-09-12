package main

import (
	"io/ioutil"
	"os"
)

func WriteFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

func MkDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		return err
	}
	return nil
}
