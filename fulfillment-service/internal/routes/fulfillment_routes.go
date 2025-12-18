package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/middleware"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/oapi"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/oapi/handler"
)

// SetupFulfillmentRoutes registers API routes with middleware.
// Note: Health check is handled by AppHandler in app_routes.go.
func SetupFulfillmentRoutes(e *echo.Echo, h *handler.Handler) {
	wrapper := oapi.ServerInterfaceWrapper{Handler: h}

	// Protected routes (require authentication)
	protected := e.Group("")
	protected.Use(middleware.AuthMiddleware)
	protected.POST("/shipping-rates", wrapper.CalculateShippingRates)

	// Admin routes (require admin role)
	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
	admin.GET("/order/:orderID", wrapper.GetFulfillmentByOrderID)
	admin.PATCH("/order/:orderID/status", wrapper.UpdateFulfillmentStatus)
}
