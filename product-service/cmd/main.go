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

	"github.com/raphaeldiscky/go-micro-template/pkg/db"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/consul"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/infra/db/postgres"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/server"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pgPool, err := db.NewPostgresConnection(&db.PostgresConfig{
		Host:            cfg.Postgres.Host,
		Port:            cfg.Postgres.Port,
		User:            cfg.Postgres.User,
		Password:        cfg.Postgres.Password,
		Name:            cfg.Postgres.Name,
		SSLMode:         cfg.Postgres.SSLMode,
		MaxIdleConns:    cfg.Postgres.MaxIdleConns,
		MaxOpenConns:    cfg.Postgres.MaxOpenConns,
		MaxConnLifetime: cfg.Postgres.MaxConnLifetime,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pgPool.Close()

	// Setup logger
	appLogger := logger.NewZapLogger(0) // 0 = debug level

	// Setup Kafka event publisher
	producer, err := mq.NewProducerKafka(&mq.KafkaProducerConfig{
		Brokers:        cfg.Kafka.Brokers,
		ReturnSuccess:  cfg.Kafka.ReturnSuccess,
		ReturnErrors:   cfg.Kafka.ReturnErrors,
		RetryMax:       cfg.Kafka.RetryMax,
		FlushFrequency: cfg.Kafka.FlushFrequency,
	})
	if err != nil {
		log.Printf("Warning: Failed to setup Kafka event publisher: %v", err)
		log.Println("Continuing without event publishing...")

		producer = nil
	} else {
		defer func() {
			if err := producer.Close(); err != nil {
				log.Printf("Error closing Kafka event producer: %v", err)
			}
		}()
		log.Println("Kafka event publisher initialized")
	}

	// Setup repository
	productRepo := postgres.NewProductRepositoryPostgres(pgPool)

	// Setup services
	topics := constant.NewProductTopics()
	productService := service.NewProductService(productRepo, producer, topics, appLogger)

	// Setup HTTP handlers
	productHandler := handler.NewProductHandler(productService)

	// Initialize HTTP server
	httpServer := server.NewHTTPServer(productHandler, cfg.HTTPServer, appLogger)

	// Register with Consul if enabled and setup cleanup
	consulCleanup := setupConsulRegistration(cfg)
	defer consulCleanup()

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

// setupConsulRegistration handles Consul service registration and returns a cleanup function.
func setupConsulRegistration(cfg *config.Config) func() {
	if !cfg.Consul.Enabled {
		log.Println("Consul service discovery is disabled")

		return func() {}
	}

	consulClient, err := consul.NewServiceRegistration(cfg.Consul.Address)
	if err != nil {
		log.Printf("Failed to create Consul client: %v", err)

		return func() {}
	}

	if err := consulClient.Register(cfg.Consul.ServiceName, cfg.Consul.ServiceHost, cfg.HTTPServer.Port); err != nil {
		log.Printf("Failed to register with Consul: %v", err)

		return func() {}
	}

	log.Printf("Service registered with Consul: %s at %s:%d",
		cfg.Consul.ServiceName, cfg.Consul.ServiceHost, cfg.HTTPServer.Port)

	// Return cleanup function
	return func() {
		if err := consulClient.Deregister(); err != nil {
			log.Printf("Failed to deregister from Consul: %v", err)
		} else {
			log.Println("Service deregistered from Consul")
		}
	}
}
