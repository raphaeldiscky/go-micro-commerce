// Package routes provides the HTTP routes for the cart service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/middleware"
)

// SetupCartRoutes sets up all cart routes.
func SetupCartRoutes(e *echo.Echo, h *handler.CartHandler) {
	v1 := e.Group("/v1/carts")

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)

	// User cart operations
	protected.GET("/me", h.GetMyCart)                 // Get my active cart
	protected.POST("", h.CreateCart)                  // Create cart
	protected.POST("/:cartID/items", h.AddItemToCart) // Add item to cart
	protected.DELETE(
		"/:cartID/items/:itemID",
		h.RemoveItemFromCart,
	) // Remove item from cart
	protected.PATCH(
		"/:cartID/items/:itemID/quantity",
		h.UpdateItemQuantity,
	) // Update item quantity
	protected.PATCH(
		"/:cartID/items/:itemID/select",
		h.SelectItemForCheckout,
	) // Select item for checkout

	// Admin routes
	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
	admin.GET("/:cartID", h.GetCartByID) // Admin: Get specific cart by ID
}
