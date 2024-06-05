package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

const dsnEnvVar = "DB_DSN"

func main() {
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
