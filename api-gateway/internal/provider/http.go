// Package provider provides HTTP client and server utilities.
package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/gateway"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/routes"
)

// SetupHTTP initializes the HTTP server routes and middleware.
func SetupHTTP(
	e *echo.Echo,
	tel *telemetry.Telemetry,
	gw *gateway.Gateway,
	cfg *config.Config,
	providers *Providers,
) {
	appHandler := handler.NewAppHandler()

	routes.SetupGatewayRoutes(e, tel, gw, providers.authMiddleware)
	routes.SetupAppRoutes(e, appHandler, tel, cfg)
}
