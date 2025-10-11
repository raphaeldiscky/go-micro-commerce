package integration_test

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/testcontainers"
)

// TestContainersSetup holds the testcontainers setup.
type TestContainersSetup struct {
	manager *testcontainers.ContainerManager
	DBPool  *pgxpool.Pool
	ctx     context.Context
}

// NewTestContainersSetup creates a new testcontainers setup.
func NewTestContainersSetup() *TestContainersSetup {
	return &TestContainersSetup{
		ctx: context.Background(),
	}
}

// SetupPostgres sets up PostgreSQL container and connection pool.
func (tc *TestContainersSetup) SetupPostgres() error {
	// Configure PostgreSQL with migrations
	pgConfig := testcontainers.DefaultPostgresConfig("testdb")
	pgConfig.MigrationPaths = []string{
		"../../db/migrations/000001_create_table_products.up.sql",
		"../../db/migrations/000002_add_version_and_reserved_quantity_for_optimistic_locking.up.sql",
	}
	pgConfig.CleanupTables = []string{"products"}

	// Create container manager
	tc.manager = testcontainers.NewContainerManager().
		WithContext(tc.ctx).
		WithPostgres(pgConfig)

	// Start containers
	if err := tc.manager.Start(); err != nil {
		return err
	}

	// Get PostgreSQL pool
	pool, err := tc.manager.GetPostgresPool()
	if err != nil {
		return err
	}

	tc.DBPool = pool

	return nil
}

// Cleanup tears down the containers and closes connections.
func (tc *TestContainersSetup) Cleanup() {
	if tc.manager != nil {
		if err := tc.manager.Cleanup(); err != nil {
			panic("failed to cleanup containers: " + err.Error())
		}
	}
}

// CleanupData cleans up test data from tables.
func (tc *TestContainersSetup) CleanupData() error {
	if tc.manager != nil {
		return tc.manager.CleanupData()
	}

	return nil
}
