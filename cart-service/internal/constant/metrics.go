package constant

import "time"

const (
	// MetricsEnabled indicates if metrics are enabled.
	MetricsEnabled = true
	// MetricsPath is the path for the metrics endpoint.
	MetricsPath = "/metrics"
	// MetricTimeout is the timeout for the metrics server.
	MetricTimeout = 5 * time.Second
)
