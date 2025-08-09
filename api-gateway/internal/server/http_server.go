// Package server provides the HTTP server for the authentication service.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/metrics"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/ratelimit"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/tracing"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/routes"
)

// HTTPServer represents the HTTP server.
type HTTPServer struct {
	echo              *echo.Echo
	config            *config.Config
	logger            logger.Logger
	gateway           *gateway.Gateway
	monitoringHandler *handler.MonitoringHandler
}

// NewHTTPServer creates a new HTTP server instance.
func NewHTTPServer(
	gw *gateway.Gateway,
	metricsInstance *metrics.Metrics,
	cfg *config.Config,
	lgr logger.Logger,
	monitoringHandler *handler.MonitoringHandler,
) *HTTPServer {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Custom Middleware
	e.Use(tracing.Middleware())
	e.Use(metricsInstance.Middleware())
	e.Use(ratelimit.Middleware(*cfg.RateLimit))

	return &HTTPServer{
		echo:              e,
		config:            cfg,
		logger:            lgr,
		gateway:           gw,
		monitoringHandler: monitoringHandler,
	}
}

// RegisterRoutes registers the authentication routes.
func (s *HTTPServer) RegisterRoutes() {
	routes.SetupMonitoringRoutes(s.echo, s.monitoringHandler)
	routes.SetupGatewayRoutes(s.echo, s.gateway)
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
