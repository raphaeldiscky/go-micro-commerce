// Package main implements the entry point for the auth service.
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

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/infra/db/postgres"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/server"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := postgres.NewConnection(cfg.Postgres)
	if err != nil {
		db.Close()
		log.Fatalf("Failed to ping database: %v", err)
	}
	defer db.Close()

	// Initialize logger (0 = Info level)
	appLogger := logger.NewZapLogger(0)

	// Initialize repositories
	userRepo := postgres.NewUserRepositoryPostgres(db)
	sessionRepo := postgres.NewSessionRepository(db)

	// Initialize event publisher
	eventPublisher := event.NewSimplePublisher(cfg.EventPublisher)

	// Initialize service
	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		cfg.JWT,
		eventPublisher,
		appLogger,
	)

	// Initialize handler
	authHandler := handler.NewAuthHandler(authService, appLogger)

	// Initialize HTTP server
	httpServer, err := server.NewHTTPServer(authHandler, cfg, appLogger)

	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}

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
