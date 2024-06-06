package main

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	csvFilePath       = "../files/query_params.csv"
	expectedNumParams = 200
	headerText        = "hostname,start_time,end_time"
)

func TestStreamParamsFilePath(t *testing.T) {
	require := require.New(t)
	paramsChan := make(chan QueryParams)
	go StreamParamsFilePath(csvFilePath, paramsChan)

	var numParams int
	for params := range paramsChan {
		numParams++

		require.NotEmpty(params.Host)
		require.NotEmpty(params.StartTime)
		require.NotEmpty(params.EndTime)
	}

	require.Equal(expectedNumParams, numParams)
}

func TestStreamParamsFilePathFileError(t *testing.T) {
	paramsChan := make(chan QueryParams)
	err := StreamParamsFilePath("nonexistent.csv", paramsChan)
	require.Error(t, err)
}

func TestStreamParamsUnexpectedContent(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		file := strings.NewReader("")
		checkZeroParams(t, file)
	})
	t.Run("not a csv", func(t *testing.T) {
		file := strings.NewReader("not a csv")
		checkZeroParams(t, file)
	})
	t.Run("header missing fields", func(t *testing.T) {
		file := strings.NewReader("hostname,start_time\n")
		checkZeroParams(t, file)
	})
	t.Run("row missing fields", func(t *testing.T) {
		file := strings.NewReader(headerText + "\n" + "host_000008,2017-01-01 08:59:22\n")
		checkZeroParams(t, file)
	})
	t.Run("invalid dates", func(t *testing.T) {
		file := strings.NewReader(headerText + "\n" + "host_000008,2017-01-01,2017-01-01\n")
		checkZeroParams(t, file)
	})
}

func checkZeroParams(t *testing.T, file io.Reader) {
	paramsChan := make(chan QueryParams)
	go StreamParams(file, paramsChan)
	_, ok := <-paramsChan
	require.False(t, ok)
}
