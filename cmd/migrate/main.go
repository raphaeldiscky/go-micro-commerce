// Package main implements the command-line tool for managing database migrations.
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"

	postgresInfra "github.com/raphaeldiscky/go-ddd-template/internal/infrastructure/db/postgres"
)

func main() {
	var (
		databaseURL    = flag.String("database-url", "", "Database connection URL")
		migrationsPath = flag.String("migrations-path", "./migrations", "Path to migration files")
		action         = flag.String("action", "up", "Migration action: up, down, version")
		steps          = flag.Int("steps", 1, "Number of migration steps for rollback")
	)

	flag.Parse()

	if *databaseURL == "" {
		*databaseURL = os.Getenv("DATABASE_URL")
		if *databaseURL == "" {
			log.Fatal(
				"Database URL is required. Use -database-url flag or DATABASE_URL environment variable",
			)
		}
	}

	absPath, err := filepath.Abs(*migrationsPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for migrations: %v", err)
	}

	pool, err := postgresInfra.NewConnection(*databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	config := postgresInfra.MigrationConfig{
		DatabaseURL:    *databaseURL,
		MigrationsPath: absPath,
	}

	switch *action {
	case "up":
		runMigrations(pool, config)
	case "down":
		rollbackMigrations(pool, config, *steps)
	case "version":
		printMigrationVersion(pool, config)
	}
}

func runMigrations(pool *pgxpool.Pool, config postgresInfra.MigrationConfig) {
	log.Println("Running migrations...")

	if err := postgresInfra.RunMigrations(pool, config); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully!")
}

func rollbackMigrations(pool *pgxpool.Pool, config postgresInfra.MigrationConfig, steps int) {
	log.Printf("Rolling back %d migration(s)...", steps)

	if err := postgresInfra.RollbackMigrations(pool, config, steps); err != nil {
		log.Fatalf("Failed to rollback migrations: %v", err)
	}

	log.Println("Rollback completed successfully!")
}

func printMigrationVersion(pool *pgxpool.Pool, config postgresInfra.MigrationConfig) {
	version, dirty, err := postgresInfra.GetMigrationVersion(pool, config)
	if err != nil {
		log.Fatalf("Failed to get migration version: %v", err)
	}

	log.Printf("Current migration version: %d", version)

	if dirty {
		log.Println("Warning: Database is in dirty state")
	}
}
