package server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/interface/http/handlers"
)

// HTTPServer wraps the Echo server
type HTTPServer struct {
	echo          *echo.Echo
	sellerHandler *handlers.SellerHandler
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(sellerHandler *handlers.SellerHandler) *HTTPServer {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":    "healthy",
			"service":   "seller-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	return &HTTPServer{
		echo:          e,
		sellerHandler: sellerHandler,
	}
}

// RegisterRoutes registers all HTTP routes
func (s *HTTPServer) RegisterRoutes() {
	s.sellerHandler.RegisterRoutes(s.echo)
}

// Start starts the HTTP server
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

// Shutdown gracefully shuts down the HTTP server
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
