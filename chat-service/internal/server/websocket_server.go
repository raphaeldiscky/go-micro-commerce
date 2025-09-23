package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// WebSocketServer represents the WebSocket server.
type WebSocketServer struct {
	echo   *echo.Echo
	hub    *websocket.ChatHub
	config *config.Config
	logger logger.Logger
}

// NewWebSocketServer creates a new WebSocket server instance.
func NewWebSocketServer(
	hub *websocket.ChatHub,
	cfg *config.Config,
	appLogger logger.Logger,
) *WebSocketServer {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: cfg.WebSocketServer.ReadTimeout,
	}))

	wsHandler := handler.NewWebSocketHandler(hub, appLogger)

	routes.SetupWebSocketRoutes(e, wsHandler)

	return &WebSocketServer{
		echo:   e,
		hub:    hub,
		config: cfg,
		logger: appLogger,
	}
}

// Start starts the WebSocket server.
func (s *WebSocketServer) Start(ctx context.Context) error {
	go func() {
		s.hub.Run(ctx)
	}()

	address := fmt.Sprintf(":%d", s.config.WebSocketServer.Port)
	s.logger.Info("Starting WebSocket server", "address", address)

	server := &http.Server{
		Addr:         address,
		Handler:      s.echo,
		ReadTimeout:  s.config.WebSocketServer.ReadTimeout,
		WriteTimeout: s.config.WebSocketServer.WriteTimeout,
		IdleTimeout:  s.config.WebSocketServer.IdleTimeout,
	}

	go func() {
		<-ctx.Done()
		s.logger.Info("Shutting down WebSocket server...")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			constant.DefaultShutdownTimeout*time.Second,
		)
		defer cancel()

		if err := s.hub.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error shutting down hub", "error", err)
		}

		if err := server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error shutting down server", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

// GetHub returns the WebSocket hub instance.
func (s *WebSocketServer) GetHub() *websocket.ChatHub {
	return s.hub
}
