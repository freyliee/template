package handler

import (
	"context"
	"time"

	"github.com/freyliee/template/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager interface {
	Do(
		ctx context.Context,
		fn func(context.Context) error,
	) error
}

type Handler struct {
	pool      *pgxpool.Pool
	timeout   time.Duration
	trips     *storage.TripRepository
	txManager TxManager
}

func New(pool *pgxpool.Pool, timeout time.Duration, trips *storage.TripRepository, txManager TxManager) *Handler {
	return &Handler{
		pool:      pool,
		timeout:   timeout,
		trips:     trips,
		txManager: txManager,
	}
}
