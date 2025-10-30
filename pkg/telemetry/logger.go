package telemetry

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// WithTraceContext adds trace ID and span ID to the logger from context.
// This method enriches logs with distributed tracing context for correlation.
func (t *Telemetry) WithTraceContext(ctx context.Context, logger logger.Logger) logger.Logger {
	if !t.tracingEnabled {
		return logger
	}

	traceID := t.GetTraceID(ctx)
	spanID := t.GetSpanID(ctx)

	fields := make(map[string]any)
	if traceID != "" {
		fields["trace_id"] = traceID
	}

	if spanID != "" {
		fields["span_id"] = spanID
	}

	if len(fields) == 0 {
		return logger
	}

	return logger.WithFields(fields)
}
