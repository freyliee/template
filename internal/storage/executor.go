package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Executor interface {
	Exec(ctx context.Context, sqlQuery string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sqlQuery string, args ...any) pgx.Row
}

func executorFromContext(ctx context.Context, pool *pgxpool.Pool) Executor {
	value, ok := ctx.Value(txKey{}).(pgx.Tx)
	if ok {
		return value
	}
	return pool
}
