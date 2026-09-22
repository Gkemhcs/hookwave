package transactor

import (
	"context"

	"github.com/Gkemhcs/hookwave/server/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionExecFunc func(queries *repository.Queries) error



type Transactor struct {
	db      *pgxpool.Pool
	queries *repository.Queries
}

func New(db *pgxpool.Pool,queries *repository.Queries)*Transactor{
	return &Transactor{
		db: db,
		queries: queries,
	}
}

func (t *Transactor) Execute(ctx context.Context, execFunc TransactionExecFunc) error {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	queries := t.queries.WithTx(tx)
	err = execFunc(queries)
	if err == nil {
		if err :=tx.Commit(ctx);err!=nil{
			return err
		}
		return nil
	}
	return err

}
