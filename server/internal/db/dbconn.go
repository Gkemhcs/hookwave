package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const pingTimeout = 5 * time.Second

type HookWaveDB struct {
	*pgxpool.Pool
}

func NewHookWaveDBClient(conn string) (*HookWaveDB, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, conn)
	if err != nil {
		return nil, err
	}
	return &HookWaveDB{
		Pool: pool,
	}, nil

}

func (db *HookWaveDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	return db.Pool.Ping(ctx)
}
