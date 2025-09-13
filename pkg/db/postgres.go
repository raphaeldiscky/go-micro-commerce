// Package db provides a PostgreSQL connection pool implementation.
package db

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// PostgresConfig holds the configuration for the PostgreSQL connection.
type PostgresConfig struct {
	Host            string
	DB              string
	User            string
	Password        string
	SSLMode         string
	Port            int
	MaxIdleConns    int
	MaxOpenConns    int
	MaxConnLifetime time.Duration
}

// NewPostgresConnection creates a new PostgreSQL connection pool.
func NewPostgresConnection(
	ctx context.Context,
	cfg *PostgresConfig,
	appLogger logger.Logger,
) (*pgxpool.Pool, error) {
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.User, cfg.Password, hostPort, cfg.DB, cfg.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set pool configuration
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	appLogger.Info("PostgreSQL connection established to %s:%d/%s", cfg.Host, cfg.Port, cfg.DB)

	return pool, nil
}
