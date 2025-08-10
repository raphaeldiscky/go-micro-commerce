// Package provider provides HTTP client and server utilities.
package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/routes"
)

// SetupHTTP initializes the HTTP server routes and middleware.
func SetupHTTP(cfg *config.Config, e *echo.Echo, appLogger logger.Logger, gw *gateway.Gateway) {
	appHandler := handler.NewAppHandler()
	monitoringController := handler.NewMonitoringHandler(cfg, appLogger)

	routes.SetupMonitoringRoutes(e, monitoringController)
	routes.SetupGatewayRoutes(e, gw)
	routes.SetupAppRoutes(e, appHandler)
}
