package provider

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/subscription"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// SetupChat initializes the chat-related routes and services.
func SetupChat(
	_ context.Context,
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	// Use the pre-initialized WebSocket hub
	hub := providers.WebSocketHub
	if hub == nil {
		appLogger.Fatal("WebSocket hub is nil in SetupChat")
	}

	// Create chat service
	chatService := service.NewChatService(
		providers.DataStore,
		appLogger,
		hub,
	)
	providers.ChatService = chatService

	// Create connection service
	nodeConfig := &service.NodeConfig{
		DefaultNodeAddress: cfg.Connection.DefaultNodeAddress,
		MaxConnections:     cfg.Connection.MaxConnections,
		ConsulAddress:      cfg.Connection.ConsulAddress,
		ChatServiceName:    cfg.Connection.ChatServiceName,
	}
	connectionService := service.NewConnectionService(
		appLogger,
		cfg.Connection.PublicKeyPath,
		cfg.Connection.JWKSUrl,
		cfg.Connection.JWKSCacheTTL,
		cfg.Connection.JWKSRefreshInterval,
		nodeConfig,
	)
	providers.ConnectionService = connectionService

	// Create handlers
	chatHandler := handler.NewChatHandler(chatService, hub, appLogger)
	connectionHandler := handler.NewConnectionHandler(connectionService, appLogger)

	// Create subscription manager for GraphQL subscriptions
	subscriptionManager := subscription.NewManager(hub, providers.ChatPubSub, appLogger)

	// Create message publisher for mutations
	messagePublisher := websocket.NewMessagePublisher(providers.ChatPubSub)

	// Create GraphQL resolver
	graphqlResolver := resolver.NewResolver(
		chatService,
		connectionService,
		subscriptionManager,
		messagePublisher,
		appLogger,
	)

	// Setup routes
	routes.SetupChatRoutes(e, chatHandler, connectionHandler)
	routes.SetupGraphQLRoutes(e, cfg, graphqlResolver)
}
