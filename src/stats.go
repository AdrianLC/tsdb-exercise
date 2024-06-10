package main

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"golang.org/x/exp/constraints"
)

type Stats struct {
	TotalTime time.Duration
	Durations []time.Duration
	sorted    bool
}

func (st *Stats) Print() {
	slog.Info("benchmark statistics:\n" + st.String())
}

func (st *Stats) String() string {
	return fmt.Sprintf(
		`	total_processing_time=%v,
	total_sum_query_time=%v,
	min_query_time=%v,
	median_query_time=%v,
	average_query_time=%v,
	max_query_time=%v`,
		st.TotalTime, st.SumTime(), st.MinTime(), st.MedianTime(), st.AvgTime(), st.MaxTime())
}

func (st *Stats) ensureSorted() {
	if !st.sorted {
		slices.Sort(st.Durations)
		st.sorted = true
	}
}

func (st *Stats) MinTime() time.Duration {
	st.ensureSorted()
	return st.Durations[0]
}

func (st *Stats) MaxTime() time.Duration {
	st.ensureSorted()
	return st.Durations[len(st.Durations)-1]
}

func (st *Stats) MedianTime() time.Duration {
	st.ensureSorted()
	return mean(st.Durations)
}

func (st *Stats) SumTime() time.Duration {
	return sum(st.Durations)
}

func (st *Stats) AvgTime() time.Duration {
	return avg(st.Durations)
}

type Number interface {
	constraints.Integer | constraints.Float
}

func mean[N Number](arr []N) N {
	l := len(arr)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		return avg(arr[l/2-1 : l/2+1])
	} else {
		return arr[l/2]
	}
}

func avg[N Number](arr []N) N {
	if len(arr) == 0 {
		return 0
	}
	return sum(arr) / N(len(arr))
}

func sum[N Number](arr []N) N {
	var sum N
	for _, v := range arr {
		sum += v
	}
	return sum
}
