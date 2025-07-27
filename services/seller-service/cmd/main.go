package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/application/services"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/config"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/infrastructure/messaging/kafka"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/infrastructure/persistence/postgres"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/interface/http/handlers"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/interface/http/server"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup database connection
	dbPool, err := pgxpool.New(context.Background(), cfg.GetDatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Test database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established")

	// Create sellers table
	if err := createSellersTable(dbPool); err != nil {
		log.Fatalf("Failed to create sellers table: %v", err)
	}
	log.Println("Database schema initialized")

	// Setup Kafka event publisher
	eventPublisher, err := kafka.NewEventPublisherKafka(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		log.Printf("Warning: Failed to setup Kafka event publisher: %v", err)
		log.Println("Continuing without event publishing...")
		eventPublisher = nil
	} else {
		defer eventPublisher.Close()
		log.Println("Kafka event publisher initialized")
	}

	// Setup repositories
	sellerRepo := postgres.NewSellerRepositoryPostgres(dbPool)

	// Setup services
	sellerService := services.NewSellerService(sellerRepo, eventPublisher)

	// Setup HTTP handlers
	sellerHandler := handlers.NewSellerHandler(sellerService)

	// Setup HTTP server
	httpServer := server.NewHTTPServer(sellerHandler)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start servers
	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Starting HTTP server on port %s", cfg.Server.HTTPPort)
		if err := httpServer.Start(cfg.Server.HTTPPort); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	go func() {
		<-sigChan
		log.Println("Shutdown signal received")
		cancel()
	}()

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Shutting down...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	} else {
		log.Println("HTTP server shut down gracefully")
	}

	// Wait for all goroutines to finish
	wg.Wait()
	log.Println("Seller service shut down complete")
}

// createSellersTable creates the sellers table if it doesn't exist
func createSellersTable(pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS sellers (
			id UUID PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(254) NOT NULL UNIQUE,
			phone VARCHAR(20) NOT NULL,
			address VARCHAR(255) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_sellers_email ON sellers(email);
		CREATE INDEX IF NOT EXISTS idx_sellers_is_active ON sellers(is_active);
		CREATE INDEX IF NOT EXISTS idx_sellers_created_at ON sellers(created_at);
	`

	_, err := pool.Exec(context.Background(), query)
	return err
}
