package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStatsString(t *testing.T) {
	stats := Stats{
		TotalTime: 10 * time.Nanosecond,
		Durations: []time.Duration{1, 2, 2, 3, 9},
	}
	require.Equal(t,
		`	total_processing_time=10ns,
	total_sum_query_time=17ns,
	min_query_time=1ns,
	median_query_time=2ns,
	average_query_time=3ns,
	max_query_time=9ns`, stats.String())
}

func TestSum(t *testing.T) {
	require := require.New(t)
	require.Zero(sum([]time.Duration{}))
	require.Equal(time.Duration(10), sum([]time.Duration{1, 2, 3, 2, 2}))
}

func TestAvg(t *testing.T) {
	require := require.New(t)
	require.Zero(avg([]time.Duration{}))
	require.Equal(time.Duration(2), avg([]time.Duration{1, 2, 3, 2, 2}))
}

func TestMean(t *testing.T) {
	require := require.New(t)
	require.Zero(mean([]time.Duration{}))
	require.Equal(time.Duration(2), mean([]time.Duration{1, 1, 2, 2, 3}))
	require.Equal(time.Duration(1), mean([]time.Duration{1, 1, 2, 3}))
}
