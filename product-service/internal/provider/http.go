// Package provider provides HTTP client and server utilities.
package provider

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/routes"
)

// SetupHTTP sets up the HTTP routes and middleware.
func SetupHTTP(
	ctx context.Context,
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	appHandler := handler.NewAppHandler()
	routes.SetupAppRoutes(e, appHandler, cfg)
	SetupProduct(ctx, cfg, e, appLogger, providers)
}
