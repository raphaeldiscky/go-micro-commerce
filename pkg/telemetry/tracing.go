package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	defaultBatchTimeout  = 5 * time.Second
	defaultExportTimeout = 30 * time.Second
)

// initTracing initializes OpenTelemetry tracing with the provided configuration.
func (t *Telemetry) initTracing(cfg Config) error {
	// Create OTLP HTTP exporter for Tempo
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(cfg.TracingURL),
		otlptracehttp.WithInsecure(), // Use insecure for local development
	)
	if err != nil {
		return err
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.TracingServiceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(cfg.TracingEnvironment),
		),
	)
	if err != nil {
		return err
	}

	// Create trace provider with configurable batch timeout
	batchTimeout := time.Duration(cfg.TracingBatchTimeout) * time.Second
	if batchTimeout == 0 {
		batchTimeout = defaultBatchTimeout
	}

	exportTimeout := time.Duration(cfg.TracingExportTimeout) * time.Second
	if exportTimeout == 0 {
		exportTimeout = defaultExportTimeout
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(batchTimeout),
			sdktrace.WithExportTimeout(exportTimeout),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.TracingSamplingRate)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)
	t.tracerProvider = tp

	// Set global propagator for W3C trace context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

// Shutdown gracefully shuts down the tracer provider.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.tracerProvider == nil {
		return nil
	}

	return t.tracerProvider.Shutdown(ctx)
}

// StartSpan starts a new span with the given name.
func (t *Telemetry) StartSpan(ctx context.Context, spanName string) (context.Context, func()) {
	if !t.tracingEnabled {
		return ctx, func() {}
	}

	tracer := otel.Tracer(t.serviceName)
	spanCtx, span := tracer.Start(ctx, spanName)

	return spanCtx, func() {
		span.End()
	}
}

// AddSpanAttributes adds attributes to the current span in the context.
func (t *Telemetry) AddSpanAttributes(ctx context.Context, attributes map[string]any) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	for key, value := range attributes {
		switch v := value.(type) {
		case string:
			span.SetAttributes(attribute.String(key, v))
		case int:
			span.SetAttributes(attribute.Int(key, v))
		case int64:
			span.SetAttributes(attribute.Int64(key, v))
		case float64:
			span.SetAttributes(attribute.Float64(key, v))
		case bool:
			span.SetAttributes(attribute.Bool(key, v))
		}
	}
}

// SetSpanError records an error on the current span and sets its status to Error.
func (t *Telemetry) SetSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// GetTraceID extracts the trace ID from the current span context.
func (t *Telemetry) GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ""
	}

	return span.SpanContext().TraceID().String()
}

// GetSpanID extracts the span ID from the current span context.
func (t *Telemetry) GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ""
	}

	return span.SpanContext().SpanID().String()
}

// InjectHeaders injects tracing context into HTTP headers for propagation.
func (t *Telemetry) InjectHeaders(ctx context.Context, headers map[string]string) {
	if !t.tracingEnabled {
		return
	}

	otel.GetTextMapPropagator().Inject(ctx, &headerCarrier{headers: headers})
}

// ExtractContext extracts tracing context from HTTP headers.
func (t *Telemetry) ExtractContext(ctx context.Context, headers map[string]string) context.Context {
	if !t.tracingEnabled {
		return ctx
	}

	return otel.GetTextMapPropagator().Extract(ctx, &headerCarrier{headers: headers})
}

// headerCarrier implements TextMapCarrier for injecting/extracting trace context.
type headerCarrier struct {
	headers map[string]string
}

func (h *headerCarrier) Get(key string) string {
	return h.headers[key]
}

func (h *headerCarrier) Set(key, value string) {
	h.headers[key] = value
}

func (h *headerCarrier) Keys() []string {
	keys := make([]string, 0, len(h.headers))
	for k := range h.headers {
		keys = append(keys, k)
	}

	return keys
}
