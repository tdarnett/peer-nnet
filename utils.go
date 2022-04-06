package main

import (
	"bufio"
	"fmt"
	"os"
)

func WriteFile(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	numBytes, err := w.Write(data)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %d bytes\n", numBytes)
	return nil
}

func MkDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		return err
	}
	return nil
}
