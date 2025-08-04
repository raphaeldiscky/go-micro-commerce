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

// Middleware creates a Prometheus metrics middleware.
func (m *Metrics) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Increment active connections
			m.activeConnections.Inc()
			defer m.activeConnections.Dec()

			// Execute the request
			err := next(c)

			// Record metrics
			duration := time.Since(start).Seconds()
			method := c.Request().Method
			path := c.Path()
			statusCode := strconv.Itoa(c.Response().Status)

			m.httpRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
			m.httpRequestDuration.WithLabelValues(method, path, statusCode).Observe(duration)

			return err
		}
	}
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

// SetCircuitBreakerState sets the circuit breaker state metric.
func (m *Metrics) SetCircuitBreakerState(service string, state int) {
	m.circuitBreakerState.WithLabelValues(service).Set(float64(state))
}

// Handler returns the Prometheus metrics handler.
func Handler() echo.HandlerFunc {
	handler := promhttp.Handler()

	return func(c echo.Context) error {
		handler.ServeHTTP(c.Response(), c.Request())

		return nil
	}
}

// CustomMetrics provides access to custom metrics.
type CustomMetrics struct {
	ServiceRequestsTotal   *prometheus.CounterVec
	ServiceRequestDuration *prometheus.HistogramVec
	ServiceErrors          *prometheus.CounterVec
}

// NewCustomMetrics creates custom metrics for services.
func NewCustomMetrics() *CustomMetrics {
	serviceRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_requests_total",
			Help: "Total number of service requests",
		},
		[]string{"service", "endpoint", "status"},
	)

	serviceRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "service_request_duration_seconds",
			Help:    "Service request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "endpoint"},
	)

	serviceErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_errors_total",
			Help: "Total number of service errors",
		},
		[]string{"service", "error_type"},
	)

	prometheus.MustRegister(serviceRequestsTotal, serviceRequestDuration, serviceErrors)

	return &CustomMetrics{
		ServiceRequestsTotal:   serviceRequestsTotal,
		ServiceRequestDuration: serviceRequestDuration,
		ServiceErrors:          serviceErrors,
	}
}
