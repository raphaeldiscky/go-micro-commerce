// Package provider provides HTTP client and server utilities.
package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/routes"
)

// SetupHTTP initializes the HTTP server routes and middleware.
func SetupHTTP(
	e *echo.Echo,
	cfg *config.Config,
	appLogger logger.Logger,
	gw *gateway.Gateway,
	providers *Providers,
) {
	appHandler := handler.NewAppHandler()
	monitoringController := handler.NewMonitoringHandler(cfg, appLogger)

	routes.SetupMonitoringRoutes(e, monitoringController)
	routes.SetupGatewayRoutes(e, gw, providers.authMiddleware)
	routes.SetupAppRoutes(e, appHandler)
}
