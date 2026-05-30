package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/consul"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/server"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// webSocketRunner manages the WebSocket server lifecycle.
type webSocketRunner struct {
	server *server.WebSocketServer
	hub    *websocket.ChatHub
	logger logger.Logger
}

// newWebSocketRunner creates a new WebSocket runner instance.
func newWebSocketRunner(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *webSocketRunner {
	wsServer := server.NewWebSocketServer(cfg, appLogger, providers)

	return &webSocketRunner{
		server: wsServer,
		hub:    providers.WebSocketHub,
		logger: appLogger,
	}
}

// Start starts the WebSocket runner.
func (r *webSocketRunner) Start(ctx context.Context) error {
	r.logger.Info("Starting WebSocket worker...")

	// EventBus is already initialized in provider, no need to start separately
	r.logger.Info("EventBus configured for cross-instance messaging",
		"active_subscriptions", r.hub.GetActiveChannelCount())

	if err := r.server.Start(ctx); err != nil {
		r.logger.Errorf("Failed to start WebSocket server: %v", err)
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the WebSocket runner.
func (r *webSocketRunner) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down WebSocket worker...")

	if err := r.hub.Shutdown(ctx); err != nil {
		r.logger.Errorf("Failed to shutdown WebSocket hub: %v", err)
		return err
	}

	r.logger.Info("WebSocket worker shut down successfully")

	return nil
}

// Name returns the runner name.
func (r *webSocketRunner) Name() string {
	return "websocket"
}

// newWebSocketCmd runs the WebSocket role.
func newWebSocketCmd() *cobra.Command {
	return roleCmd(
		"websocket",
		"Run the WebSocket server",
		func(app *appContext) ([]Runner, func()) {
			runner := newWebSocketRunner(app.cfg, app.logger, app.providers)

			return []Runner{runner}, registerConsulWebSocket(app.cfg, app.logger)
		},
	)
}

// registerConsulWebSocket registers the WebSocket service with Consul and
// returns a deregister cleanup func. It is a no-op when Consul is disabled.
func registerConsulWebSocket(cfg *config.Config, appLogger logger.Logger) func() {
	if !cfg.Consul.Enabled {
		appLogger.Infof("Consul service discovery is disabled")

		return func() {}
	}

	consulClient, err := consul.NewServiceRegistration(cfg.Consul.Address, appLogger)
	if err != nil {
		return func() {}
	}

	if err = consulClient.RegisterWebSocket(
		cfg.App.Name+"-ws",
		cfg.WebSocketServer.Host,
		cfg.WebSocketServer.Port,
	); err != nil {
		return func() {}
	}

	return func() {
		if deregErr := consulClient.Deregister(); deregErr != nil {
			appLogger.Errorf("Failed to deregister from Consul: %v", deregErr)
		}
	}
}
