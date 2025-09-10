// Package metrics provides Prometheus metrics for the API gateway.
package metrics

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the gateway.
type Metrics struct {
	// HTTP request metrics.
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec

	// Gateway specific metrics.
	gatewayRequestsTotal   *prometheus.CounterVec
	gatewayRequestDuration *prometheus.HistogramVec
	circuitBreakerState    *prometheus.GaugeVec
	activeConnections      prometheus.Gauge
}

// NewMetrics creates and initializes all Prometheus metrics.
func NewMetrics() *Metrics {
	m := &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),

		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status_code"},
		),

		gatewayRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_requests_total",
				Help: "Total number of gateway requests",
			},
			[]string{"service", "method", "status_code"},
		),

		gatewayRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_request_duration_seconds",
				Help:    "Gateway request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method"},
		),

		circuitBreakerState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "circuit_breaker_state",
				Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
			},
			[]string{"service"},
		),

		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "gateway_active_connections",
				Help: "Number of active connections",
			},
		),
	}

	// Register metrics with Prometheus
	prometheus.MustRegister(
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.gatewayRequestsTotal,
		m.gatewayRequestDuration,
		m.circuitBreakerState,
		m.activeConnections,
	)

	return m
}

// RecordGatewayRequest records gateway-specific metrics.
func (m *Metrics) RecordGatewayRequest(
	service, method string,
	statusCode int,
	duration time.Duration,
) {
	m.gatewayRequestsTotal.WithLabelValues(service, method, strconv.Itoa(statusCode)).Inc()
	m.gatewayRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// Handler returns the Prometheus metrics handler.
func Handler() echo.HandlerFunc {
	handler := promhttp.Handler()

	return func(c echo.Context) error {
		handler.ServeHTTP(c.Response(), c.Request())

		return nil
	}
}
