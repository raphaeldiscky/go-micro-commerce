package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/server"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// WebSocketWorker manages the WebSocket server lifecycle.
type WebSocketWorker struct {
	server *server.WebSocketServer
	hub    *websocket.ChatHub
	logger logger.Logger
}

// NewWebSocketWorker creates a new WebSocket worker instance.
func NewWebSocketWorker(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *WebSocketWorker {
	wsServer := server.NewWebSocketServer(cfg, appLogger, providers)

	return &WebSocketWorker{
		server: wsServer,
		hub:    providers.WebSocketHub,
		logger: appLogger,
	}
}

// Start starts the WebSocket worker.
func (w *WebSocketWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting WebSocket worker...")

	// EventBus is already initialized in provider, no need to start separately
	w.logger.Info("EventBus configured for cross-instance messaging",
		"active_subscriptions", w.hub.GetActiveChannelCount())

	if err := w.server.Start(ctx); err != nil {
		w.logger.Errorf("Failed to start WebSocket server: %v", err)
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the WebSocket worker.
func (w *WebSocketWorker) Shutdown(ctx context.Context) error {
	w.logger.Info("Shutting down WebSocket worker...")

	if err := w.hub.Shutdown(ctx); err != nil {
		w.logger.Errorf("Failed to shutdown WebSocket hub: %v", err)
		return err
	}

	w.logger.Info("WebSocket worker shut down successfully")

	return nil
}

// Name returns the worker name.
func (w *WebSocketWorker) Name() string {
	return "websocket"
}
