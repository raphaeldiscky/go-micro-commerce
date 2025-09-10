// Package routes provides the HTTP routes for the search service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/handler"
)

// SetupAppRoutes sets up all app routes.
func SetupAppRoutes(e *echo.Echo, app *handler.AppHandler) {
	// Health and readiness checks
	e.GET("/health", app.Health)
	e.RouteNotFound("/*", app.RouteNotFound)
}
