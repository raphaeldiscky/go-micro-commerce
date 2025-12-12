// Package routes provides the API gateway routes.
package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/gateway"
)

// SetupGatewayRoutes sets up the API gateway routes.
// API versioning is centrally managed here with /v1 prefix.
func SetupGatewayRoutes(
	e *echo.Echo,
	tel *telemetry.Telemetry,
	gw *gateway.Gateway,
	h *middleware.AuthMiddleware,
) {
	_ = tel

	// Debug endpoint to check service discovery (no versioning)
	e.GET("/debug/services", gw.DebugServices())

	// Health check routes (no versioning - infrastructure endpoints)
	health := e.Group("")
	health.GET("/auth/health", gw.ProxyToService("auth-service", "/health"))
	health.GET("/products/health", gw.ProxyToService("product-service", "/health"))
	health.GET("/orders/health", gw.ProxyToService("order-service", "/health"))
	health.GET("/notifications/health", gw.ProxyToService("notification-service", "/health"))
	health.GET("/fulfillments/health", gw.ProxyToService("fulfillment-service", "/health"))
	health.GET("/payments/health", gw.ProxyToService("payment-service", "/health"))
	health.GET("/searchs/health", gw.ProxyToService("search-service", "/health"))
	health.GET("/chats/health", gw.ProxyToService("chat-service", "/health"))
	health.GET("/carts/health", gw.ProxyToService("cart-service", "/health"))
	health.GET(
		"/chats/ws/health",
		gw.ProxyToService("chat-service-ws", "/ws/health"),
	) // use native websocket, not GraphQL subscriptions
	health.GET(
		"/notifications/sse/health",
		gw.ProxyToService("notification-service-sse", "/sse/health"),
	)

	// GraphQL Federation Gateway (no versioning - GraphQL has its own schema versioning)
	// This allows both authenticated and unauthenticated queries to work
	optionalAuth := e.Group("")
	optionalAuth.Use(h.OptionalAuthorization())
	optionalAuth.Any("/graph", gw.ProxyToService("apollo-router", "/"))

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

	// gRPC/ConnectRPC routes (no versioning - uses protobuf service naming)
	grpcProtected := e.Group("")
	grpcProtected.Use(h.Authorization())
	grpcProtected.Any("/product.v1.ProductService/*", gw.ProxyToConnectRPC("product-service-grpc"))

	// SSE debug routes (no versioning - protocol-level endpoint)
	sseProtected := e.Group("")
	sseProtected.Use(h.Authorization())
	sseProtected.GET(
		"/notifications/sse/debug/subscriptions",
		gw.ProxyToService("notification-service-sse", "/sse/debug/subscriptions"),
	)

	// Public auth routes (no authentication required)
	v1Public := e.Group("/v1")
	v1Public.POST("/auth/login", gw.ProxyToService("auth-service", "/login"))
	v1Public.POST("/auth/register", gw.ProxyToService("auth-service", "/register"))
	v1Public.POST("/auth/refresh-token", gw.ProxyToService("auth-service", "/refresh-token"))
	v1Public.POST("/auth/logout", gw.ProxyToService("auth-service", "/logout"))
	v1Public.POST("/auth/verify", gw.ProxyToService("auth-service", "/verify"))
	v1Public.POST(
		"/auth/resend-verification",
		gw.ProxyToService("auth-service", "/resend-verification"),
	)

	// Protected API v1 routes (authentication required)
	v1Protected := e.Group("/v1")
	v1Protected.Use(h.Authorization())
	v1Protected.Any("/products/*", gw.ProxyToService("product-service", ""))
	v1Protected.Any("/auth/*", gw.ProxyToService("auth-service", ""))
	v1Protected.Any("/orders/*", gw.ProxyToService("order-service", ""))
	v1Protected.Any("/notifications/*", gw.ProxyToService("notification-service", ""))
	v1Protected.Any("/fulfillments/*", gw.ProxyToService("fulfillment-service", ""))
	v1Protected.Any("/payments/*", gw.ProxyToService("payment-service", ""))
	v1Protected.Any("/searchs/*", gw.ProxyToService("search-service", ""))
	v1Protected.Any("/chats/*", gw.ProxyToService("chat-service", ""))
	v1Protected.Any("/carts/*", gw.ProxyToService("cart-service", ""))
}
