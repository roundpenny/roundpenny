// Copyright (c) 2026 RoundPenny. All rights reserved.

package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	*pgxpool.Pool
}

func NewPool(ctx context.Context) (*Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://roundup:roundup@localhost:5432/roundup?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	var pool *pgxpool.Pool
	err = retryWithBackoff(ctx, func(ctx context.Context) error {
		var err error
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			return fmt.Errorf("create pool: %w", err)
		}
		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return fmt.Errorf("ping: %w", err)
		}
		return nil
	}, 3)
	if err != nil {
		return nil, err
	}

	return &Pool{pool}, nil
}
