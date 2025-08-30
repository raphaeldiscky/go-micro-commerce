// Package routes provides the HTTP routes for the payment service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/middleware"
)

// SetupPaymentRoutes sets up all payment routes.
func SetupPaymentRoutes(e *echo.Echo, h *handler.PaymentHandler) {
	v1 := e.Group("/v1")

	protected := v1.Group("/payments")
	protected.Use(middleware.AuthMiddleware)

	// Process a payment
	protected.POST("/:paymentID/process", h.ProcessPayment)

	// Get payment by order ID
	protected.GET("/order/:orderID", h.GetPaymentByOrderID)
}
