// Package routes provides the HTTP routes for the product service.
package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/handler"
)

// SetupAppRoutes sets up all app routes.
func SetupAppRoutes(
	e *echo.Echo,
	app *handler.AppHandler,
	tel *telemetry.Telemetry,
	cfg *config.Config,
) {
	// Health and readiness checks
	e.GET("/health", app.Health)
	// Metrics endpoint
	if tel != nil {
		e.GET(cfg.Metrics.Path, tel.MetricsHandler())
	}

	e.RouteNotFound("/*", app.RouteNotFound)
}
