// Package dto provides data transfer objects for the API gateway.
package dto

import "time"

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks"`
}

// MetricsResponse represents basic application metrics.
type MetricsResponse struct {
	Timestamp      time.Time `json:"timestamp"`
	Uptime         string    `json:"uptime"`
	MemoryUsage    uint64    `json:"memory_usage_bytes"`
	GoroutineCount int       `json:"goroutine_count"`
	RequestCount   int64     `json:"request_count,omitempty"`
	ErrorCount     int64     `json:"error_count,omitempty"`
}
