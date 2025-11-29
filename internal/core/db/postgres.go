package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetDB(ctx context.Context, url string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	slog.InfoContext(ctx, "Database connected successfully")
	return conn, nil
}
