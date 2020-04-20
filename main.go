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
	var rootFolder = flag.String("directory", "", "root folder, ex: /tmp")
	flag.NewFlagSet("file", flag.ExitOnError)
	flag.NewFlagSet("directory", flag.ExitOnError)
	flag.Parse()

	if *rootFolder == "" {
		*rootFolder = "."
	}

	if *fileName == "" {
		fmt.Println("flag -file required.")
		os.Exit(2)
	}

	extractor, err := NewExtractor(*rootFolder, *fileName)
	if err != nil {
		log.Fatalf("error on init extractor: %v", err)
	}

	nodes := extractor.Extract()
	content, err := json.MarshalIndent(nodes, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	_, err = fmt.Fprint(os.Stdout, string(content))
	//err = json.NewEncoder(os.Stdout).Encode(nodes)
	if err != nil {
		log.Fatal(err)
	}
}
