// Package main implements the API Gateway for the microservices architecture.
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

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/metrics"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/server"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration", err)
	}

	// Initialize logger
	appLogger := logger.NewZapLogger(0)

	// Initialize services
	discoveryService, err := service.NewConsulServiceDiscovery(cfg.ServiceDiscovery)
	if err != nil {
		log.Fatal("Failed to initialize service discovery", err)
	}

	// Initialize circuit breaker
	circuitBreaker := service.NewCircuitBreaker()

	// Initialize load balancer
	loadBalancer := service.NewLoadBalancer()

	// Initialize metrics
	metricsInstance := metrics.NewMetrics()

	// Initialize API Gateway
	gw := gateway.New(gateway.Config{
		Logger:           appLogger,
		ServiceDiscovery: discoveryService,
		CircuitBreaker:   circuitBreaker,
		LoadBalancer:     loadBalancer,
		Metrics:          metricsInstance,
		Config:           cfg,
	})

	// Initialize monitoring handler
	monitoringHandler := handler.NewMonitoringHandler(appLogger, "1.0.0")

	// Initialize HTTP server
	httpServer := server.NewHTTPServer(gw, metricsInstance, cfg, appLogger, monitoringHandler)

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
	log.Println("API Gateway shut down complete")
}
