package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	slog.Info("ping")
	return pool.Ping(ctx)
}
