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
	public.GET("/auths/health", gw.ProxyToService("auth-service", "/health"))
	public.GET("/products/health", gw.ProxyToService("product-service", "/health"))
	public.GET("/orders/health", gw.ProxyToService("order-service", "/health"))
	public.GET("/notifications/health", gw.ProxyToService("notification-service", "/health"))
	public.GET("/fulfillments/health", gw.ProxyToService("fulfillment-service", "/health"))
	public.GET("/payments/health", gw.ProxyToService("payment-service", "/health"))
	public.GET("/searchs/health", gw.ProxyToService("search-service", "/health"))
	public.GET("/chats/health", gw.ProxyToService("chat-service", "/health"))
	public.GET("/carts/health", gw.ProxyToService("cart-service", "/health"))
	public.GET(
		"/chats/ws/health",
		gw.ProxyToService("chat-service-ws", "/ws/health"),
	) // use native websocket, not GraphQL subscriptions
	public.GET(
		"/notifications/sse/health",
		gw.ProxyToService("notification-service-sse", "/sse/health"),
	)

	// GraphQL Federation Gateway (with optional auth - validates JWT if present)
	// This allows both authenticated and unauthenticated queries to work
	optionalAuth := e.Group("")
	optionalAuth.Use(h.OptionalAuthorization())
	optionalAuth.Any("/graph", gw.ProxyToService("graphql-gateway", "/"))

	// GraphQL Subscriptions WebSocket (bypass Apollo Router, proxy directly to chat-service)
	// Apollo Router doesn't support WebSocket subscriptions, so we route directly
	// Note: Use chat-service (port 8085) NOT chat-service-ws (port 9095)
	// GraphQL subscriptions are on the HTTP server, not the native WebSocket server
	optionalAuth.GET(
		"/graph/subscriptions/ws",
		gw.ProxyWebSocket("chat-service", "/graph/subscriptions/ws"),
	)
	// GraphQL SSE Subscriptions (bypass Apollo Router, proxy directly to notification-service)
	// SSE uses standard HTTP streaming (text/event-stream), not WebSocket protocol
	// Supports both GET (query in URL) and POST (query in body) methods
	// Uses ProxySSE for long-lived streaming without timeouts
	optionalAuth.GET(
		"/graph/subscriptions/sse",
		gw.ProxySSE("notification-service-sse", "/graph/subscriptions/sse"),
	)
	optionalAuth.POST(
		"/graph/subscriptions/sse",
		gw.ProxySSE("notification-service-sse", "/graph/subscriptions/sse"),
	)

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
	protected.Any("/product.v1.ProductService/*", gw.ProxyToConnectRPC("product-service-grpc"))
	protected.Any("/auth/*", gw.ProxyToService("auth-service", ""))
	protected.Any("/orders/*", gw.ProxyToService("order-service", ""))
	protected.GET(
		"/notifications/sse/debug/subscriptions",
		gw.ProxyToService("notification-service-sse", "/sse/debug/subscriptions"),
	)
	protected.Any("/notifications/*", gw.ProxyToService("notification-service", ""))
	protected.Any("/fulfillments/*", gw.ProxyToService("fulfillment-service", ""))
	protected.Any("/payments/*", gw.ProxyToService("payment-service", ""))
	protected.Any("/searchs/*", gw.ProxyToService("search-service", ""))
	protected.Any("/chats/*", gw.ProxyToService("chat-service", ""))
	protected.Any("/carts/*", gw.ProxyToService("cart-service", ""))
}
