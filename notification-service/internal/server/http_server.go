// Package server provides the HTTP server for the notificationentication service.
package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/validation"
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
	providers *provider.Providers,
) *HTTPServer {
	e := echo.New()

	// Set custom validator
	e.Validator = validation.NewValidator()

	// Middleware
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
		Addr:         ":" + port,
		Handler:      s.echo,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
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
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
}
