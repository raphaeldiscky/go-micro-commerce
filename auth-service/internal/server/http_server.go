package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/infra/db/postgres"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/service"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/validation"
)

// HTTPServer represents the HTTP server.
type HTTPServer struct {
	echo   *echo.Echo
	config *config.Config
	logger logger.Logger
}

// NewHTTPServer creates a new HTTP server instance.
func NewHTTPServer(config *config.Config, logger logger.Logger) (*HTTPServer, error) {
	e := echo.New()

	// Set custom validator
	e.Validator = validation.NewValidator()

	return &HTTPServer{
		echo:   e,
		config: config,
		logger: logger,
	}, nil
}

// Start starts the HTTP server.
func (s *HTTPServer) Start(port string) error {
	// Initialize postgres connection
	db, err := postgres.NewConnection(s.config.Postgres)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepositoryPostgres(db)
	sessionRepo := postgres.NewSessionRepository(db)

	// Initialize event publisher
	eventPublisher := event.NewSimplePublisher(s.config.EventPublisher)

	// Initialize service
	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		s.config.JWT,
		eventPublisher,
		s.logger,
	)

	// Initialize handler
	authHandler := handler.NewAuthHandler(authService, s.logger)

	// Setup routes
	routes.SetupAuthRoutes(s.echo, authHandler)

	// Start server
	go func() {
		addr := fmt.Sprintf(":%s", port)
		s.logger.Info("Starting HTTP server", "address", addr)

		if err := s.echo.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	s.logger.Info("Server exited")

	return nil
}

// Shutdown gracefully shuts down the HTTP server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
