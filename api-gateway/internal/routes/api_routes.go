// Package routes provides the API gateway routes.
package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
)

// SetupAPIGatewayRoutes sets up the API gateway routes.
func SetupAPIGatewayRoutes(e *echo.Echo, gw *gateway.Gateway) {
	// Gateway health endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	// Debug endpoint to check service discovery
	e.GET("/debug/services", gw.DebugServices())

	v1 := e.Group("/api/v1")
	// Auth service
	v1.Any("/auth/*", gw.ProxyToService("auth-service", ""))
	v1.Any("/products/*", gw.ProxyToService("product-service", ""))
	v1.Any("/users/*", gw.ProxyToService("auth-service", ""))
	v1.Any("/orders/*", gw.ProxyToService("order-service", ""))
	v1.Any("/notifications/*", gw.ProxyToService("notification-service", ""))
}
