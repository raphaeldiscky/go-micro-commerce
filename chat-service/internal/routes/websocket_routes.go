package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
)

// SetupWebSocketRoutes sets up all WebSocket routes.
func SetupWebSocketRoutes(e *echo.Echo, wsHandler *handler.WebSocketHandler) {
	ws := e.Group("/ws")
	ws.GET("", wsHandler.HandleWebSocket)
	ws.GET("/health", wsHandler.Health)
	ws.GET("/stats", wsHandler.GetConnectionStats)
}
