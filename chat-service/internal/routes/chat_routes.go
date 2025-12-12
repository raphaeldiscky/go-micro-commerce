package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/middleware"
)

// SetupChatRoutes sets up all chat routes.
func SetupChatRoutes(
	e *echo.Echo,
	h *handler.ChatHandler,
	connHandler *handler.ConnectionHandler,
) {
	protected := e.Group("")
	protected.Use(middleware.AuthMiddleware)

	// Connection management routes
	protected.POST("/connect", connHandler.RequestConnection)
	protected.GET("/nodes/health", connHandler.GetNodeHealth)

	// Chat data management routes (read-only and CRUD operations)
	protected.GET("/conversations", h.GetUserConversations)
	protected.POST("/conversations", h.CreateConversation)
	protected.GET("/:conversationID", h.GetConversation)
	protected.GET("/:conversationID/messages", h.GetMessages)
	protected.POST("/:conversationID/join", h.JoinConversation)
	protected.GET("/:conversationID/participants", h.GetParticipants)
	protected.GET("/users/online", h.GetOnlineUsers)

	// NOTE: The following operations are handled via WebSocket messages on /ws:
	// - Send Message -> WebSocket message type "chat"
	// - Update Presence -> WebSocket message type "presence"
	// - Typing Indicator -> WebSocket message type "typing"
	// - Delivery Receipt -> WebSocket message type "delivery_receipt"
	// - Read Receipt -> WebSocket message type "read_receipt"
}
