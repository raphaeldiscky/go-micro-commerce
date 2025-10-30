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
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/cmd/worker"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/middleware/metrics"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.App.LoggerLevel)

	// Initialize telemetry
	telemetryCleanup := setupTelemetry(cfg, appLogger)
	defer telemetryCleanup()

	consulCleanup := setupConsulRegistration(cfg, appLogger)
	defer consulCleanup()

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	discoveryService := service.NewConsulDiscoveryService(cfg.ServiceDiscovery, appLogger)
	metricsInstance := metrics.NewMetrics()
	circuitBreaker := service.NewCircuitBreakerService(appLogger, cfg, metricsInstance)

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
