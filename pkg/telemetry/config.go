// Package telemetry provides shared observability utilities for distributed tracing,
// metrics collection, and structured logging across microservices.
package telemetry

// Config holds configuration for telemetry (tracing and metrics).
type Config struct {
	// Tracing configuration
	TracingEnabled       bool    `mapstructure:"TRACING_ENABLED"`
	TracingURL           string  `mapstructure:"TRACING_URL"`
	TracingServiceName   string  `mapstructure:"TRACING_SERVICE_NAME"`
	TracingSamplingRate  float64 `mapstructure:"TRACING_SAMPLING_RATE"`
	TracingEnvironment   string  `mapstructure:"TRACING_ENVIRONMENT"`
	TracingBatchTimeout  int     `mapstructure:"TRACING_BATCH_TIMEOUT"`
	TracingExportTimeout int     `mapstructure:"TRACING_EXPORT_TIMEOUT"`

	// Metrics configuration
	MetricsEnabled bool   `mapstructure:"METRICS_ENABLED"`
	MetricsPath    string `mapstructure:"METRICS_PATH"`
}
