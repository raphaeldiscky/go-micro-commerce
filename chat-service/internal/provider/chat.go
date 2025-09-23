package provider

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/pubsub"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// SetupChat initializes the chat-related routes and services.
func SetupChat(
	_ context.Context,
	_ *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	// Create chat pub/sub service
	chatPubSub := pubsub.NewChatPubSub(
		providers.RedisPublisher,
		providers.RedisSubscriber,
		appLogger,
	)

	// Create WebSocket hub with Redis pub/sub integration
	hub := websocket.NewChatHub(providers.ConnectionRepository, appLogger, chatPubSub)
	providers.WebSocketHub = hub

	// Create chat service
	chatService := service.NewChatService(
		providers.DataStore,
		appLogger,
		hub,
	)
	providers.ChatService = chatService

	// Create chat handler with WebSocket hub integration
	chatHandler := handler.NewChatHandler(chatService, hub, appLogger)

	// Setup routes
	routes.SetupChatRoutes(e, chatHandler)
}
