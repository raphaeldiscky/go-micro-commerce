// Package telemetry provides shared observability utilities for distributed tracing,
// metrics collection, and structured logging across microservices.
package telemetry

// TracingConfig holds configuration for OpenTelemetry tracing.
type TracingConfig struct {
	Enabled       bool    `mapstructure:"TRACING_ENABLED"`
	URL           string  `mapstructure:"TRACING_URL"`
	ServiceName   string  `mapstructure:"TRACING_SERVICE_NAME"`
	SamplingRate  float64 `mapstructure:"TRACING_SAMPLING_RATE"`
	Environment   string  `mapstructure:"TRACING_ENVIRONMENT"`
	BatchTimeout  int     `mapstructure:"TRACING_BATCH_TIMEOUT"`
	ExportTimeout int     `mapstructure:"TRACING_EXPORT_TIMEOUT"`
}

// MetricsConfig holds configuration for Prometheus metrics.
type MetricsConfig struct {
	Enabled bool   `mapstructure:"METRICS_ENABLED"`
	Path    string `mapstructure:"METRICS_PATH"`
}
