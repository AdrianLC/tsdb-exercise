package main

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkerPoolRouting(t *testing.T) {
	// prepare a callback that keeps track of which workers receive the route keys
	workerSeenForEachRouteKey := map[string]map[int]struct{}{}
	var lock sync.RWMutex
	fn := func(workerIndex int, params QueryParams) {
		lock.Lock()
		defer lock.Unlock()

		seenWorkers, ok := workerSeenForEachRouteKey[params.RouteKey()]
		if !ok {
			seenWorkers = map[int]struct{}{}
		}
		seenWorkers[workerIndex] = struct{}{}
		workerSeenForEachRouteKey[params.RouteKey()] = seenWorkers
	}
	worker := &workerPool{workFunc: fn}
	worker.Start(10)

	err := StreamParamsFilePath(csvFilePath, worker.QueryCallback())
	require.NoError(t, err)

	worker.Stop()
	for routeKey, seenWorkers := range workerSeenForEachRouteKey {
		require.Len(t, seenWorkers, 1, "routekey %s received by more than one worker", routeKey)
	}
}

func TestWorkerPoolStats(t *testing.T) {
	var calls atomic.Int32
	fn := func(int, QueryParams) {
		calls.Add(1)
	}
	worker := &workerPool{workFunc: fn}
	worker.Start(10)

	err := StreamParamsFilePath(csvFilePath, worker.QueryCallback())
	require.NoError(t, err)

	worker.Stop()
	stats := worker.GetStats()
	require.NotZero(t, stats.TotalTime)
	require.Len(t, stats.Durations, int(calls.Load()))
}
