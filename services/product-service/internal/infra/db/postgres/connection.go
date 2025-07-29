package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewConnection creates a new database connection pool
func NewConnection(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool
	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// CreateProductsTable creates the products table if it doesn't exist
func CreateProductsTable(pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL CHECK (price > 0),
			seller_id UUID NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_products_seller_id ON products(seller_id);
		CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);
	`

	_, err := pool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	return nil
}
