package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
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

func (p QueryParams) RouteKey() string {
	return p.Host
}

func NewPool(dsn string, numConns int) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error("invalid database connection string", "err", err)
		return nil, err
	}
	// since this is for benchmark we want to have a fixed number of ready connections
	// otherwise the first queries would be a bit slower due to the extra time to open new connections
	config.MinConns = int32(numConns)
	config.MaxConns = int32(numConns)

	ctx := context.Background()
	dbpool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		return nil, err
	}
	// force before first query to open all connections as warm up
	err = forceOpenAllConnections(ctx, dbpool, numConns)
	return dbpool, err
}

func forceOpenAllConnections(ctx context.Context, dbpool *pgxpool.Pool, numConns int) error {
	conns := make([]*pgxpool.Conn, 0, numConns)
	var g errgroup.Group
	for i := 0; i < numConns; i++ {
		g.Go(func() error {
			conn, err := dbpool.Acquire(ctx)
			if err != nil {
				return err
			}
			conns = append(conns, conn)
			return nil
		})
	}
	err := g.Wait()
	for _, conn := range conns {
		conn.Release()
	}
	if err != nil {
		slog.Error("failed to eagerly acquire connections", "err", err)
	}
	return err
}

// QueryCallback runs the query synchronously
func QueryCallback(dbpool *pgxpool.Pool) QueryFunc {
	return func(params QueryParams) error {
		return executeQuery(dbpool, params)
	}
}

func executeQuery(dbpool *pgxpool.Pool, params QueryParams) error {
	log := slog.With("hostname", params.Host, "start_time", params.StartTime, "end_time", params.EndTime)
	log.Info("executing query")

	rows, err := dbpool.Query(context.Background(), sqlQuery, params.Host, params.StartTime, params.EndTime)
	if err != nil {
		log.Error("failed to execute query", "err", err)
		return err
	}
	defer rows.Close()

	var numRows int
	for rows.Next() {
		var u CPUUsage
		if err := rows.Scan(&u.Bucket, &u.MinCPU, &u.MaxCPU); err != nil {
			log.Warn("failed to scan row", "err", err)
			return err
		}
		slog.With("bucket", u.Bucket, "min_cpu", u.MinCPU, "max_cpu", u.MaxCPU).Debug("row")
		numRows++
	}
	if err := rows.Err(); err != nil {
		log.Warn("error reading rows", "err", err)
	}

	log.With("num_rows", numRows).Debug("query completed")
	return nil
}
