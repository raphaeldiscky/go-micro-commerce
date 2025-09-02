// Package provider provides HTTP client and server utilities.
package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/routes"
)

// SetupHTTP initializes the HTTP server routes and middleware.
func SetupHTTP(_ *config.Config, e *echo.Echo, _ logger.Logger) {
	appHandler := handler.NewAppHandler()
	routes.SetupAppRoutes(e, appHandler)
}
