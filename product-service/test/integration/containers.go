package integration

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestContainersSetup holds the testcontainers setup.
type TestContainersSetup struct {
	PgContainer testcontainers.Container
	DbPool      *pgxpool.Pool
	ctx         context.Context
}

// NewTestContainersSetup creates a new testcontainers setup.
func NewTestContainersSetup() *TestContainersSetup {
	return &TestContainersSetup{
		ctx: context.Background(),
	}
}

// SetupPostgres sets up PostgreSQL container and connection pool.
func (tc *TestContainersSetup) SetupPostgres() error {
	// Setup PostgreSQL container
	pgContainer, err := postgres.Run(tc.ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute)),
	)
	if err != nil {
		return err
	}

	tc.PgContainer = pgContainer

	// Get connection string
	connStr, err := pgContainer.ConnectionString(tc.ctx, "sslmode=disable")
	if err != nil {
		return err
	}

	// Setup database connection pool
	dbPool, err := pgxpool.New(tc.ctx, connStr)
	if err != nil {
		return err
	}

	tc.DbPool = dbPool

	// Create products table
	return tc.createProductsTable()
}

// Cleanup tears down the containers and closes connections.
func (tc *TestContainersSetup) Cleanup() {
	if tc.DbPool != nil {
		tc.DbPool.Close()
	}

	if tc.PgContainer != nil {
		err := tc.PgContainer.Terminate(tc.ctx)
		if err != nil {
			panic("failed to terminate PostgreSQL container: " + err.Error())
		}
	}
}

// CleanupData cleans up test data from tables.
func (tc *TestContainersSetup) CleanupData() error {
	// Use TRUNCATE instead of DELETE for better performance and to reset sequences
	_, err := tc.DbPool.Exec(tc.ctx, "TRUNCATE TABLE products RESTART IDENTITY CASCADE")

	return err
}

// createProductsTable creates the products table for testing.
func (tc *TestContainersSetup) createProductsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			quantity INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
		CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);
	`

	_, err := tc.DbPool.Exec(tc.ctx, query)

	return err
}
