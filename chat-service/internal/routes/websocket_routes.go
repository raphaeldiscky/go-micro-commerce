package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
)

// SetupWebSocketRoutes sets up all WebSocket routes.
func SetupWebSocketRoutes(e *echo.Echo, wsHandler *handler.WebSocketHandler) {
	e.GET("/ws/health", wsHandler.WebSocketHealth)

	v1 := e.Group("/v1")
	v1.GET("/ws", wsHandler.HandleWebSocket)
	v1.GET("/ws/stats", wsHandler.GetConnectionStats)
}
