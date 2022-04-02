package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func WriteFile(path string, data []byte) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	numBytes, err := w.Write(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("wrote %d bytes\n", numBytes)
}

func MkDir(path string) {
	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}
}
