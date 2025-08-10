// Package provider provides HTTP client and server utilities.
package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/routes"
)

func SetupHTTP(cfg *config.Config, e *echo.Echo) {
	appHandler := handler.NewAppHandler()
	routes.SetupAppRoutes(e, appHandler)
	SetupProduct(cfg, e)
}
