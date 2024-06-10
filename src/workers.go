package main

import (
	"log/slog"
	"sync"
	"time"

	"github.com/cespare/xxhash"
	"github.com/jackc/pgx/v5/pgxpool"
)

const workerChannelSize = 10 // this avoids blocking other workers when a hotspot happens for a given hostname
// The channel can still buffer 10 queries to the same hostname without blocking the routing to other workers

type WorkFunc func(int, QueryParams) // so we can swap it out from tests

type workerPool struct {
	dbpool     *pgxpool.Pool
	wg         sync.WaitGroup
	inputChans []chan QueryParams // one per worker
	workFunc   WorkFunc
	// for stats
	workTimes [][]time.Duration // one slice per worker to be appended after finish
	startTime time.Time
	totalTime time.Duration
}

// NewWorkerPool creates a new worker pool
// The pool can run multiple queries in parallel.
// It forwards the params to the same worker consistently.
func NewWorkerPool(dbpool *pgxpool.Pool) *workerPool {
	return &workerPool{dbpool: dbpool, workFunc: defaulWorkFunc(dbpool)}
}

func defaulWorkFunc(dbpool *pgxpool.Pool) WorkFunc {
	return func(workerIndex int, params QueryParams) {
		executeQuery(dbpool, params)
	}
}

// Start spaws the number of requested workers
func (w *workerPool) Start(numWorkers int) {
	w.wg.Add(numWorkers)
	w.inputChans = make([]chan QueryParams, numWorkers)
	w.workTimes = make([][]time.Duration, numWorkers)
	for i := 0; i < numWorkers; i++ {
		w.inputChans[i] = make(chan QueryParams, workerChannelSize)
		go w.runWorker(i)
	}
	w.startTime = time.Now()
}

func (w *workerPool) Stop() {
	for _, ch := range w.inputChans {
		close(ch)
	}
	w.wg.Wait()
	w.totalTime = time.Since(w.startTime)
}

func (w *workerPool) runWorker(workerIndex int) {
	defer w.wg.Done()
	workTimes := []time.Duration{}
	for params := range w.inputChans[workerIndex] {
		start := time.Now()
		w.workFunc(workerIndex, params)
		workTimes = append(workTimes, time.Since(start))
	}
	w.workTimes[workerIndex] = workTimes
}

// QueryCallback queues the params to run asynchronously in the worker pool
func (w *workerPool) QueryCallback() QueryFunc {
	return func(params QueryParams) error {
		w.routeQuery(params)
		return nil
	}
}

func (w *workerPool) routeQuery(params QueryParams) {
	hash := xxhash.Sum64([]byte(params.RouteKey()))
	workerNum := hash % uint64(len(w.inputChans))
	slog.With("route_key", params.RouteKey(), "worker_num", workerNum).Debug("forwarding params to worker")
	w.inputChans[workerNum] <- params
}

func (w *workerPool) GetStats() *Stats {
	st := &Stats{TotalTime: w.totalTime}
	for _, workTimes := range w.workTimes {
		st.Durations = append(st.Durations, workTimes...)
	}
	return st
}
