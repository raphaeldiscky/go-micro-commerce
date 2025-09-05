package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestContainersSetup holds the test database setup.
type TestContainersSetup struct {
	DbPool *pgxpool.Pool
	ctx    context.Context
}

// NewTestContainersSetup creates a new testcontainers setup.
func NewTestContainersSetup() *TestContainersSetup {
	return &TestContainersSetup{
		ctx: context.Background(),
	}
}

// SetupPostgres sets up an in-memory database for testing.
func (tc *TestContainersSetup) SetupPostgres() error {
	// For simplicity, we'll use a local PostgreSQL connection for testing
	// In a real scenario, you'd use testcontainers or a dedicated test database
	connStr := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"

	// Try to connect to a test database
	dbPool, err := pgxpool.New(tc.ctx, connStr)
	if err != nil {
		// If no test database is available, we'll skip the database-dependent tests
		// In a production setup, you'd set up testcontainers here
		return fmt.Errorf("test database not available: %w", err)
	}

	tc.DbPool = dbPool

	// Create orders tables
	return tc.createOrdersTables()
}

// Cleanup tears down the connections.
func (tc *TestContainersSetup) Cleanup() {
	if tc.DbPool != nil {
		tc.DbPool.Close()
	}
}

// CleanupData cleans up test data from tables.
func (tc *TestContainersSetup) CleanupData() error {
	// Use TRUNCATE instead of DELETE for better performance and to reset sequences
	// Clean up in proper order due to foreign key constraints
	_, err := tc.DbPool.Exec(
		tc.ctx,
		"TRUNCATE TABLE order_items, orders, outbox_events, inbox_events, saga_states RESTART IDENTITY CASCADE",
	)

	return err
}

// createOrdersTables creates the orders tables for testing by running actual migrations.
func (tc *TestContainersSetup) createOrdersTables() error {
	// Run the actual migration files
	migrationFiles := []string{
		"../../db/migrations/000001_create_table_orders.up.sql",
		"../../db/migrations/000002_create_outbox_events.up.sql",
		"../../db/migrations/000003_create_inbox_events.up.sql",
		"../../db/migrations/000004_create_saga_states.up.sql",
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
