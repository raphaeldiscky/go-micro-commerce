// Package tracing provides OpenTelemetry tracing functionality for the API gateway.
package tracing

import (
	"context"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
)

// InitTracing initializes OpenTelemetry tracing.
func InitTracing(cfg *config.TracingConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// Create OTLP HTTP exporter for Tempo
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(cfg.URL),
		otlptracehttp.WithInsecure(), // Use insecure for local development
	)
	if err != nil {
		return err
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return err
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRate)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

// Middleware creates OpenTelemetry tracing middleware.
func Middleware() echo.MiddlewareFunc {
	return otelecho.Middleware("api-gateway")
}

// StartSpan starts a new span with the given name.
func StartSpan(ctx context.Context, name string) (spanCtx context.Context, endFunc func()) {
	tracer := otel.Tracer("api-gateway")
	spanCtx, span := tracer.Start(ctx, name)

	return spanCtx, func() {
		span.End()
	}
}

// AddSpanAttributes adds attributes to the current span.
func AddSpanAttributes(ctx context.Context, attributes map[string]interface{}) {
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

// SetSpanError sets an error on the current span.
func SetSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// GetTraceID returns the trace ID from the context.
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ""
	}

	return span.SpanContext().TraceID().String()
}

// GetSpanID returns the span ID from the context.
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ""
	}

	return span.SpanContext().SpanID().String()
}

// InjectHeaders injects tracing headers into HTTP headers.
func InjectHeaders(ctx context.Context, headers map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, &headerCarrier{headers: headers})
}

// ExtractContext extracts tracing context from HTTP headers.
func ExtractContext(ctx context.Context, headers map[string][]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, &headerExtractor{headers: headers})
}

// headerCarrier implements TextMapCarrier for injecting headers.
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

// headerExtractor implements TextMapCarrier for extracting headers.
type headerExtractor struct {
	headers map[string][]string
}

func (h *headerExtractor) Get(key string) string {
	values := h.headers[key]
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func (h *headerExtractor) Set(key, value string) {
	h.headers[key] = []string{value}
}

func (h *headerExtractor) Keys() []string {
	keys := make([]string, 0, len(h.headers))
	for k := range h.headers {
		keys = append(keys, k)
	}

	return keys
}
