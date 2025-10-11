package integration_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	redisclient "github.com/redis/go-redis/v9"
)

const (
	postgresReadyLogOccurrence = 2
	postgresStartupTimeoutMin  = 5
	redisReadyLogMessage       = "Ready to accept connections"
	redisStartupTimeoutMin     = 2
)

// TestContainersSetup holds the testcontainers setup.
type TestContainersSetup struct {
	PgContainer    *postgres.PostgresContainer
	RedisContainer *redis.RedisContainer
	DBPool         *pgxpool.Pool
	RedisClient    redisclient.UniversalClient
	ctx            context.Context
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
		postgres.WithDatabase("chat_test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(postgresReadyLogOccurrence).
				WithStartupTimeout(postgresStartupTimeoutMin*time.Minute)),
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

	tc.DBPool = dbPool

	// Run migrations
	return tc.runMigrations()
}

// SetupRedis sets up Redis container for pub/sub.
func (tc *TestContainersSetup) SetupRedis() error {
	// Setup Redis container
	redisContainer, err := redis.Run(tc.ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog(redisReadyLogMessage).
				WithStartupTimeout(redisStartupTimeoutMin*time.Minute)),
	)
	if err != nil {
		return err
	}

	tc.RedisContainer = redisContainer

	// Get connection string
	connStr, err := redisContainer.ConnectionString(tc.ctx)
	if err != nil {
		return err
	}

	// Parse connection string to remove "redis://" prefix if present
	addr := connStr
	if len(connStr) > 8 && connStr[:8] == "redis://" {
		addr = connStr[8:]
	}

	// Use UniversalClient which works with both single and cluster Redis
	tc.RedisClient = redisclient.NewUniversalClient(&redisclient.UniversalOptions{
		Addrs: []string{addr},
	})

	// Test connection
	if err = tc.RedisClient.Ping(tc.ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// Cleanup tears down the containers and closes connections.
func (tc *TestContainersSetup) Cleanup() {
	if tc.DBPool != nil {
		tc.DBPool.Close()
	}

	if tc.RedisClient != nil {
		_ = tc.RedisClient.Close()
	}

	if tc.PgContainer != nil {
		if err := tc.PgContainer.Terminate(tc.ctx); err != nil {
			panic("failed to terminate PostgreSQL container: " + err.Error())
		}
	}

	if tc.RedisContainer != nil {
		if err := tc.RedisContainer.Terminate(tc.ctx); err != nil {
			panic("failed to terminate Redis container: " + err.Error())
		}
	}
}

// CleanupData cleans up test data from tables.
func (tc *TestContainersSetup) CleanupData() error {
	// Use TRUNCATE instead of DELETE for better performance and to reset sequences
	// Clean up in proper order due to foreign key constraints
	_, err := tc.DBPool.Exec(
		tc.ctx,
		"TRUNCATE TABLE messages, participants, conversations, connections RESTART IDENTITY CASCADE",
	)

	// Flush Redis
	if tc.RedisClient != nil {
		if err = tc.RedisClient.FlushAll(tc.ctx).Err(); err != nil {
			return fmt.Errorf("failed to flush Redis: %w", err)
		}
	}

	return err
}

// GetPostgresConnectionString returns the PostgreSQL connection string.
func (tc *TestContainersSetup) GetPostgresConnectionString() (string, error) {
	if tc.PgContainer == nil {
		return "", errors.New("PostgreSQL container not initialized")
	}

	return tc.PgContainer.ConnectionString(tc.ctx, "sslmode=disable")
}

// GetRedisAddr returns the Redis address.
func (tc *TestContainersSetup) GetRedisAddr() (string, error) {
	if tc.RedisContainer == nil {
		return "", errors.New("Redis container not initialized")
	}

	return tc.RedisContainer.ConnectionString(tc.ctx)
}

// runMigrations runs the chat-service database migrations.
func (tc *TestContainersSetup) runMigrations() error {
	// Run the actual migration files
	migrationFiles := []string{
		"../../db/migrations/000001_create_conversations_table.up.sql",
		"../../db/migrations/000002_create_messages_table.up.sql",
		"../../db/migrations/000003_create_participants_table.up.sql",
		"../../db/migrations/000004_create_connections_table.up.sql",
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
		return errors.New("invalid file type: only .sql files are allowed")
	}

	// Additional validation: ensure path contains migrations directory
	if !strings.Contains(absPath, "migrations") {
		return errors.New("invalid path: file must be in migrations directory")
	}

	// Read the migration file
	query, err := os.ReadFile(absPath) // #nosec G304 - path is validated above
	if err != nil {
		return err
	}

	// Execute the migration
	_, err = tc.DBPool.Exec(tc.ctx, string(query))

	return err
}
