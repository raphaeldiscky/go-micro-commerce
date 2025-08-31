package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// createProductsTable creates the products table for testing by running actual migrations.
func (tc *TestContainersSetup) createProductsTable() error {
	// Run the actual migration files
	migrationFiles := []string{
		"../../db/migrations/000001_create_table_products.up.sql",
		"../../db/migrations/000002_add_version_and_reserved_quantity_for_optimistic_locking.up.sql",
	}

	for _, migrationFile := range migrationFiles {
		if err := tc.runMigrationFile(migrationFile); err != nil {
			return err
		}
	}

	return nil
}

// runMigrationFile reads and executes a migration file.
func (tc *TestContainersSetup) runMigrationFile(filePath string) error {
	// Get the absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// Security check: ensure the file is a .sql file in the migrations directory
	if !strings.HasSuffix(absPath, ".sql") {
		return fmt.Errorf("invalid file type: only .sql files are allowed")
	}

	// Additional validation: ensure path contains migrations directory
	if !strings.Contains(absPath, "migrations") {
		return fmt.Errorf("invalid path: file must be in migrations directory")
	}

	// Read the migration file
	query, err := os.ReadFile(absPath) // #nosec G304 - path is validated above
	if err != nil {
		return err
	}

	// Execute the migration
	_, err = tc.DbPool.Exec(tc.ctx, string(query))

	return err
}
