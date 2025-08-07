// Package server provides the HTTP server for the product service.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	handlers "github.com/raphaeldiscky/go-micro-template/product-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/validation"
)

// HTTPServer wraps the Echo server.
type HTTPServer struct {
	echo           *echo.Echo
	config         *config.HTTPServerConfig
	productHandler *handlers.ProductHandler
	logger         logger.Logger
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(
	productHandler *handlers.ProductHandler,
	cfg *config.HTTPServerConfig,
	lgr logger.Logger,
) *HTTPServer {
	e := echo.New()

	// Register validator
	e.Validator = validation.NewValidator()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	return &HTTPServer{
		echo:           e,
		productHandler: productHandler,
		config:         cfg,
		logger:         lgr,
	}
}

// RegisterRoutes registers all HTTP routes.
func (s *HTTPServer) RegisterRoutes() {
	routes.SetupProductRoutes(s.echo, s.productHandler)
}

// Start starts the HTTP server.
func (s *HTTPServer) Start(port string) error {
	s.RegisterRoutes()

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
