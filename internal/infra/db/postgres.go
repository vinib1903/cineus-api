package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresConfig contém as configurações de conexão.
type PostgresConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DefaultPostgresConfig retorna configurações padrão.
func DefaultPostgresConfig(url string) PostgresConfig {
	return PostgresConfig{
		URL:             url,
		MaxConns:        10,               // Máximo de conexões no pool
		MinConns:        2,                // Mínimo de conexões mantidas abertas
		MaxConnLifetime: 1 * time.Hour,    // Tempo máximo de vida de uma conexão
		MaxConnIdleTime: 30 * time.Minute, // Tempo máximo de uma conexão ociosa
	}
}

// NewPostgresPool cria um novo pool de conexões com o PostgreSQL.
func NewPostgresPool(ctx context.Context, cfg PostgresConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
