package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
)

// SetupCart initializes the cart-related routes and services.
func SetupCart(
	e *echo.Echo,
	cfg *config.Config,
	appLogger logger.Logger,
	tel *telemetry.Telemetry,
	providers *Providers,
) {
	cartService := service.NewCartService(
		providers.DataStore,
		appLogger,
		tel,
	)
	providers.CartService = cartService
	cartHandler := handler.NewCartHandler(cartService, tel)

	routes.SetupCartRoutes(e, cartHandler)

	graphResolver := SetupGraphQLResolver(cartService, providers.CheckoutSessionService)
	routes.SetupGraphQLRoutes(e, cfg, graphResolver, appLogger)
}
