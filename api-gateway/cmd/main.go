// Package main implements the API Gateway for the microservices architecture.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/metrics"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/ratelimit"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/tracing"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/monitoring"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/routes"
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

	// Initialize metrics
	metricsInstance := metrics.NewMetrics()

	// Initialize services
	discoveryService, err := service.NewConsulServiceDiscovery(cfg.ServiceDiscovery)
	if err != nil {
		log.Fatal("Failed to initialize service discovery", err)
	}

	// Initialize circuit breaker
	circuitBreaker := service.NewCircuitBreaker()

	// Initialize load balancer
	loadBalancer := service.NewLoadBalancer()

	// Create Echo instance
	e := echo.New()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(tracing.Middleware())
	e.Use(metricsInstance.Middleware())
	e.Use(ratelimit.Middleware(*cfg.RateLimit))

	// Initialize monitoring handler
	monitoringHandler := monitoring.NewHandler(appLogger, "1.0.0")
	monitoringHandler.RegisterRoutes(e)

	// Metrics endpoint (Prometheus format)
	e.GET("/metrics", metrics.Handler())

	// Initialize API Gateway
	gw := gateway.New(gateway.Config{
		Logger:           appLogger,
		ServiceDiscovery: discoveryService,
		CircuitBreaker:   circuitBreaker,
		LoadBalancer:     loadBalancer,
		Metrics:          metricsInstance,
		Config:           cfg,
	})

	// Setup routes
	routes.SetupAPIGatewayRoutes(e, gw)

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

	// Start server
	go func() {
		if err := e.Start(fmt.Sprintf(":%d", cfg.HTTPServer.Port)); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("Failed to start server", err)
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
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	} else {
		log.Println("HTTP server shut down gracefully")
	}

	// Wait for all goroutines to finish
	wg.Wait()
	log.Println("API Gateway shut down complete")
}
