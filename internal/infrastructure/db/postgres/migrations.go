package postgres

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// MigrationConfig holds the configuration for database migrations.
type MigrationConfig struct {
	DatabaseURL    string
	MigrationsPath string
}

// RunMigrations executes database migrations.
func RunMigrations(pool *pgxpool.Pool, config MigrationConfig) error {
	// Convert pgx pool to sql.DB for migrate
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing sqlDB: %v", err)
		}
	}()

	// Create postgres driver instance
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	defer func() {
		if err, errCheck := m.Close(); err != nil || errCheck != nil {
			log.Printf("Error closing migrate instance: %v", err)
		}
	}()

	// Run migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// RollbackMigrations rolls back database migrations by N steps.
func RollbackMigrations(pool *pgxpool.Pool, config MigrationConfig, steps int) error {
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing sqlDB: %v", err)
		}
	}()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	defer func() {
		if err, errCheck := m.Close(); err != nil || errCheck != nil {
			log.Printf("Error closing migrate instance: %v", err)
		}
	}()

	if err := m.Steps(-steps); err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	return nil
}

// GetMigrationVersion returns the current migration version.
func GetMigrationVersion(
	pool *pgxpool.Pool,
	config MigrationConfig,
) (version uint, dirty bool, err error) {
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing sqlDB: %v", err)
		}
	}()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	defer func() {
		if err, errCheck := m.Close(); err != nil || errCheck != nil {
			log.Printf("Error closing migrate instance: %v", err)
		}
	}()

	version, dirty, err = m.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}
