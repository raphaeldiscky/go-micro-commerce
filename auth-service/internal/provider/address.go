package provider

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/service"
)

// SetupAddress initializes the address-related components.
func SetupAddress(
	_ context.Context,
	_ *config.Config,
	e *echo.Echo,
	appLogger logger.Logger,
	providers *Providers,
) {
	// Initialize address service
	addressService := service.NewAddressService(
		providers.DataStore,
		appLogger,
	)

	// Initialize address handler
	addressHandler := handler.NewAddressHandler(addressService)

	// Setup address routes
	routes.SetupAddressRoutes(e, addressHandler)
}
