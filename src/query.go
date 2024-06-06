package main

import "time"

type QueryParams struct {
	hostname           string
	startTime, endTime time.Time
}
