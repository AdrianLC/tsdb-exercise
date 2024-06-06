package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	dsnEnvVar      = "DB_DSN"
	defaultWorkers = 10
)

type programArgs struct {
	csvFilePath *string
	numWorkers  *int
}

func main() {
	args := parseFlags()
	queryParams := make(chan QueryParams)

	querier, err := NewQuerier(os.Getenv(dsnEnvVar), PoolOptions{MinConns: 1, MaxConns: *args.numWorkers}, queryParams)
	if err != nil {
		os.Exit(1)
	}
	defer querier.Close()

	if *args.csvFilePath != "" {
		go StreamParamsFilePath(*args.csvFilePath, queryParams)
	} else {
		fmt.Println("Awaiting CSV input from stdin:")
		go StreamParams(os.Stdin, queryParams)
	}
	querier.Run()
}

func parseFlags() programArgs {
	args := programArgs{
		csvFilePath: flag.String("csv", "", "path to CSV file with input query params"),
		numWorkers:  flag.Int("workers", defaultWorkers, "number of workers sending queries in parallel"),
	}
	flag.Parse()
	return args
}
