package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/middleware"
)

// SetupSSERoutes sets up all SSE routes.
func SetupSSERoutes(e *echo.Echo, h *handler.NotificationSSEHandler) {
	sse := e.Group("/sse")
	sse.GET("/health", h.Health)

	protected := sse.Group("")
	protected.Use(middleware.AuthMiddleware)
	protected.GET("/stream", h.StreamNotifications)
}
