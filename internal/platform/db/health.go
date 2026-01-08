package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthCheck checks database connectivity
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return pool.Ping(ctx)
}

