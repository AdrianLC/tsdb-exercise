package main

import (
	"encoding/csv"
	"io"
	"log/slog"
	"os"
	"time"
)

func StreamParamsFilePath(filePath string, fn QueryFunc) error {
	log := slog.With("file_path", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("could not open csv: %w", err)
		return err
	}
	defer file.Close()
	log.Info("reading csv")
	return StreamParams(file, fn)
}

func StreamParams(file io.Reader, fn QueryFunc) error {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 3
	reader.TrimLeadingSpace = true

	_, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		slog.Warn("unexpected error reading csv header: %w", err)
		// continue anyway, perhaps only the header is wrong
	}

	currentRow := 1
	log := slog.With("row_number", currentRow)

	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Warn("unexpected error reading csv row: %w", err)
			continue
		}

		startTime, err := parseTimestamp(row[1], log)
		if err != nil {
			continue
		}
		endTime, err := parseTimestamp(row[2], log)
		if err != nil {
			continue
		}

		fn(QueryParams{row[0], startTime, endTime})

		currentRow++
		log = slog.With("row_number", currentRow)
	}

	log.Info("csv finished")
	return nil
}

func parseTimestamp(value string, log *slog.Logger) (time.Time, error) {
	t, err := time.Parse(time.DateTime, value)
	if err != nil {
		log.Warn("unexpected timestamp value: %w", err)
	}
	return t, err
}
