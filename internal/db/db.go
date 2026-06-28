// Package db owns the connection to PostgreSQL.
//
// We use pgxpool: a "pool" of reusable DB connections. Opening a new connection
// per request is slow, so the pool keeps a handful open and hands them out.
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a PostgreSQL connection pool and verifies it works.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}
