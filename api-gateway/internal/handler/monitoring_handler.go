// Package handler for monitoring and gateways
package handler

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/middleware/tracing"
)

// MonitoringHandler handles monitoring and health check endpoints.
type MonitoringHandler struct {
	config    *config.Config
	logger    logger.Logger
	startTime time.Time
}

// NewMonitoringHandler creates a new monitoring handler.
func NewMonitoringHandler(cfg *config.Config, appLogger logger.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		config:    cfg,
		logger:    appLogger,
		startTime: time.Now(),
	}
}

// Health returns the health status of the API gateway.
func (h *MonitoringHandler) Health(c echo.Context) error {
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

	response := dto.HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		Checks:    checks,
	}

	return echoutils.ResponseOK(c, response)
}

// Ready returns the readiness status of the API gateway.
func (h *MonitoringHandler) Ready(c echo.Context) error {
	// Check if the service is ready to accept requests
	// Add any specific readiness checks here
	response := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now(),
	}

	return echoutils.ResponseOK(c, response)
}

// Metrics returns basic application metrics.
func (h *MonitoringHandler) Metrics(c echo.Context) error {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)

	uptime := time.Since(h.startTime)

	response := dto.MetricsResponse{
		Timestamp:      time.Now(),
		Uptime:         uptime.String(),
		MemoryUsage:    m.Alloc,
		GoroutineCount: runtime.NumGoroutine(),
	}

	return echoutils.ResponseOK(c, response)
}

// Info returns general information about the API gateway.
func (h *MonitoringHandler) Info(c echo.Context) error {
	response := map[string]interface{}{
		"service":    "api-gateway",
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

	return echoutils.ResponseOK(c, response)
}

// TestTrace creates a test trace for monitoring validation.
func (h *MonitoringHandler) TestTrace(c echo.Context) error {
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
		"trace_id", traceID,
		"span_id", spanID,
	)

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Test trace created successfully",
		"trace_id":  traceID,
		"span_id":   spanID,
		"timestamp": time.Now(),
	}

	return echoutils.ResponseOK(c, response)
}

// TestError creates a test error for monitoring validation.
func (h *MonitoringHandler) TestError(c echo.Context) error {
	ctx := c.Request().Context()

	// Create a span for error testing
	spanCtx, endFunc := tracing.StartSpan(ctx, "test-error-operation")
	defer endFunc()

	// Create a test error
	testErr := fmt.Errorf("this is a test error for monitoring validation")

	// Record the error in the span
	tracing.SetSpanError(spanCtx, testErr)

	h.logger.Error("Test error created for monitoring",
		"error", testErr,
		"trace_id", tracing.GetTraceID(spanCtx),
	)

	response := map[string]interface{}{
		"status":    "error",
		"error":     testErr.Error(),
		"trace_id":  tracing.GetTraceID(spanCtx),
		"timestamp": time.Now(),
	}

	return echoutils.ResponseJSON(
		c,
		http.StatusInternalServerError,
		"Test error created successfully",
		response,
		nil,
	)
}

// checkMemory performs a basic memory usage check.
func (h *MonitoringHandler) checkMemory() string {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)

	// Consider memory usage > 1GB as warning
	if m.Alloc > 1024*1024*1024 {
		return constant.WarningStatus
	}

	return constant.HealthyStatus
}

// checkGoroutines performs a basic goroutine count check.
func (h *MonitoringHandler) checkGoroutines() string {
	count := runtime.NumGoroutine()

	// Consider > 1000 goroutines as warning
	if count > 1000 {
		return constant.WarningStatus
	}

	return constant.HealthyStatus
}
