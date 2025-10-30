package telemetry

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	//nolint:gochecknoglobals // Prometheus metrics must be package-level to register with global registry
	httpRequestsTotal *prometheus.CounterVec

	//nolint:gochecknoglobals // Prometheus metrics must be package-level to register with global registry
	httpRequestDuration *prometheus.HistogramVec
)

// InitMetrics initializes Prometheus metrics for HTTP requests.
func InitMetrics(serviceName string) {
	if httpRequestsTotal != nil {
		return
	}

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
			ConstLabels: prometheus.Labels{
				"service": serviceName,
			},
		},
		[]string{"method", "path", "status_code"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP request duration in seconds",
			ConstLabels: prometheus.Labels{
				"service": serviceName,
			},
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status_code"},
	)

	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
	)
}

// MetricsMiddleware creates an Echo middleware that records HTTP metrics.
func MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status
			method := c.Request().Method
			path := c.Path()

			if httpRequestsTotal != nil {
				httpRequestsTotal.WithLabelValues(
					method,
					path,
					strconv.Itoa(status),
				).Inc()
			}

			if httpRequestDuration != nil {
				httpRequestDuration.WithLabelValues(
					method,
					path,
					strconv.Itoa(status),
				).Observe(duration.Seconds())
			}

			return err
		}
	}
}

// MetricsHandler returns an Echo handler for Prometheus metrics endpoint.
func MetricsHandler() echo.HandlerFunc {
	handler := promhttp.Handler()

	return func(c echo.Context) error {
		handler.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

// RecordMetric is a helper function to record custom metrics.
func RecordMetric(method, path string, statusCode int, duration time.Duration) {
	if httpRequestsTotal != nil {
		httpRequestsTotal.WithLabelValues(
			method,
			path,
			strconv.Itoa(statusCode),
		).Inc()
	}

	if httpRequestDuration != nil {
		httpRequestDuration.WithLabelValues(
			method,
			path,
			strconv.Itoa(statusCode),
		).Observe(duration.Seconds())
	}
}
