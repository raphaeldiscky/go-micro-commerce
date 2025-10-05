// Package main implements the API for the api gateway
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/consul"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/cmd/worker"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/middleware/metrics"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/middleware/tracing"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.App.LoggerLevel)

	consulCleanup := setupConsulRegistration(cfg, appLogger)
	defer consulCleanup()

	if err = tracing.InitTracing(cfg.Tracing); err != nil {
		appLogger.Fatalf("failed to initialize tracing: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	discoveryService := service.NewConsulDiscoveryService(cfg.ServiceDiscovery, appLogger)
	circuitBreaker := service.NewCircuitBreakerService(appLogger, cfg)
	metricsInstance := metrics.NewMetrics()

	// Initialize API Gateway
	gw := gateway.NewAPIGateway(gateway.Config{
		Logger:           appLogger,
		ServiceDiscovery: discoveryService,
		CircuitBreaker:   circuitBreaker,
		Metrics:          metricsInstance,
		Config:           cfg,
	})

	if err = worker.Start(ctx, cfg, gw, appLogger); err != nil {
		appLogger.Fatalf("Worker failed to start: %v", err)
	}

	appLogger.Info("Application shutdown completed")
}

// setupConsulRegistration handles Consul service registration and returns a cleanup function.
func setupConsulRegistration(cfg *config.Config, appLogger logger.Logger) func() {
	if cfg.ServiceDiscovery.Type != "consul" {
		appLogger.Info("Consul service registration is disabled")
		return func() {}
	}

	consulClient, err := consul.NewServiceRegistration(
		cfg.ServiceDiscovery.ConsulAddress,
		appLogger,
	)
	if err != nil {
		appLogger.Errorf("Failed to create Consul client: %v", err)
		return func() {}
	}

	if err = consulClient.RegisterHTTP(cfg.App.Name, cfg.HTTPServer.Host, cfg.HTTPServer.Port); err != nil {
		appLogger.Errorf("Failed to register with Consul: %v", err)
		return func() {}
	}

	appLogger.Infof("Successfully registered %s with Consul", cfg.App.Name)

	return func() {
		if deregErr := consulClient.Deregister(); deregErr != nil {
			appLogger.Errorf("Failed to deregister from Consul: %v", deregErr)
		}
	}
}
