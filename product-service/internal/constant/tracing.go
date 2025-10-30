package constant

import "time"

const (
	// TracingEnabled indicates if tracing is enabled.
	TracingEnabled = true
	// TracingURL is the URL of the OpenTelemetry collector.
	TracingURL = "localhost:4318"
	// TracingServiceName is the name of the service for tracing.
	TracingServiceName = "product-service"
	// TracingSamplingRate is the sampling rate for tracing (0.0 to 1.0).
	TracingSamplingRate = 0.1
	// TracingEnvironment is the environment for tracing.
	TracingEnvironment = "development"
	// TracingBatchTimeout is the batch timeout in seconds for tracing.
	TracingBatchTimeout = 5 * time.Second
	// TracingExportTimeout is the export timeout in seconds for tracing.
	TracingExportTimeout = 30 * time.Second
)
