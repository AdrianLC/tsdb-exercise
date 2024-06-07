package main

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkerPool(t *testing.T) {
	require := require.New(t)

	// prepare an callback that keeps track of which workers receive the route keys
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
	require.NoError(err)

	worker.Stop()
	for routeKey, seenWorkers := range workerSeenForEachRouteKey {
		require.Len(seenWorkers, 1, "routekey %s received by more than one worker", routeKey)
	}
}
