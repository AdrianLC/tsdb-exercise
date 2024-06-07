package main

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type QueryFunc func(QueryParams) error

const sqlQuery = `
	SELECT time_bucket('1 minute', ts, 'UTC') AS bucket,
		min(usage) AS min_cpu,
		max(usage) AS max_cpu
	FROM cpu_usage
	WHERE host = $1 AND ts >= $2 AND ts < $3
	GROUP BY bucket
	ORDER BY bucket
`

type CPUUsage struct {
	Bucket time.Time `db:"bucket"`
	MinCPU float64   `db:"min_cpu"`
	MaxCPU float64   `db:"max_cpu"`
}

type QueryParams struct {
	Host               string
	StartTime, EndTime time.Time
}

type querier struct {
	dbpool *pgxpool.Pool
}

func NewQuerier(dsn string, numConns int) (*querier, error) {
	// since this is for benchmark we want to have a fixed number of ready connections
	// otherwise the first calls would be a bit slower due to the extra time to open new connections
	connStr := dsn + "&pool_min_conns=" + strconv.Itoa(numConns) +
		"&pool_max_conns=" + strconv.Itoa(numConns)

	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		slog.Error("failed to connect to database: %w", err)
		return nil, err
	}
	// force before first query to open all connections
	dbpool.AcquireAllIdle(ctx)
	err = dbpool.Ping(ctx) // pgxpool.New might not really return an error yet

	return &querier{dbpool: dbpool}, err
}

func (q querier) Close() {
	q.dbpool.Close()
}

func (q querier) QueryCallback() QueryFunc {
	return func(params QueryParams) error {
		return q.executeQuery(params)
	}
}

func (q querier) executeQuery(params QueryParams) error {
	log := slog.With("hostname", params.Host, "start_time", params.StartTime, "end_time", params.EndTime)
	log.Info("executing query")

	rows, err := q.dbpool.Query(context.Background(), sqlQuery, params.Host, params.StartTime, params.EndTime)
	if err != nil {
		log.Error("failed to execute query: %w", err)
		return err
	}
	defer rows.Close()

	var numRows int
	for rows.Next() {
		var u CPUUsage
		if err := rows.Scan(&u.Bucket, &u.MinCPU, &u.MaxCPU); err != nil {
			log.Error("failed to scan row: %w", err)
			return err
		}
		slog.With("bucket", u.Bucket, "min_cpu", u.MinCPU, "max_cpu", u.MaxCPU).Debug("row")
		numRows++
	}

	log.With("num_rows", numRows).Debug("query completed")
	return nil
}
