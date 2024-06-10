package main

import (
	"log/slog"
	"sync"
	"time"

	"github.com/cespare/xxhash"
	"github.com/jackc/pgx/v5/pgxpool"
)

const workerChannelSize = 10

type WorkFunc func(int, QueryParams) // so we can swap it out from tests

type workerPool struct {
	dbpool     *pgxpool.Pool
	wg         sync.WaitGroup
	inputChans []chan QueryParams
	workFunc   WorkFunc
}

// NewWorkerPool creates a new worker pool
// The pool can run multiple queries in parallel.
// It forwards the params to the same worker consistently.
func NewWorkerPool(dbpool *pgxpool.Pool, stats Stats) *workerPool {
	return &workerPool{dbpool: dbpool, workFunc: defaulWorkFunc(dbpool, stats)}
}

func defaulWorkFunc(dbpool *pgxpool.Pool, stats Stats) WorkFunc {
	return func(workerIndex int, params QueryParams) {
		start := time.Now()
		executeQuery(dbpool, params)
		stats.Record(time.Since(start))
	}
}

// Start spaws the number of requested workers
func (w *workerPool) Start(numWorkers int) {
	w.wg.Add(numWorkers)
	w.inputChans = make([]chan QueryParams, numWorkers)
	for i := 0; i < numWorkers; i++ {
		w.inputChans[i] = make(chan QueryParams, workerChannelSize)
		go w.runWorker(i)
	}
}

func (w *workerPool) Stop() {
	for _, ch := range w.inputChans {
		close(ch)
	}
	w.wg.Wait()
}

func (w *workerPool) runWorker(workerIndex int) {
	defer w.wg.Done()
	for params := range w.inputChans[workerIndex] {
		w.workFunc(workerIndex, params)
	}
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
