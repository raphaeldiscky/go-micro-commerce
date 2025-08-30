// Package main implements the API for the api gateway
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/cmd/worker"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/middleware/metrics"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.Logger.Level)

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	discoveryService := service.NewConsulDiscoveryService(cfg.ServiceDiscovery, appLogger)
	circuitBreaker := service.NewCircuitBreakerService(appLogger)
	loadBalancer := service.NewLoadBalancerService(appLogger)
	metricsInstance := metrics.NewMetrics()

	// Initialize API Gateway
	gw := gateway.NewAPIGateway(gateway.Config{
		Logger:           appLogger,
		ServiceDiscovery: discoveryService,
		CircuitBreaker:   circuitBreaker,
		LoadBalancer:     loadBalancer,
		Metrics:          metricsInstance,
		Config:           cfg,
	})

	if err := worker.Start(ctx, cfg, gw, appLogger); err != nil {
		appLogger.Fatalf("Worker failed to start: %v", err)
	}

	appLogger.Info("Application shutdown completed")
}
