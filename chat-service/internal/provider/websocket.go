package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/routes"
)

// SetupNativeWebsocket initializes the Websocket server routes for GraphQL subscriptions.
func SetupNativeWebsocket(
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

	routes.SetupNativeWebSocketRoutes(e, wsHandler)
}
