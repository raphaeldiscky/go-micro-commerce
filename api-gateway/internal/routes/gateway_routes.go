// Package routes provides the API gateway routes.
package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/metrics"
)

// SetupGatewayRoutes sets up the API gateway routes.
func SetupGatewayRoutes(e *echo.Echo, gw *gateway.Gateway) {
	// Gateway health endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})
	// Metrics endpoint (Prometheus format)
	e.GET("/metrics", metrics.Handler())
	// Debug endpoint to check service discovery
	e.GET("/debug/services", gw.DebugServices())

	// Auth service
	e.Any("/auth/*", gw.ProxyToService("auth-service", ""))
	e.Any("/products/*", gw.ProxyToService("product-service", ""))
	e.Any("/users/*", gw.ProxyToService("auth-service", ""))
	e.Any("/orders/*", gw.ProxyToService("order-service", ""))
	e.Any("/notifications/*", gw.ProxyToService("notification-service", ""))
}
