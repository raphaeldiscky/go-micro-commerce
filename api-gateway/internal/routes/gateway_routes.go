// Package routes provides the API gateway routes.
package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/middleware/metrics"
)

// SetupGatewayRoutes sets up the API gateway routes.
func SetupGatewayRoutes(e *echo.Echo, gw *gateway.Gateway, h *middleware.AuthMiddleware) {
	// Metrics endpoint (Prometheus format)
	e.GET("/metrics", metrics.Handler())
	// Debug endpoint to check service discovery
	e.GET("/debug/services", gw.DebugServices())
	// Public routes
	public := e.Group("")
	public.POST("/auth/v1/login", gw.ProxyToService("auth-service", "/v1/login"))
	public.POST("/auth/v1/register", gw.ProxyToService("auth-service", "/v1/register"))
	public.POST("/auth/v1/refresh-token", gw.ProxyToService("auth-service", "/v1/refresh-token"))
	public.POST("/auth/v1/logout", gw.ProxyToService("auth-service", "/v1/logout"))
	public.POST("/auth/v1/verify", gw.ProxyToService("auth-service", "/v1/verify"))
	public.POST(
		"/auth/v1/resend-verification",
		gw.ProxyToService("auth-service", "/v1/resend-verification"),
	)

	// Protected routes
	protected := e.Group("")
	protected.Use(h.Authorization())
	protected.Any("/products/*", gw.ProxyToService("product-service", ""))
	protected.Any("/auth/v1/users/*", gw.ProxyToService("auth-service", "/v1/users"))
	protected.Any("/orders/*", gw.ProxyToService("order-service", ""))
	protected.Any("/notifications/*", gw.ProxyToService("notification-service", ""))
}
