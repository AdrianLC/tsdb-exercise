package main

import (
	"flag"
	"log/slog"
	"os"
)

const (
	dsnEnvVar      = "DB_DSN"
	defaultWorkers = 10
	maxConnections = 25 // I checked this in the settings in the Timescale service
)

type programArgs struct {
	csvFilePath *string
	numWorkers  *int
}

func main() {
	args := parseFlags()

	numConns := min(*args.numWorkers, maxConnections)
	if numConns < *args.numWorkers {
		slog.With("num_connections", numConns, "num_workers", *args.numWorkers).
			Warn("Number of connections is lower than number of workers due to max connections limit")
	}

	querier, err := NewQuerier(os.Getenv(dsnEnvVar), numConns)
	if err != nil {
		os.Exit(1)
	}
	defer querier.Close()

	if *args.csvFilePath != "" {
		err = StreamParamsFilePath(*args.csvFilePath, querier.QueryCallback())
	} else {
		slog.Info("Reading CSV input from stdin")
		err = StreamParams(os.Stdin, querier.QueryCallback())
	}

	if err != nil {
		os.Exit(1)
	}
}

func parseFlags() programArgs {
	args := programArgs{
		csvFilePath: flag.String("csv", "", "path to CSV file with input query params"),
		numWorkers:  flag.Int("workers", defaultWorkers, "number of workers sending queries in parallel"),
	}
	flag.Parse()
	return args
}
