// Package provider provides HTTP client and server utilities.
package provider

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/routes"
)

// SetupHTTP sets up the HTTP routes and middleware.
func SetupHTTP(
	ctx context.Context,
	cfg *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	tel *telemetry.Telemetry,
	providers *Providers,
) {
	appHandler := handler.NewAppHandler()
	routes.SetupAppRoutes(e, appHandler, tel, cfg)
	SetupCheckoutSession(ctx, cfg, e, appLogger, tel, providers)
	SetupCart(e, cfg, appLogger, tel, providers)
}
