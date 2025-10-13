package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph/resolver"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/subscription"
)

// SetupWebsocket initializes the Websocket server routes for GraphQL subscriptions.
func SetupWebsocket(
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	hub := providers.WebSocketHub
	if hub == nil {
		appLogger.Fatal("WebSocket hub is nil during worker creation")
	}

	connectionService := providers.ConnectionService
	if connectionService == nil {
		appLogger.Fatal("Connection service is nil during worker creation")
	}

	chatService := providers.ChatService
	if chatService == nil {
		appLogger.Fatal("Chat service is nil during worker creation")
	}

	wsHandler := handler.NewWebSocketHandler(
		hub,
		appLogger,
		cfg.WebSocketServer,
		connectionService,
		chatService,
	)

	// Create subscription manager for GraphQL subscriptions
	subscriptionManager := subscription.NewManager(hub, providers.EventBus, appLogger)

	// Create GraphQL resolver
	graphqlResolver := resolver.NewResolver(
		chatService,
		connectionService,
		subscriptionManager,
		hub,
		appLogger,
	)

	routes.SetupWebSocketRoutes(e, wsHandler)
	routes.SetupGraphQLWsRoutes(e, graphqlResolver, appLogger)
}
