package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	custommiddleware "github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/provider"
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
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *WebSocketServer {
	e := echo.New()

	registerWebSocketMiddlewares(e, cfg)

	provider.SetupNativeWebsocket(cfg, e, appLogger, providers)

	return &WebSocketServer{
		echo:   e,
		hub:    providers.WebSocketHub,
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
			constant.DefaultShutdownTimeout,
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

// registerWebSocketMiddlewares registers custom websocket middleware for the HTTP server.
func registerWebSocketMiddlewares(e *echo.Echo, cfg *config.Config) {
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return uuid.New().String()
		},
	}))
	e.Use(middleware.LoggerWithConfig(
		middleware.LoggerConfig{
			Format: "[${time_rfc3339}] ${method} ${uri} ${status} ${latency_human}\n",
		},
	))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3001", // React development server
			"http://127.0.0.1:3001",
			"https://go.micro.commerce:3001",
			"http://localhost:3002",
			"http://127.0.0.1:3002",
			"http://localhost:3003",
			"http://127.0.0.1:3003",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderUpgrade,
			echo.HeaderConnection,
			"Sec-WebSocket-Key",      // Required for WebSocket
			"Sec-WebSocket-Version",  // Required for WebSocket
			"Sec-WebSocket-Protocol", // Required for GraphQL subscriptions
		},
	}))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            cfg.WebSocketServer.HSTSMaxAge,
		ContentSecurityPolicy: "default-src 'self'",
	}))
	e.Use(
		middleware.RateLimiter(
			middleware.NewRateLimiterMemoryStore(cfg.WebSocketServer.RateLimiter),
		),
	) // 1000 req/sec
	e.Use(middleware.BodyLimit("10M"))
	e.Use(custommiddleware.ErrorHandler())
}

// GetHub returns the WebSocket hub instance.
func (s *WebSocketServer) GetHub() *websocket.ChatHub {
	return s.hub
}
