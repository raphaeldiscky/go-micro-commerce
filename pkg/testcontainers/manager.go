// Package testcontainers provides utilities for working with testcontainers.
package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	redisclient "github.com/redis/go-redis/v9"
)

// ContainerManager manages multiple testcontainers for integration testing.
type ContainerManager struct {
	ctx context.Context

	// PostgreSQL
	postgres      *PostgresContainer
	dbPool        *pgxpool.Pool
	postgresReady bool

	// Redis
	redis      *RedisContainer
	redisReady bool

	// Error tracking
	errors []error
	mu     sync.RWMutex
}

// NewContainerManager creates a new container manager.
func NewContainerManager() *ContainerManager {
	return &ContainerManager{
		ctx:    context.Background(),
		errors: make([]error, 0),
	}
}

// WithContext sets a custom context for container operations.
func (m *ContainerManager) WithContext(ctx context.Context) *ContainerManager {
	m.ctx = ctx
	return m
}

// WithPostgres configures PostgreSQL container.
// Parameters:
//   - config: PostgreSQL configuration
func (m *ContainerManager) WithPostgres(config *PostgresConfig) *ContainerManager {
	m.postgres = &PostgresContainer{
		config: config,
		ctx:    m.ctx,
	}

	return m
}

// WithRedis configures Redis container.
// Parameters:
//   - config: Redis configuration (optional, uses defaults if nil)
func (m *ContainerManager) WithRedis(config *RedisConfig) *ContainerManager {
	if config == nil {
		config = DefaultRedisConfig()
	}

	m.redis = &RedisContainer{
		config: config,
		ctx:    m.ctx,
	}

	return m
}

const (
	bufferSize = 2
)

// Start initializes and starts all configured containers concurrently.
func (m *ContainerManager) Start() error {
	var wg sync.WaitGroup

	errChan := make(chan error, bufferSize) // Buffer for postgres and redis errors

	// Start PostgreSQL
	if m.postgres != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := m.startPostgres(); err != nil {
				errChan <- fmt.Errorf("postgres setup failed: %w", err)
			}
		}()
	}

	// Start Redis
	if m.redis != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := m.startRedis(); err != nil {
				errChan <- fmt.Errorf("redis setup failed: %w", err)
			}
		}()
	}

	// Wait for all containers to start
	wg.Wait()
	close(errChan)

	// Collect errors
	for err := range errChan {
		m.addError(err)
	}

	if len(m.errors) > 0 {
		return fmt.Errorf(
			"container startup failed with %d error(s): %w",
			len(m.errors),
			m.errors[0],
		)
	}

	return nil
}

// startPostgres starts the PostgreSQL container and runs migrations.
func (m *ContainerManager) startPostgres() error {
	if err := m.postgres.Start(); err != nil {
		return err
	}

	// Get connection pool
	pool, err := m.postgres.GetPool()
	if err != nil {
		return err
	}

	m.dbPool = pool
	m.postgresReady = true

	// Run migrations if configured
	if len(m.postgres.config.MigrationPaths) > 0 {
		if err = m.postgres.RunMigrations(); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// startRedis starts the Redis container.
func (m *ContainerManager) startRedis() error {
	if err := m.redis.Start(); err != nil {
		return err
	}

	m.redisReady = true

	return nil
}

// GetPostgresPool returns the PostgreSQL connection pool.
func (m *ContainerManager) GetPostgresPool() (*pgxpool.Pool, error) {
	if !m.postgresReady {
		return nil, errors.New("postgres container not ready")
	}

	return m.dbPool, nil
}

// GetPostgresConnectionString returns the PostgreSQL connection string.
func (m *ContainerManager) GetPostgresConnectionString() (string, error) {
	if m.postgres == nil {
		return "", errors.New("postgres not configured")
	}

	return m.postgres.GetConnectionString()
}

// GetRedisClient returns the Redis client.
func (m *ContainerManager) GetRedisClient() (*redisclient.Client, error) {
	if !m.redisReady {
		return nil, errors.New("redis container not ready")
	}

	return m.redis.GetClient()
}

// GetRedisAddr returns the Redis address.
func (m *ContainerManager) GetRedisAddr() (string, error) {
	if m.redis == nil {
		return "", errors.New("redis not configured")
	}

	return m.redis.GetAddr()
}

// CleanupData truncates all tables in PostgreSQL and flushes Redis.
func (m *ContainerManager) CleanupData() error {
	var errs []error

	// Cleanup PostgreSQL
	if m.postgresReady && m.dbPool != nil {
		if len(m.postgres.config.CleanupTables) > 0 {
			if err := TruncateTables(m.ctx, m.dbPool, m.postgres.config.CleanupTables); err != nil {
				errs = append(errs, fmt.Errorf("postgres cleanup failed: %w", err))
			}
		}
	}

	// Cleanup Redis
	if m.redisReady && m.redis != nil {
		client, err := m.redis.GetClient()
		if err == nil {
			if err = FlushRedis(m.ctx, client); err != nil {
				errs = append(errs, fmt.Errorf("redis cleanup failed: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// Cleanup terminates all containers and closes connections.
func (m *ContainerManager) Cleanup() error {
	var errs []error

	// Close PostgreSQL pool
	if m.dbPool != nil {
		m.dbPool.Close()
	}

	// Terminate PostgreSQL container
	if m.postgres != nil {
		if err := m.postgres.Terminate(); err != nil {
			errs = append(errs, fmt.Errorf("postgres terminate failed: %w", err))
		}
	}

	// Terminate Redis container
	if m.redis != nil {
		if err := m.redis.Terminate(); err != nil {
			errs = append(errs, fmt.Errorf("redis terminate failed: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// addError adds an error to the error list (thread-safe).
func (m *ContainerManager) addError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errors = append(m.errors, err)
}

// GetErrors returns all errors encountered during container operations.
func (m *ContainerManager) GetErrors() []error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.errors
}
