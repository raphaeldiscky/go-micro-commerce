// Package routes provides the HTTP routes for the product service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/middleware"
)

// SetupOrderRoutes sets up all order routes.
func SetupOrderRoutes(e *echo.Echo, h *handler.OrderHandler) {
	v1 := e.Group("/v1")

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)
	protected.POST("", h.CreateOrder)
	protected.POST("/proto", h.CreateOrderWithProto)
	protected.GET("/user", h.GetLoggedInOrders)
	protected.POST("/cancel/:orderID", h.CancelOrder)
	protected.POST("/payment-request/:orderID", h.RequestPaymentOrder)

	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
	admin.GET("", h.GetOrders)
	admin.GET("/:orderID", h.GetOrderByID)
	admin.GET("/customer/:customerID", h.GetOrdersByCustomer)
}
