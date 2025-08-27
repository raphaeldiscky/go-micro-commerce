// Package routes provides the HTTP routes for the product service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/middleware"
)

// SetupPaymentRoutes sets up all order routes.
func SetupPaymentRoutes(e *echo.Echo, h *handler.PaymentHandler) {
	v1 := e.Group("/v1")

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)
	protected.POST("/pay/:orderID", h.PayOrder)
}
