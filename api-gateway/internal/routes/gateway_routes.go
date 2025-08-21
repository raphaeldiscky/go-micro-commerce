// Package routes provides the API gateway routes.
package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/metrics"
)

// SetupGatewayRoutes sets up the API gateway routes.
func SetupGatewayRoutes(e *echo.Echo, gw *gateway.Gateway, h *middleware.AuthMiddleware) {
	// Metrics endpoint (Prometheus format)
	e.GET("/metrics", metrics.Handler())
	// Debug endpoint to check service discovery
	e.GET("/debug/services", gw.DebugServices())

	// Public routes
	e.Any("/auth/*", gw.ProxyToService("auth-service", ""))
	e.Any("/products/*", gw.ProxyToService("product-service", ""))

	// Protected routes
	protected := e.Group("")
	protected.Use(h.Authorization())
	protected.Any("/users/*", gw.ProxyToService("auth-service", ""))
	protected.Any("/orders/*", gw.ProxyToService("order-service", ""))
	protected.Any("/notifications/*", gw.ProxyToService("notification-service", ""))
}
