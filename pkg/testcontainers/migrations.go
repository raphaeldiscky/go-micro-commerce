package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations executes SQL migration files from the given paths.
// Supports glob patterns like "../../db/migrations/*.sql".
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationPaths []string) error {
	if len(migrationPaths) == 0 {
		return nil
	}

	// Collect all migration files
	var migrationFiles []string

	for _, pattern := range migrationPaths {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid migration path pattern %q: %w", pattern, err)
		}

		migrationFiles = append(migrationFiles, matches...)
	}

	if len(migrationFiles) == 0 {
		return fmt.Errorf("no migration files found for patterns: %v", migrationPaths)
	}

	// Sort files alphabetically (standard migration order)
	sort.Strings(migrationFiles)

	// Execute each migration file
	for _, file := range migrationFiles {
		if err := executeMigrationFile(ctx, pool, file); err != nil {
			return fmt.Errorf("migration failed for %q: %w", file, err)
		}
	}

	return nil
}

// executeMigrationFile reads and executes a single SQL migration file.
func executeMigrationFile(ctx context.Context, pool *pgxpool.Pool, filePath string) error {
	// Get absolute path to handle relative paths correctly
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Security check: ensure the file is a .sql file
	if !strings.HasSuffix(absPath, ".sql") {
		return errors.New("invalid file type: only .sql files are allowed")
	}

	// Read file content
	content, err := os.ReadFile(absPath) // #nosec G304 - path is validated above
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	sql := string(content)

	// Basic validation - check for empty files
	if strings.TrimSpace(sql) == "" {
		return errors.New("migration file is empty")
	}

	// Execute the migration
	_, err = pool.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
