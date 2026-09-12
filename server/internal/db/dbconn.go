// Package db provides the Postgres connection pool and migration runner
// used to bootstrap Hookwave's database access. Everything here is
// infrastructure setup — the actual queries live in internal/repository.
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// pingTimeout bounds how long Ping waits for the database to respond.
const pingTimeout = 5 * time.Second

// HookWaveDB wraps a pgx connection pool.
type HookWaveDB struct {
	*pgxpool.Pool
}

// New creates a connection pool for the given Postgres connection string.
// It does not verify connectivity — call Ping for that.
func New(conn string) (*HookWaveDB, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, conn)
	if err != nil {
		return nil, err
	}
	return &HookWaveDB{
		Pool: pool,
	}, nil

}

// Ping verifies the database is reachable, bounded by pingTimeout.
func (db *HookWaveDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	return db.Pool.Ping(ctx)
}
