// Package monitoring provides health check and monitoring endpoints for the API gateway.
package monitoring

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/tracing"
)

// Handler handles monitoring and health check endpoints.
type Handler struct {
	logger    *zap.Logger
	startTime time.Time
	version   string
}

// NewHandler creates a new monitoring handler.
func NewHandler(logger *zap.Logger, version string) *Handler {
	return &Handler{
		logger:    logger,
		startTime: time.Now(),
		version:   version,
	}
}

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

// Health returns the health status of the API gateway.
func (h *Handler) Health(c echo.Context) error {
	uptime := time.Since(h.startTime)

	// Perform basic health checks
	checks := map[string]string{
		"api_gateway": constant.HealthyStatus,
		"memory":      h.checkMemory(),
		"goroutines":  h.checkGoroutines(),
	}

	// Determine overall status
	status := constant.HealthyStatus

	for _, check := range checks {
		if check != constant.HealthyStatus {
			status = constant.DegradedStatus

			break
		}
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		Version:   h.version,
		Checks:    checks,
	}

	httpStatus := http.StatusOK
	if status != constant.HealthyStatus {
		httpStatus = http.StatusServiceUnavailable
	}

	return c.JSON(httpStatus, response)
}

// Ready returns the readiness status of the API gateway.
func (h *Handler) Ready(c echo.Context) error {
	// Check if the service is ready to accept requests
	// Add any specific readiness checks here
	response := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now(),
	}

	return c.JSON(http.StatusOK, response)
}

// Metrics returns basic application metrics.
func (h *Handler) Metrics(c echo.Context) error {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)

	uptime := time.Since(h.startTime)

	response := MetricsResponse{
		Timestamp:      time.Now(),
		Uptime:         uptime.String(),
		MemoryUsage:    m.Alloc,
		GoroutineCount: runtime.NumGoroutine(),
	}

	return c.JSON(http.StatusOK, response)
}

// Info returns general information about the API gateway.
func (h *Handler) Info(c echo.Context) error {
	response := map[string]interface{}{
		"service":    "api-gateway",
		"version":    h.version,
		"timestamp":  time.Now(),
		"uptime":     time.Since(h.startTime).String(),
		"go_version": runtime.Version(),
		"endpoints": map[string]string{
			"health":      "/health",
			"ready":       "/ready",
			"metrics":     "/metrics",     // Prometheus metrics
			"app-metrics": "/app-metrics", // JSON application metrics
			"info":        "/info",
		},
	}

	return c.JSON(http.StatusOK, response)
}

// TestTrace creates a test trace for monitoring validation.
func (h *Handler) TestTrace(c echo.Context) error {
	ctx := c.Request().Context()

	// Create a test span
	spanCtx, endFunc := tracing.StartSpan(ctx, "test-trace-operation")
	defer endFunc()

	// Add some attributes to the span
	tracing.AddSpanAttributes(spanCtx, map[string]interface{}{
		"test.endpoint": "/test/trace",
		"test.user":     "monitoring-test",
		"test.action":   "trace-validation",
	})

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Get trace and span IDs
	traceID := tracing.GetTraceID(spanCtx)
	spanID := tracing.GetSpanID(spanCtx)

	h.logger.Info("Test trace created",
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
	)

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Test trace created successfully",
		"trace_id":  traceID,
		"span_id":   spanID,
		"timestamp": time.Now(),
	}

	return c.JSON(http.StatusOK, response)
}

// TestError creates a test error for monitoring validation.
func (h *Handler) TestError(c echo.Context) error {
	ctx := c.Request().Context()

	// Create a span for error testing
	spanCtx, endFunc := tracing.StartSpan(ctx, "test-error-operation")
	defer endFunc()

	// Create a test error
	testErr := fmt.Errorf("this is a test error for monitoring validation")

	// Record the error in the span
	tracing.SetSpanError(spanCtx, testErr)

	h.logger.Error("Test error created for monitoring",
		zap.Error(testErr),
		zap.String("trace_id", tracing.GetTraceID(spanCtx)),
	)

	response := map[string]interface{}{
		"status":    "error",
		"message":   "Test error created successfully",
		"error":     testErr.Error(),
		"trace_id":  tracing.GetTraceID(spanCtx),
		"timestamp": time.Now(),
	}

	return c.JSON(http.StatusInternalServerError, response)
}

// checkMemory performs a basic memory usage check.
func (h *Handler) checkMemory() string {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)

	// Consider memory usage > 1GB as warning
	if m.Alloc > 1024*1024*1024 {
		return constant.WarningStatus
	}

	return constant.HealthyStatus
}

// checkGoroutines performs a basic goroutine count check.
func (h *Handler) checkGoroutines() string {
	count := runtime.NumGoroutine()

	// Consider > 1000 goroutines as warning
	if count > 1000 {
		return constant.WarningStatus
	}

	return constant.HealthyStatus
}

// RegisterRoutes registers all monitoring routes.
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	// Health and readiness endpoints
	e.GET("/health", h.Health)
	e.GET("/ready", h.Ready)
	e.GET("/app-metrics", h.Metrics) // Changed from /metrics to /app-metrics
	e.GET("/info", h.Info)

	// Test endpoints for monitoring validation
	monitoring := e.Group("/test")
	monitoring.GET("/trace", h.TestTrace)
	monitoring.GET("/error", h.TestError)
}
