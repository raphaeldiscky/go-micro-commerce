package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/middleware"
)

// SetupChatRoutes sets up all chat routes.
func SetupChatRoutes(e *echo.Echo, h *handler.ChatHandler, connHandler *handler.ConnectionHandler) {
	v1 := e.Group("/v1")

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)

	// Connection management routes
	protected.POST("/connect", connHandler.RequestConnection)
	protected.GET("/nodes/health", connHandler.GetNodeHealth)
	protected.POST("/validate-ticket", connHandler.ValidateTicket)

	// Chat conversation routes
	protected.POST("/", h.CreateConversation)
	protected.GET("/:conversationID", h.GetConversation)
	protected.POST("/:conversationID/messages", h.SendMessage)
	protected.GET("/:conversationID/messages", h.GetMessages)
	protected.POST("/:conversationID/join", h.JoinConversation)
	protected.GET("/:conversationID/participants", h.GetParticipants)
}
