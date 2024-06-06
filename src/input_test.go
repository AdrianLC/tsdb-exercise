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
	var numParams int
	fn := func(params QueryParams) error {
		numParams++
		require.NotEmpty(params.Host)
		require.NotEmpty(params.StartTime)
		require.NotEmpty(params.EndTime)
		return nil
	}
	err := StreamParamsFilePath(csvFilePath, fn)
	require.NoError(err)
	require.Equal(expectedNumParams, numParams)
}

func TestStreamParamsFilePathFileError(t *testing.T) {
	fn := func(params QueryParams) error {
		require.FailNow(t, "unexpected call")
		return nil
	}
	err := StreamParamsFilePath("nonexistent.csv", fn)
	require.Error(t, err)
}

func TestStreamParamsUnexpectedContent(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		file := strings.NewReader("")
		checkZeroCalls(t, file)
	})
	t.Run("not a csv", func(t *testing.T) {
		file := strings.NewReader("not a csv")
		checkZeroCalls(t, file)
	})
	t.Run("header missing fields", func(t *testing.T) {
		file := strings.NewReader("hostname,start_time\n")
		checkZeroCalls(t, file)
	})
	t.Run("row missing fields", func(t *testing.T) {
		file := strings.NewReader(headerText + "\n" + "host_000008,2017-01-01 08:59:22\n")
		checkZeroCalls(t, file)
	})
	t.Run("invalid dates", func(t *testing.T) {
		file := strings.NewReader(headerText + "\n" + "host_000008,2017-01-01,2017-01-01\n")
		checkZeroCalls(t, file)
	})
}

func checkZeroCalls(t *testing.T, file io.Reader) {
	fn := func(params QueryParams) error {
		require.FailNow(t, "unexpected call")
		return nil
	}
	err := StreamParams(file, fn)
	require.NoError(t, err)
}
