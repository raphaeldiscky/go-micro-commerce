package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/middleware"
)

// SetupSSERoutes sets up all SSE routes.
func SetupSSERoutes(
	e *echo.Echo,
	h *handler.NotificationSSEHandler,
	debugHandler *handler.DebugHandler,
) {
	sse := e.Group("/sse")
	sse.GET("/health", h.Health)

	// Debug endpoints (should be protected in production)
	debug := sse.Group("/debug")
	debug.GET("/subscriptions", debugHandler.GetActiveSubscriptions)

	protected := sse.Group("")
	protected.Use(middleware.AuthMiddleware)
	protected.GET("/stream", h.StreamNotifications)
}
