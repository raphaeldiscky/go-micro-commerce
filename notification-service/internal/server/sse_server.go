// Package server provides the SSE server for GraphQL subscriptions.
package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	custommiddleware "github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/validation"
)

// SSEServer represents the dedicated SSE server for GraphQL subscriptions.
type SSEServer struct {
	echo   *echo.Echo
	config *config.Config
	logger logger.Logger
}

// NewSSEServer creates a new SSE server instance.
func NewSSEServer(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *SSEServer {
	e := echo.New()

	e.Validator = validation.NewValidator()

	// Middlewares optimized for SSE connections
	registerSSEMiddlewares(e, cfg)

	// Setup only GraphQL SSE routes
	provider.SetupSSE(cfg, e, appLogger, providers)

	return &SSEServer{
		echo:   e,
		config: cfg,
		logger: appLogger,
	}
}

// Start starts the SSE server.
func (s *SSEServer) Start() error {
	port := strconv.Itoa(s.config.SSEServer.Port)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           s.echo,
		ReadTimeout:       s.config.SSEServer.Timeout,
		WriteTimeout:      s.config.SSEServer.Timeout,
		IdleTimeout:       s.config.SSEServer.Timeout,
		ReadHeaderTimeout: s.config.HTTPServer.ReadHeaderTimeout,
		MaxHeaderBytes:    s.config.HTTPServer.MaxHeaderBytes,
	}

	s.echo.Logger.Infof("Starting SSE server on port %s", port)

	return s.echo.StartServer(server)
}

// Shutdown gracefully shuts down the SSE server.
func (s *SSEServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Attempting to shut down the SSE server...")

	if err := s.echo.Shutdown(ctx); err != nil {
		s.logger.Errorf("Error shutting down SSE server: %v", err)

		return err
	}

	s.logger.Info("SSE server shut down gracefully")

	return nil
}

// registerSSEMiddlewares registers custom middleware optimized for SSE connections.
func registerSSEMiddlewares(e *echo.Echo, cfg *config.Config) {
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
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderCacheControl,
		},
	}))
	// Relaxed security for SSE connections
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            cfg.HTTPServer.HSTSMaxAge,
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline' 'unsafe-eval' data: blob: https://cdn.jsdelivr.net https://unpkg.com;",
	}))
	// More lenient rate limiting for SSE connections
	e.Use(
		middleware.RateLimiter(
			middleware.NewRateLimiterMemoryStore(cfg.SSEServer.RateLimiter),
		), // 100 req/sec for SSE
	)
	// Larger body limit for GraphQL queries
	e.Use(middleware.BodyLimit("10M"))
	e.Use(custommiddleware.ErrorHandler())
}
