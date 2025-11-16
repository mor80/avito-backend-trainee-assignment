package postgres

import (
	"context"
	"fmt"
	"time"

	"mor80/service-reviewer/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg config.Postgres) (*pgxpool.Pool, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pgxCfg.MaxConns = 10
	pgxCfg.MinConns = 2
	pgxCfg.MaxConnLifetime = time.Hour
	pgxCfg.MaxConnIdleTime = 30 * time.Minute

	var pool *pgxpool.Pool
	const attempts = 5

	for i := 1; i <= attempts; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, pgxCfg)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			pingErr := pool.Ping(pingCtx)
			cancel()

			if pingErr == nil {
				return pool, nil
			}

			err = fmt.Errorf("ping attempt %d/%d failed: %w", i, attempts, pingErr)
			pool.Close()
		} else {
			err = fmt.Errorf("new pool attempt %d/%d failed: %w", i, attempts, err)
		}

		if i < attempts {
			time.Sleep(time.Second * time.Duration(i))
		}
	}

	return nil, err
}
