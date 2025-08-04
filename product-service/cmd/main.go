// Package main initializes and runs the product service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/services/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/services/product-service/internal/infra/kafka"
	handlers "github.com/raphaeldiscky/go-micro-template/services/product-service/internal/interface/http/handler"
	"github.com/raphaeldiscky/go-micro-template/services/product-service/internal/interface/http/server"
	"github.com/raphaeldiscky/go-micro-template/services/product-service/internal/repository"
	services "github.com/raphaeldiscky/go-micro-template/services/product-service/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup database connection
	dbPool, err := pgxpool.New(context.Background(), cfg.GetURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		dbPool.Close()
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Database is ready, set up defer for cleanup
	defer dbPool.Close()

	log.Println("Database connection established")

	// Setup logger
	appLogger := logger.NewZapLogger(0) // 0 = debug level

	// Setup Kafka event publisher
	eventPublisher, err := kafka.NewEventPublisherKafka(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		log.Printf("Warning: Failed to setup Kafka event publisher: %v", err)
		log.Println("Continuing without event publishing...")

		eventPublisher = nil
	} else {
		defer func() {
			if err := eventPublisher.Close(); err != nil {
				log.Printf("Error closing Kafka event publisher: %v", err)
			}
		}()
		log.Println("Kafka event publisher initialized")
	}

	// Setup repository
	productRepo := repository.NewProductRepository(dbPool)

	// Setup services
	productService := services.NewProductService(productRepo, eventPublisher, appLogger)

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
		log.Printf("Starting HTTP server on port %d", cfg.HTTPServer.Port)

		portStr := fmt.Sprintf("%d", cfg.HTTPServer.Port)
		if err := httpServer.Start(portStr); err != nil {
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
