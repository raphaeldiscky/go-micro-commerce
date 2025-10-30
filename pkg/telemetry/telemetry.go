// Package telemetry provides shared observability utilities for distributed tracing,
// metrics collection, and structured logging across microservices.
package telemetry

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Telemetry manages distributed tracing and metrics collection for a service.
// It wraps OpenTelemetry and Prometheus in a dependency-injectable struct.
type Telemetry struct {
	serviceName    string
	tracerProvider *sdktrace.TracerProvider
	tracingEnabled bool
	metricsEnabled bool

	// Prometheus metrics - must be package-level due to global registry constraint
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec

	// Singleton enforcement
	once sync.Once
}

// NewTelemetry creates a new Telemetry instance with the provided configuration.
// This function initializes OpenTelemetry tracing and Prometheus metrics based on the config.
func NewTelemetry(cfg Config) (*Telemetry, error) {
	tel := &Telemetry{
		serviceName:    cfg.TracingServiceName,
		tracingEnabled: cfg.TracingEnabled,
		metricsEnabled: cfg.MetricsEnabled,
	}

	// Initialize tracing if enabled
	if cfg.TracingEnabled {
		if err := tel.initTracing(cfg); err != nil {
			return nil, err
		}
	}

	// Initialize metrics if enabled
	if cfg.MetricsEnabled {
		tel.initMetrics()
	}

	return tel, nil
}
