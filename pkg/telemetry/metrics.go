package telemetry

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// initMetrics initializes Prometheus metrics for HTTP requests.
// Uses sync.Once to ensure metrics are only registered once with the global registry.
func (t *Telemetry) initMetrics() {
	t.once.Do(func() {
		t.httpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
				ConstLabels: prometheus.Labels{
					"service": t.serviceName,
				},
			},
			[]string{"method", "path", "status_code"},
		)

		t.httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_request_duration_seconds",
				Help: "HTTP request duration in seconds",
				ConstLabels: prometheus.Labels{
					"service": t.serviceName,
				},
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status_code"},
		)

		// Register standard metrics with global Prometheus registry
		prometheus.MustRegister(
			t.httpRequestsTotal,
			t.httpRequestDuration,
		)

		// Initialize gateway-specific metrics if this is the api-gateway
		if t.serviceName == "api-gateway" {
			t.backendRequestsTotal = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "api_gateway_backend_requests_total",
					Help: "Total number of requests to backend services",
				},
				[]string{"service", "method", "status_code"},
			)

			t.backendRequestDuration = prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "api_gateway_backend_request_duration_seconds",
					Help:    "Backend service request duration in seconds",
					Buckets: prometheus.DefBuckets,
				},
				[]string{"service", "method"},
			)

			t.circuitBreakerState = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "api_gateway_circuit_breaker_state",
					Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
				},
				[]string{"service"},
			)

			// Register gateway metrics
			prometheus.MustRegister(
				t.backendRequestsTotal,
				t.backendRequestDuration,
				t.circuitBreakerState,
			)
		}
	})
}

// MetricsHandler returns an Echo handler for Prometheus metrics endpoint.
func (t *Telemetry) MetricsHandler() echo.HandlerFunc {
	handler := promhttp.Handler()

	return func(c echo.Context) error {
		handler.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

// RecordMetric is a helper function to record custom metrics.
func (t *Telemetry) RecordMetric(method, path string, statusCode int, duration time.Duration) {
	if !t.metricsEnabled {
		return
	}

	if t.httpRequestsTotal != nil {
		t.httpRequestsTotal.WithLabelValues(
			method,
			path,
			strconv.Itoa(statusCode),
		).Inc()
	}

	if t.httpRequestDuration != nil {
		t.httpRequestDuration.WithLabelValues(
			method,
			path,
			strconv.Itoa(statusCode),
		).Observe(duration.Seconds())
	}
}

// RecordBackendRequest records backend service request metrics (API Gateway only).
func (t *Telemetry) RecordBackendRequest(
	service, method string,
	statusCode int,
	duration time.Duration,
) {
	if !t.metricsEnabled || t.backendRequestsTotal == nil {
		return
	}

	t.backendRequestsTotal.WithLabelValues(service, method, strconv.Itoa(statusCode)).Inc()
	t.backendRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// SetCircuitBreakerState sets the circuit breaker state for a service (API Gateway only).
// state: 0=closed, 1=half-open, 2=open
func (t *Telemetry) SetCircuitBreakerState(service string, state float64) {
	if !t.metricsEnabled || t.circuitBreakerState == nil {
		return
	}

	t.circuitBreakerState.WithLabelValues(service).Set(state)
}
