package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/middleware"
)

// SetupWebSocketRoutes sets up all WebSocket routes.
func SetupWebSocketRoutes(e *echo.Echo, wsHandler *handler.WebSocketHandler) {
	e.GET("/ws/health", wsHandler.WebSocketHealth)

	v1 := e.Group("/v1")
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)
	protected.GET("/ws", wsHandler.HandleWebSocket)
	protected.GET("/ws/stats", wsHandler.GetConnectionStats)
	protected.GET("/ws/admin", wsHandler.HandleAdminWebSocket)
}
