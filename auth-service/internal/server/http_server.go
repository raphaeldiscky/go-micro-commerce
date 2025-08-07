// Package server provides the HTTP server for the authentication service.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/validation"
)

// HTTPServer represents the HTTP server.
type HTTPServer struct {
	echo        *echo.Echo
	config      *config.Config
	logger      logger.Logger
	authHandler *handler.AuthHandler
}

// NewHTTPServer creates a new HTTP server instance.
func NewHTTPServer(
	authHandler *handler.AuthHandler,
	cfg *config.Config,
	lgr logger.Logger,
) (*HTTPServer, error) {
	e := echo.New()

	// Set custom validator
	e.Validator = validation.NewValidator()

	return &HTTPServer{
		echo:        e,
		config:      cfg,
		logger:      lgr,
		authHandler: authHandler,
	}, nil
}

// RegisterAuthRoutes registers the authentication routes.
func (s *HTTPServer) RegisterAuthRoutes() {
	routes.SetupAuthRoutes(s.echo, s.authHandler)
}

// Start starts the HTTP server.
func (s *HTTPServer) Start(port string) error {
	s.RegisterAuthRoutes()

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
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
