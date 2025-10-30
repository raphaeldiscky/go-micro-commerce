package telemetry

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// WithTraceContext adds trace ID and span ID to the logger from context.
func WithTraceContext(ctx context.Context, log logger.Logger) logger.Logger {
	traceID := GetTraceID(ctx)
	spanID := GetSpanID(ctx)

	fields := make(map[string]any)
	if traceID != "" {
		fields["trace_id"] = traceID
	}

	if spanID != "" {
		fields["span_id"] = spanID
	}

	if len(fields) == 0 {
		return log
	}

	return log.WithFields(fields)
}
