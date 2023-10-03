package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
)

const (
	prefixObfuscate = "//obfuscate"
	prefixHash      = "//hash"
)

var (
	outFile = flag.String("output", "", "Output file name")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: obfuscator [flags] [path ...]\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) <= 0 {
		fmt.Fprintf(os.Stderr, "no files to parse provided.\n")
		usage()
	}

	// parse our source code files
	src, err := ParseFiles(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing files: "+err.Error())
	}

	// generate the new source code file
	var buf bytes.Buffer
	if err := src.Generate(&buf); err != nil {
		log.Fatal(err)
	}
	data, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	if *outFile == "" {
		_, err = os.Stdout.Write(data)
	} else {
		err = os.WriteFile(*outFile, data, 0644)
	}
	if err != nil {
		log.Fatal(err)
	}
}
