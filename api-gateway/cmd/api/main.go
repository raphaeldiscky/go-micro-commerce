// Package main implements the API for the product service.
package main

import (
	"log"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/cmd/worker"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/metrics"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.Logger.Level)

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

	worker.Start(cfg, appLogger, gw)
}
