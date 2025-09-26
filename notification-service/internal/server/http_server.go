// Package server provides the HTTP server for the notificationentication service.
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

// HTTPServer represents the HTTP server.
type HTTPServer struct {
	echo   *echo.Echo
	config *config.Config
	logger logger.Logger
}

// NewHTTPServer creates a new HTTP server instance.
func NewHTTPServer(
	cfg *config.Config,
	appLogger logger.Logger,
) *HTTPServer {
	e := echo.New()

	// Set custom validator
	e.Validator = validation.NewValidator()

	// Middlewares
	registerMiddlewares(e, cfg)

	// Setup HTTP
	provider.SetupHTTP(cfg, e, appLogger)

	return &HTTPServer{
		echo:   e,
		config: cfg,
		logger: appLogger,
	}
}

// Start starts the HTTP server.
func (s *HTTPServer) Start() error {
	port := strconv.Itoa(s.config.HTTPServer.Port)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           s.echo,
		ReadTimeout:       s.config.HTTPServer.ReadTimeout,
		WriteTimeout:      s.config.HTTPServer.WriteTimeout,
		IdleTimeout:       s.config.HTTPServer.IdleTimeout,
		ReadHeaderTimeout: s.config.HTTPServer.ReadHeaderTimeout,
		MaxHeaderBytes:    s.config.HTTPServer.MaxHeaderBytes,
	}

	s.echo.Logger.Infof("Starting HTTP server on port %s", port)

	return s.echo.StartServer(server)
}

// Shutdown gracefully shuts down the HTTP server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Attempting to shut down the HTTP server...")

	if err := s.echo.Shutdown(ctx); err != nil {
		s.logger.Errorf("Error shutting down HTTP server: %v", err)

		return err
	}

	s.logger.Info("HTTP server shut down gracefully")

	return nil
}

// registerMiddlewares registers custom middleware for the HTTP server.
func registerMiddlewares(e *echo.Echo, cfg *config.Config) {
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
		},
	}))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            cfg.HTTPServer.HSTSMaxAge,
		ContentSecurityPolicy: "default-src 'self'",
	}))
	e.Use(
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(cfg.HTTPServer.RateLimiter)),
	) // 1000 req/sec
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: cfg.HTTPServer.IdleTimeout,
	}))
	e.Use(middleware.BodyLimit("10M"))
	e.Use(custommiddleware.ErrorHandler())
}
