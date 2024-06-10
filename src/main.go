package main

import (
	"flag"
	"log/slog"
	"os"
)

const (
	dsnEnvVar      = "DB_DSN"
	defaultWorkers = 10
	maxConnections = 50
	// max_connections was increased from 25 to 100 after some debugging
	// I have notes to share about this
)

type programArgs struct {
	csvFilePath *string
	numWorkers  *int
}

func main() {
	// stack this first so we can still use defer for other calls
	var err error
	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()

	args := parseFlags()

	numConns := min(*args.numWorkers, maxConnections)
	if numConns < *args.numWorkers {
		slog.With("num_connections", numConns, "num_workers", *args.numWorkers).
			Warn("Number of connections is lower than number of workers due to max connections limit")
	}

	dbpool, err := NewPool(os.Getenv(dsnEnvVar), numConns)
	if dbpool != nil {
		defer dbpool.Close()
	}
	if err != nil {
		return
	}

	stats, err := NewStats()
	if err != nil {
		return
	}
	defer stats.Stop()

	workers := NewWorkerPool(dbpool, stats)
	workers.Start(*args.numWorkers)

	if *args.csvFilePath != "" {
		err = StreamParamsFilePath(*args.csvFilePath, workers.QueryCallback())
	} else {
		slog.Info("Reading CSV input from stdin")
		err = StreamParams(os.Stdin, workers.QueryCallback())
	}

	workers.Stop()
	stats.Print()
}

func parseFlags() programArgs {
	args := programArgs{
		csvFilePath: flag.String("csv", "", "path to CSV file with input query params"),
		numWorkers:  flag.Int("workers", defaultWorkers, "number of workers sending queries in parallel"),
	}
	flag.Parse()
	return args
}
