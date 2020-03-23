package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var fileName = flag.String("file", "", "file path, ex: main.go")
	flag.NewFlagSet("file", flag.ExitOnError)
	flag.Parse()

	if *fileName == "" {
		fmt.Println("flag -file required.")
		os.Exit(2)
	}

	extractor, err := NewExtractor(*fileName)
	if err != nil {
		log.Fatalf("error on init extractor: %v", err)
	}

	nodes := extractor.Extract()
	err = json.NewEncoder(os.Stdout).Encode(nodes)
	if err != nil {
		log.Fatal(err)
	}
}
