// Package server provides the HTTP server for the product service.
package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	custommiddleware "github.com/raphaeldiscky/go-micro-template/pkg/middleware"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/validation"
)

// HTTPServer wraps the Echo server.
type HTTPServer struct {
	echo   *echo.Echo
	config *config.Config
	logger logger.Logger
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *HTTPServer {
	e := echo.New()

	// Register validator
	e.Validator = validation.NewValidator()

	// Middlewares
	RegisterMiddlewares(e)

	// Setup HTTP
	provider.SetupHTTP(cfg, e, appLogger, providers)

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
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	s.echo.Logger.Infof("Starting HTTP server on port %s", port)

	return s.echo.StartServer(server)
}

// Shutdown gracefully shuts down the HTTP server.
func (s *HTTPServer) Shutdown() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(s.config.HTTPServer.GracePeriod)*time.Second,
	)
	defer cancel()

	s.logger.Info("Attempting to shut down the HTTP server...")

	if err := s.echo.Shutdown(ctx); err != nil {
		s.logger.Fatal("Error shutting down HTTP server:", err)
	}

	s.logger.Info("HTTP server shut down gracefully")
}

// RegisterMiddlewares registers custom middleware for the HTTP server.
func RegisterMiddlewares(e *echo.Echo) {
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
		AllowOrigins: []string{"*"}, // Configure this properly for production
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
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1000))) // 1000 req/sec
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))
	e.Use(middleware.BodyLimit("10M"))
	e.Use(custommiddleware.ErrorHandler())
}
