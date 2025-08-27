// Package provider provides HTTP client and server utilities.
package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/routes"
)

// SetupHTTP sets up the HTTP routes and middleware.
func SetupHTTP(cfg *config.Config, e *echo.Echo, appLogger logger.Logger, providers *Providers) {
	appHandler := handler.NewAppHandler()
	routes.SetupAppRoutes(e, appHandler)
	SetupOrder(cfg, e, appLogger, providers)
}
