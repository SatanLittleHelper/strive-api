package database

import (
	"context"
	"fmt"
	"time"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func New(cfg *config.Config, log *logger.Logger) (*Database, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolConfig.MaxConns = cfg.DB.MaxConns
	poolConfig.MinConns = cfg.DB.MinConns
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Database connected successfully",
		"host", cfg.DB.Host,
		"port", cfg.DB.Port,
		"database", cfg.DB.DBName,
		"max_conns", cfg.DB.MaxConns,
		"min_conns", cfg.DB.MinConns,
	)

	return &Database{
		pool:   pool,
		logger: log,
	}, nil
}

func (db *Database) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *Database) Close() {
	db.pool.Close()
	db.logger.Info("Database connection pool closed")
}

func (db *Database) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return db.pool.Ping(ctx)
}
