package constant

import "time"

const (
	// TracingSamplingRate is the sampling rate for tracing.
	TracingSamplingRate = 1.0
	// TracingBatchTimeout is the batch timeout for tracing.
	TracingBatchTimeout = 5 * time.Second
	// TracingExportTimeout is the export timeout for tracing.
	TracingExportTimeout = 5 * time.Second
)
