// Package routes provides the HTTP routes for the checkout session service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/middleware"
)

// SetupCheckoutSessionRoutes sets up all checkout session routes.
func SetupCheckoutSessionRoutes(e *echo.Echo, h *handler.CheckoutSessionHandler) {
	v1 := e.Group("/v1/checkout")

	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware)
	protected.POST("", h.CreateCheckoutSession)
	protected.GET("/:sessionID", h.GetCheckoutSessionByID)
	protected.POST("/:sessionID/place-order", h.PlaceOrder)
}
