package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/middleware"
)

// SetupFulfillmentRoutes sets up all fulfillment routes.
func SetupFulfillmentRoutes(e *echo.Echo, h *handler.FulfillmentHandler) {
	v1 := e.Group("/v1")

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)

	protected.POST("/shipping-rates", h.CalculateShippingRates)
	// Get fulfillment by order ID
	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
	admin.GET("/order/:orderID", h.GetFulfillmentByOrderID)
	admin.PATCH("/order/:orderID/status", h.UpdateFulfillmentStatusByOrderID)
}
