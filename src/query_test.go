package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// from the first row in the CSV file
var testParams = QueryParams{"host_000008", time.Date(2017, 1, 1, 8, 59, 22, 0, time.UTC), time.Date(2017, 1, 1, 9, 59, 22, 0, time.UTC)}

func TestExecuteQuery(t *testing.T) {
	dbpool, err := NewPool(os.Getenv(dsnEnvVar), 1)
	require.NoError(t, err)
	defer dbpool.Close()
	err = executeQuery(dbpool, testParams)
	require.NoError(t, err)
}

func TestExecuteErrorConnFailed(t *testing.T) {
	dbpool, err := NewPool("postgres://user:pass@localhost:35688/tsdb?sslmode=disable", 1)
	require.Error(t, err)
	if dbpool != nil {
		dbpool.Close()
	}
}
