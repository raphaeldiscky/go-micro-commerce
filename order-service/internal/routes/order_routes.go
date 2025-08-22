// Package routes provides the HTTP routes for the product service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/handler"
)

// SetupOrderRoutes sets up all order routes.
func SetupOrderRoutes(e *echo.Echo, h *handler.OrderHandler) {
	v1 := e.Group("/v1")

	orders := v1.Group("")
	orders.POST("", h.CreateOrder)
	orders.GET("", h.GetOrders)
	orders.GET("/:orderID", h.GetOrder)
	orders.GET("/customer/:customerID", h.GetOrdersByCustomer)
	orders.PUT("/:orderID", h.UpdateOrder)
	orders.POST("/cancel/:orderID", h.CancelOrder)
	orders.POST("/pay/:orderID", h.PayOrder)
}
