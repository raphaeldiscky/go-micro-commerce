package telemetry

import (
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

		// Register metrics with global Prometheus registry
		prometheus.MustRegister(
			t.httpRequestsTotal,
			t.httpRequestDuration,
		)
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
			string(rune(statusCode)),
		).Inc()
	}

	if t.httpRequestDuration != nil {
		t.httpRequestDuration.WithLabelValues(
			method,
			path,
			string(rune(statusCode)),
		).Observe(duration.Seconds())
	}
}
