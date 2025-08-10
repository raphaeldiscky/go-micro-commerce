// Package provider provides HTTP client and server utilities.
package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/routes"
)

// SetupHTTP initializes the HTTP server routes and middleware.
func SetupHTTP(cfg *config.Config, e *echo.Echo, appLogger logger.Logger, providers *Providers) {
	appHandler := handler.NewAppHandler()
	routes.SetupAppRoutes(e, appHandler)
	SetupAuth(cfg, e, appLogger, providers)
}
