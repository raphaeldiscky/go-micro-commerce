package telemetry

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// EchoMiddleware creates an Echo middleware that traces HTTP requests using OpenTelemetry.
// This method should be used for distributed tracing across services.
func (t *Telemetry) EchoMiddleware() echo.MiddlewareFunc {
	if !t.tracingEnabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	return otelecho.Middleware(t.serviceName)
}

// MetricsMiddleware creates an Echo middleware that records HTTP metrics.
// This method collects request count and duration metrics for Prometheus.
func (t *Telemetry) MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !t.metricsEnabled {
				return next(c)
			}

			start := time.Now()

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status
			method := c.Request().Method
			path := c.Path()

			if t.httpRequestsTotal != nil {
				t.httpRequestsTotal.WithLabelValues(
					method,
					path,
					strconv.Itoa(status),
				).Inc()
			}

			if t.httpRequestDuration != nil {
				t.httpRequestDuration.WithLabelValues(
					method,
					path,
					strconv.Itoa(status),
				).Observe(duration.Seconds())
			}

			return err
		}
	}
}
