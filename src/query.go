package main

import (
	"context"
	"strconv"
	"time"

	"log/slog"

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

type PoolOptions struct {
	MinConns, MaxConns int
}

func NewQuerier(dsn string, poolOpts PoolOptions) (*querier, error) {
	connStr := dsn + "&pool_min_conns=" + strconv.Itoa(poolOpts.MinConns) +
		"&pool_max_conns=" + strconv.Itoa(poolOpts.MaxConns)

	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		slog.Error("failed to connect to database: %w", err)
		return nil, err
	}

	return &querier{dbpool: dbpool}, nil
}

func (q *querier) Close() {
	q.dbpool.Close()
}

func (q *querier) QueryCallback() QueryFunc {
	return func(params QueryParams) error {
		return q.executeQuery(params)
	}
}

func (q *querier) executeQuery(params QueryParams) error {
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
