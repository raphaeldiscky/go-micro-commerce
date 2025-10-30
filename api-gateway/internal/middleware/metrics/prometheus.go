// Package metrics provides Prometheus metrics for the API gateway.
package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds all Prometheus metrics for the gateway.
type Metrics struct {
	// Gateway backend service metrics.
	backendRequestsTotal   *prometheus.CounterVec
	backendRequestDuration *prometheus.HistogramVec
	circuitBreakerState    *prometheus.GaugeVec
}

// NewMetrics creates and initializes all Prometheus metrics.
func NewMetrics() *Metrics {
	m := &Metrics{
		backendRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_gateway_backend_requests_total",
				Help: "Total number of requests to backend services",
			},
			[]string{"service", "method", "status_code"},
		),

		backendRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_gateway_backend_request_duration_seconds",
				Help:    "Backend service request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method"},
		),

		circuitBreakerState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "api_gateway_circuit_breaker_state",
				Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
			},
			[]string{"service"},
		),
	}

	// Register metrics with Prometheus
	prometheus.MustRegister(
		m.backendRequestsTotal,
		m.backendRequestDuration,
		m.circuitBreakerState,
	)

	return m
}

// RecordGatewayRequest records backend service request metrics.
func (m *Metrics) RecordGatewayRequest(
	service, method string,
	statusCode int,
	duration time.Duration,
) {
	m.backendRequestsTotal.WithLabelValues(service, method, strconv.Itoa(statusCode)).Inc()
	m.backendRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// SetCircuitBreakerState sets the circuit breaker state for a service.
// state: 0=closed, 1=half-open, 2=open
func (m *Metrics) SetCircuitBreakerState(service string, state float64) {
	m.circuitBreakerState.WithLabelValues(service).Set(state)
}
