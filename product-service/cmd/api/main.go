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
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/cmd/worker"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.App.LoggerLevel)

	// Initialize telemetry
	telemetryCleanup := setupTelemetry(cfg, appLogger)
	defer telemetryCleanup()

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

// setupTelemetry initializes OpenTelemetry tracing and Prometheus metrics.
func setupTelemetry(cfg *config.Config, appLogger logger.Logger) func() {
	// Initialize tracing
	if err := telemetry.InitTracing(&telemetry.TracingConfig{
		Enabled:       cfg.Tracing.Enabled,
		URL:           cfg.Tracing.URL,
		ServiceName:   cfg.Tracing.ServiceName,
		SamplingRate:  cfg.Tracing.SamplingRate,
		Environment:   cfg.Tracing.Environment,
		BatchTimeout:  cfg.Tracing.BatchTimeout,
		ExportTimeout: cfg.Tracing.ExportTimeout,
	}); err != nil {
		appLogger.Errorf("Failed to initialize tracing: %v", err)
	} else {
		appLogger.Info("OpenTelemetry tracing initialized successfully")
	}

	// Initialize metrics
	if cfg.Metrics.Enabled {
		telemetry.InitMetrics(cfg.App.Name)
		appLogger.Info("Prometheus metrics initialized successfully")
	}

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), constant.MetricTimeout)
		defer cancel()

		if err := telemetry.ShutdownTracing(ctx); err != nil {
			appLogger.Errorf("Failed to shutdown tracing: %v", err)
		} else {
			appLogger.Info("Tracing shutdown successfully")
		}
	}
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

	if err = consulClient.RegisterConnectRPC(cfg.GRPCServer.ServiceName, cfg.GRPCServer.Host, cfg.GRPCServer.Port); err != nil {
		appLogger.Errorf("Failed to register Connect-RPC service with Consul: %v", err)

		return func() {}
	}

	return func() {
		if err = consulClient.Deregister(); err != nil {
			appLogger.Errorf("Failed to deregister services from Consul: %v", err)
		}
	}
}
