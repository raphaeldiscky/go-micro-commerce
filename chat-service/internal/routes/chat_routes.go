package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/middleware"
)

// SetupChatRoutes sets up all chat routes.
func SetupChatRoutes(e *echo.Echo, h *handler.ChatHandler) {
	v1 := e.Group("/v1")

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)

	// Chat conversation routes
	protected.POST("/conversations", h.CreateConversation)
	protected.GET("/conversations/:conversationID", h.GetConversation)
	protected.POST("/conversations/:conversationID/messages", h.SendMessage)
	protected.GET("/conversations/:conversationID/messages", h.GetMessages)
	protected.POST("/conversations/:conversationID/join", h.JoinConversation)
	protected.PATCH("/conversations/:conversationID/status", h.UpdateConversationStatus)
}
