// Package server provides the HTTP server for the product service.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shopspring/decimal"

	handlers "github.com/raphaeldiscky/go-micro-template/services/product-service/internal/interface/http/handler"
)

// HTTPServer wraps the Echo server.
type HTTPServer struct {
	echo           *echo.Echo
	productHandler *handlers.ProductHandler
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(productHandler *handlers.ProductHandler) *HTTPServer {
	e := echo.New()

	// Create validator instance
	validate := validator.New()

	// Register custom validator for decimal fields
	if err := validate.RegisterValidation("decimal_gt", func(fl validator.FieldLevel) bool {
		if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
			return dec.GreaterThan(decimal.Zero)
		}

		return false
	}); err != nil {
		panic("failed to register decimal_gt validator: " + err.Error())
	}

	// Register validator
	e.Validator = &CustomValidator{validator: validate}

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

// CustomValidator wraps the go-playground validator.
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the struct using go-playground validator.
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}

	return nil
}
