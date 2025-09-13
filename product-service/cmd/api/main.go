// Package main implements the API for the product service.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/consul"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/cmd/worker"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.App.LoggerLevel)

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	consulCleanup := setupConsulRegistration(cfg, appLogger)
	defer consulCleanup()

	if err = worker.Start(ctx, cfg, appLogger); err != nil {
		appLogger.Fatalf("Worker failed to start: %v", err)
	}

	appLogger.Info("Application shutdown completed")
}

// setupConsulRegistration handles Consul service registration and returns a cleanup function.
func setupConsulRegistration(cfg *config.Config, appLogger logger.Logger) func() {
	if !cfg.Consul.Enabled {
		appLogger.Info("Consul service discovery is disabled")
		return func() {}
	}

	consulClient, err := consul.NewServiceRegistration(cfg.Consul.Address, appLogger)
	if err != nil {
		return func() {}
	}

	if err = consulClient.RegisterHTTP(cfg.App.Name, cfg.HTTPServer.Host, cfg.HTTPServer.Port); err != nil {
		appLogger.Errorf("Failed to register HTTP service with Consul: %v", err)

		return func() {}
	}

	if err = consulClient.RegisterGRPC(cfg.GRPCServer.ServiceName, cfg.GRPCServer.Host, cfg.GRPCServer.Port); err != nil {
		appLogger.Errorf("Failed to register gRPC service with Consul: %v", err)

		return func() {}
	}

	return func() {
		if err = consulClient.Deregister(); err != nil {
			appLogger.Errorf("Failed to deregister services from Consul: %v", err)
		}
	}
}
