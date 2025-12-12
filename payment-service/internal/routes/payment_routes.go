// Package routes provides the HTTP routes for the payment service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/middleware"
)

// SetupPaymentRoutes sets up all payment routes.
func SetupPaymentRoutes(e *echo.Echo, h *handler.PaymentHandler) {
	protected := e.Group("")
	protected.Use(middleware.AuthMiddleware)

	// Process a payment
	protected.POST("/order/:orderID/process", h.ProcessPayment)

	// Get payment by order ID (for payment page - users can see their own payments)
	protected.GET("/order/:orderID", h.GetPaymentByOrderID)

	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
}
