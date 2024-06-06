package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
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
	if *args.csvFilePath != "" {
		go StreamParamsFilePath(*args.csvFilePath, queryParams)
	} else {
		fmt.Println("Awaiting CSV input from stdin:")
		go StreamParams(os.Stdin, queryParams)
	}
	for range queryParams {
	}

	testDBConn()
}

func parseFlags() programArgs {
	args := programArgs{
		csvFilePath: flag.String("csv", "", "path to CSV file with input query params"),
		numWorkers:  flag.Int("workers", defaultWorkers, "number of workers sending queries in parallel"),
	}
	flag.Parse()
	return args
}

func testDBConn() {
	// temporary code to move to test connectivity
	dsn := os.Getenv(dsnEnvVar)
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	type Extension struct {
		Extname    string
		Extversion string
	}

	rows, err := conn.Query(ctx, "select extname, extversion from pg_extension")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}

	extensions, err := pgx.CollectRows(rows, pgx.RowToStructByName[Extension])
	if err != nil {
		fmt.Fprintf(os.Stderr, "CollectRows failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%+v\n", extensions)
}
