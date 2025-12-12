// Package routes provides the HTTP routes for the product service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/middleware"
)

// SetupOrderRoutes sets up all order routes.
func SetupOrderRoutes(e *echo.Echo, h *handler.OrderHandler) {
	protected := e.Group("")
	protected.Use(middleware.AuthMiddleware)
	protected.POST("/place-order", h.PlaceOrder)
	protected.POST("/saga", h.CreateOrderWithSaga)
	protected.POST("/temporal", h.CreateOrderWithTemporal)
	protected.GET("/user", h.GetLoggedInOrders)
	protected.POST("/cancel/:orderID", h.CancelOrder)
	protected.POST("/payment-request/:orderID", h.RequestPaymentOrder)

	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
	admin.GET("", h.GetOrders)
	admin.GET("/:orderID", h.GetOrderByID)
	admin.GET("/customer/:customerID", h.GetOrdersByCustomer)
}
