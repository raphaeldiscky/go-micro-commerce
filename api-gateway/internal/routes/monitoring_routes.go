package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/handler"
)

// SetupMonitoringRoutes registers all monitoring routes.
func SetupMonitoringRoutes(e *echo.Echo, h *handler.MonitoringHandler) {
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
