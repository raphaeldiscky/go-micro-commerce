package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/middleware"
)

// SetupFulfillmentRoutes sets up all fulfillment routes.
func SetupFulfillmentRoutes(e *echo.Echo, h *handler.FulfillmentHandler) {
	v1 := e.Group("/v1")

	protected := v1.Group("/fulfillments")
	protected.Use(middleware.AuthMiddleware)

	// Update fulfillment status
	protected.PUT("/:fulfillmentID/status", h.UpdateFulfillmentStatus)

	// Get fulfillment by order ID
	protected.GET("/order/:orderID", h.GetFulfillmentByOrderID)
}
