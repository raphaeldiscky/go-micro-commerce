// Package main implements the API for the product service.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/raphaeldiscky/go-micro-template/pkg/consul"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/order-service/cmd/worker"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.Logger.Level)

	// Create main context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	// Setup Consul registration
	consulCleanup := setupConsulRegistration(cfg, appLogger)
	defer consulCleanup()

	// Start worker with graceful shutdown
	if err := worker.Start(ctx, cfg, appLogger); err != nil {
		appLogger.Fatalf("Worker failed to start: %v", err)
	}

	appLogger.Info("Application shutdown completed")
}

// setupConsulRegistration handles Consul service registration and returns a cleanup function.
func setupConsulRegistration(cfg *config.Config, appLogger logger.Logger) func() {
	if !cfg.Consul.Enabled {
		appLogger.Infof("Consul service discovery is disabled")

		return func() {}
	}

	consulClient, err := consul.NewServiceRegistration(cfg.Consul.Address)
	if err != nil {
		appLogger.Errorf("Failed to create Consul client: %v", err)

		return func() {}
	}

	if err := consulClient.RegisterHTTP(cfg.Consul.ServiceName, cfg.Consul.ServiceHost, cfg.HTTPServer.Port); err != nil {
		appLogger.Errorf("Failed to register with Consul: %v", err)

		return func() {}
	}

	appLogger.Infof("Service registered with Consul: %s at %s:%d",
		cfg.Consul.ServiceName, cfg.Consul.ServiceHost, cfg.HTTPServer.Port)

	return func() {
		if err := consulClient.Deregister(); err != nil {
			appLogger.Errorf("Failed to deregister from Consul: %v", err)
		} else {
			appLogger.Infof("Service deregistered from Consul")
		}
	}
}
