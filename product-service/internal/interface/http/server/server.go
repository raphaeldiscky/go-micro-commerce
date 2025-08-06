// Package server provides the HTTP server for the product service.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	handlers "github.com/raphaeldiscky/go-micro-template/product-service/internal/interface/http/handler"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/interface/http/validation"
)

// HTTPServer wraps the Echo server.
type HTTPServer struct {
	echo           *echo.Echo
	productHandler *handlers.ProductHandler
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(productHandler *handlers.ProductHandler) *HTTPServer {
	e := echo.New()

	// Register validator
	e.Validator = validation.NewValidator()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":    "healthy",
			"service":   "product-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	return &HTTPServer{
		echo:           e,
		productHandler: productHandler,
	}
}

// RegisterRoutes registers all HTTP routes.
func (s *HTTPServer) RegisterRoutes() {
	s.productHandler.RegisterRoutes(s.echo)
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
