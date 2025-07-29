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

	services "github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/app/service"
	"github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/config"
	"github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/infra/db/postgres"
	"github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/infra/kafka"
	handlers "github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/interface/http/handler"
	"github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/interface/http/server"
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
	productRepo := postgres.NewProductRepositoryPostgres(dbPool)

	// Setup services
	productService := services.NewProductService(productRepo, eventPublisher)

	// Setup HTTP handlers
	productHandler := handlers.NewProductHandler(productService)

	// Setup HTTP server
	httpServer := server.NewHTTPServer(productHandler)

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
	log.Println("Product service shut down complete")
}
