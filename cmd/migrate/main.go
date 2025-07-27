package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
		// Try to get from environment
		*databaseURL = os.Getenv("DATABASE_URL")
		if *databaseURL == "" {
			log.Fatal(
				"Database URL is required. Use -database-url flag or DATABASE_URL environment variable",
			)
		}
	}

	// Get absolute path for migrations
	absPath, err := filepath.Abs(*migrationsPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for migrations: %v", err)
	}

	// Connect to database using pgx
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
		fmt.Println("Running migrations...")

		if err := postgresInfra.RunMigrations(pool, config); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		fmt.Println("Migrations completed successfully!")

	case "down":
		fmt.Printf("Rolling back %d migration(s)...\n", *steps)

		if err := postgresInfra.RollbackMigrations(pool, config, *steps); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}

		fmt.Println("Rollback completed successfully!")

	case "version":
		version, dirty, err := postgresInfra.GetMigrationVersion(pool, config)
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}

		fmt.Printf("Current migration version: %d\n", version)

		if dirty {
			fmt.Println("Warning: Database is in dirty state")
		}

	default:
		log.Fatalf("Unknown action: %s. Use 'up', 'down', or 'version'", *action)
	}
}
