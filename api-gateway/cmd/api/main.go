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
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.App.LoggerLevel)

	// Initialize telemetry
	tel, telemetryCleanup := setupTelemetry(cfg, appLogger)
	defer telemetryCleanup()

	consulCleanup := setupConsulRegistration(cfg, appLogger)
	defer consulCleanup()

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	// Initialize service discovery based on configuration type
	var discoveryService service.Discovery

	switch cfg.ServiceDiscovery.Type {
	case "kubernetes", "k8s":
		discoveryService = service.NewKubernetesDiscoveryService(cfg.ServiceDiscovery, appLogger)
		appLogger.Info("Using Kubernetes DNS-based service discovery")
	case "consul":
		discoveryService = service.NewConsulDiscoveryService(cfg.ServiceDiscovery, appLogger)
		appLogger.Info("Using Consul service discovery")
	default:
		appLogger.Fatalf(
			"Unsupported service discovery type: %s. Supported types: kubernetes, consul",
			cfg.ServiceDiscovery.Type,
		)
	}

	circuitBreaker := service.NewCircuitBreakerService(appLogger, cfg, tel)

	// Initialize API Gateway
	gw := gateway.NewAPIGateway(gateway.Config{
		Logger:           appLogger,
		ServiceDiscovery: discoveryService,
		CircuitBreaker:   circuitBreaker,
		Telemetry:        tel,
		Config:           cfg,
	})

	if err = worker.Start(ctx, cfg, gw, appLogger, tel); err != nil {
		appLogger.Fatalf("Worker failed to start: %v", err)
	}

	appLogger.Info("Application shutdown completed")
}

// setupTelemetry initializes OpenTelemetry tracing and Prometheus metrics.
func setupTelemetry(cfg *config.Config, appLogger logger.Logger) (*telemetry.Telemetry, func()) {
	// Create telemetry instance
	tel, err := telemetry.NewTelemetry(telemetry.Config{
		TracingEnabled:       cfg.Tracing.Enabled,
		TracingURL:           cfg.Tracing.URL,
		TracingServiceName:   cfg.Tracing.ServiceName,
		TracingSamplingRate:  cfg.Tracing.SamplingRate,
		TracingEnvironment:   cfg.Tracing.Environment,
		TracingBatchTimeout:  cfg.Tracing.BatchTimeout,
		TracingExportTimeout: cfg.Tracing.ExportTimeout,
		MetricsEnabled:       cfg.Metrics.Enabled,
		MetricsPath:          cfg.Metrics.Path,
	})
	if err != nil {
		appLogger.Errorf("Failed to initialize telemetry: %v", err)
		return nil, func() {}
	}

	appLogger.Info("Telemetry initialized successfully")

	return tel, func() {
		ctx, cancel := context.WithTimeout(context.Background(), constant.MetricTimeout)
		defer cancel()

		if err = tel.Shutdown(ctx); err != nil {
			appLogger.Errorf("Failed to shutdown telemetry: %v", err)
		} else {
			appLogger.Info("Telemetry shutdown successfully")
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
