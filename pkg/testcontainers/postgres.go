package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/pg"
)

const (
	pgReadyLogOccurrence = 2
	pgStartupTimeout     = 5 * time.Minute
	pgMaxIdleConns       = 10
	pgMaxOpenConns       = 25
	pgMaxConnLifetime    = 5 * time.Minute
	defaultPgImage       = "postgres:18-alpine"
)

// PostgresConfig holds configuration for PostgreSQL testcontainer.
type PostgresConfig struct {
	// Database name
	Database string
	// Username for authentication
	Username string
	// Password for authentication
	Password string
	// Docker image
	Image string
	// Migration file paths (e.g., "../../db/migrations/*.sql")
	MigrationPaths []string
	// Tables to cleanup between tests (for CleanupData)
	CleanupTables []string
	// Custom initialization SQL commands
	InitSQL []string
}

// DefaultPostgresConfig returns a default PostgreSQL configuration.
func DefaultPostgresConfig(database string) *PostgresConfig {
	return &PostgresConfig{
		Database:       database,
		Username:       "testuser",
		Password:       "testpass",
		Image:          defaultPgImage,
		MigrationPaths: []string{},
		CleanupTables:  []string{},
		InitSQL:        []string{},
	}
}

// PostgresContainer wraps the testcontainers PostgreSQL container.
type PostgresContainer struct {
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
	config    *PostgresConfig
	ctx       context.Context
}

// NewPostgresContainer creates a new PostgreSQL container instance.
func NewPostgresContainer(ctx context.Context, config *PostgresConfig) *PostgresContainer {
	if ctx == nil {
		ctx = context.Background()
	}

	if config == nil {
		config = DefaultPostgresConfig("testdb")
	}

	return &PostgresContainer{
		config: config,
		ctx:    ctx,
	}
}

// Start initializes and starts the PostgreSQL container.
func (p *PostgresContainer) Start() error {
	// Use default image if not specified
	image := p.config.Image
	if image == "" {
		image = defaultPgImage
	}

	// Create container options
	opts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase(p.config.Database),
		postgres.WithUsername(p.config.Username),
		postgres.WithPassword(p.config.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(pgReadyLogOccurrence).
				WithStartupTimeout(pgStartupTimeout),
		),
	}

	// Add custom initialization SQL if provided
	if len(p.config.InitSQL) > 0 {
		opts = append(opts, postgres.WithInitScripts(p.config.InitSQL...))
	}

	// Start the container
	container, err := postgres.Run(
		p.ctx,
		image,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("failed to start postgres container: %w", err)
	}

	p.container = container

	// Get container host and port
	host, err := container.Host(p.ctx)
	if err != nil {
		return fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(p.ctx, "5432")
	if err != nil {
		return fmt.Errorf("failed to get mapped port: %w", err)
	}

	port, err := strconv.Atoi(mappedPort.Port())
	if err != nil {
		return fmt.Errorf("failed to parse port: %w", err)
	}

	// Use existing pkg/pg client initialization
	dbConfig := &pg.PostgresConfig{
		Host:            host,
		Port:            port,
		DB:              p.config.Database,
		User:            p.config.Username,
		Password:        p.config.Password,
		SSLMode:         "disable",
		MaxIdleConns:    pgMaxIdleConns,
		MaxOpenConns:    pgMaxOpenConns,
		MaxConnLifetime: pgMaxConnLifetime,
	}

	// Create logger for connection (use simple logger for tests)
	testLogger := logger.NewLogrusLogger(0)

	pool, err := pg.NewPostgresConnection(p.ctx, dbConfig, testLogger)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	p.pool = pool

	return nil
}

// GetPool returns the PostgreSQL connection pool.
func (p *PostgresContainer) GetPool() (*pgxpool.Pool, error) {
	if p.pool == nil {
		return nil, errors.New("postgres pool not initialized")
	}

	return p.pool, nil
}

// GetConnectionString returns the PostgreSQL connection string.
func (p *PostgresContainer) GetConnectionString() (string, error) {
	if p.container == nil {
		return "", errors.New("postgres container not started")
	}

	return p.container.ConnectionString(p.ctx, "sslmode=disable")
}

// RunMigrations executes all configured migration files.
func (p *PostgresContainer) RunMigrations() error {
	if p.pool == nil {
		return errors.New("postgres pool not initialized")
	}

	if len(p.config.MigrationPaths) == 0 {
		return nil // No migrations configured
	}

	return RunMigrations(p.ctx, p.pool, p.config.MigrationPaths)
}

// Terminate stops and removes the PostgreSQL container.
func (p *PostgresContainer) Terminate() error {
	// Close connection pool
	if p.pool != nil {
		p.pool.Close()
	}

	// Terminate container
	if p.container != nil {
		if err := p.container.Terminate(p.ctx); err != nil {
			return fmt.Errorf("failed to terminate postgres container: %w", err)
		}
	}

	return nil
}
