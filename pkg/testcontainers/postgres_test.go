package testcontainers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/testcontainers"
)

func TestPostgresContainer_StartAndConnect(t *testing.T) {
	ctx := context.Background()

	// Create PostgreSQL container with default config
	config := testcontainers.DefaultPostgresConfig("testdb")
	pgContainer := testcontainers.NewPostgresContainer(ctx, config)

	// Start container
	err := pgContainer.Start()
	require.NoError(t, err, "Failed to start PostgreSQL container")

	defer func() {
		err = pgContainer.Terminate()
		require.NoError(t, err, "Failed to terminate PostgreSQL container")
	}()

	// Get connection pool
	pool, err := pgContainer.GetPool()
	require.NoError(t, err, "Failed to get connection pool")
	require.NotNil(t, pool, "Connection pool should not be nil")

	// Test connection with a simple query
	var result int

	err = pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err, "Failed to execute test query")
	require.Equal(t, 1, result, "Query result should be 1")
}

func TestPostgresContainer_GetConnectionString(t *testing.T) {
	ctx := context.Background()

	// Create and start container
	config := testcontainers.DefaultPostgresConfig("testdb")
	pgContainer := testcontainers.NewPostgresContainer(ctx, config)

	err := pgContainer.Start()
	require.NoError(t, err, "Failed to start PostgreSQL container")

	defer func() {
		err = pgContainer.Terminate()
		require.NoError(t, err, "Failed to terminate PostgreSQL container")
	}()

	// Get connection string
	connStr, err := pgContainer.GetConnectionString()
	require.NoError(t, err, "Failed to get connection string")
	require.NotEmpty(t, connStr, "Connection string should not be empty")
	require.Contains(
		t,
		connStr,
		"sslmode=disable",
		"Connection string should contain sslmode=disable",
	)
}

func TestPostgresContainer_CreateTable(t *testing.T) {
	ctx := context.Background()

	// Create and start container
	config := testcontainers.DefaultPostgresConfig("testdb")
	pgContainer := testcontainers.NewPostgresContainer(ctx, config)

	err := pgContainer.Start()
	require.NoError(t, err, "Failed to start PostgreSQL container")

	defer func() {
		err = pgContainer.Terminate()
		require.NoError(t, err, "Failed to terminate PostgreSQL container")
	}()

	// Get pool
	pool, err := pgContainer.GetPool()
	require.NoError(t, err, "Failed to get connection pool")

	// Create a test table
	_, err = pool.Exec(ctx, `
		CREATE TABLE test_users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert test data
	_, err = pool.Exec(ctx, `
		INSERT INTO test_users (name, email)
		VALUES ($1, $2)
	`, "John Doe", "john@example.com")
	require.NoError(t, err, "Failed to insert test data")

	// Query test data
	var name, email string

	err = pool.QueryRow(ctx, "SELECT name, email FROM test_users WHERE email = $1", "john@example.com").
		Scan(&name, &email)
	require.NoError(t, err, "Failed to query test data")
	require.Equal(t, "John Doe", name, "Name should match")
	require.Equal(t, "john@example.com", email, "Email should match")
}

func TestPostgresContainer_ErrorBeforeStart(t *testing.T) {
	ctx := context.Background()

	// Create container without starting
	config := testcontainers.DefaultPostgresConfig("testdb")
	pgContainer := testcontainers.NewPostgresContainer(ctx, config)

	// Should return error when getting pool before start
	pool, err := pgContainer.GetPool()
	require.Error(t, err, "Should return error when getting pool before start")
	require.Nil(t, pool, "Pool should be nil before start")

	// Should return error when getting connection string before start
	connStr, err := pgContainer.GetConnectionString()
	require.Error(t, err, "Should return error when getting connection string before start")
	require.Empty(t, connStr, "Connection string should be empty before start")
}
