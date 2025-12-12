// Package routes provides the HTTP routes for the authentication service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/middleware"
)

// SetupAddressRoutes sets up all address routes.
// All routes are protected and require authentication.
func SetupAddressRoutes(e *echo.Echo, h *handler.AddressHandler) {
	// Address routes (all protected - user must be authenticated)
	protected := e.Group("/users/addresses")
	protected.Use(middleware.AuthMiddleware)

	// Create new address
	protected.POST("", h.CreateAddress)

	// List all user addresses (sorted by is_default DESC, created_at DESC)
	protected.GET("", h.ListAddresses)

	// Get default address
	protected.GET("/default", h.GetDefaultAddress)

	// Get specific address by ID
	protected.GET("/:id", h.GetAddress)

	// Update address
	protected.PUT("/:id", h.UpdateAddress)

	// Delete address
	protected.DELETE("/:id", h.DeleteAddress)

	// Set address as default
	protected.PATCH("/:id/default", h.SetDefaultAddress)
}
