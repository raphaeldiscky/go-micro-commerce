// Package routes provides the HTTP routes for the product service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/middleware"
)

// SetupProductRoutes sets up all product routes.
func SetupProductRoutes(e *echo.Echo, h *handler.ProductHandler) {
	v1 := e.Group("/v1")

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)
	protected.GET("", h.ListProducts)
	protected.GET("/:productID", h.GetProduct)

	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
	admin.POST("", h.CreateProduct)
	admin.GET("/:productID", h.GetProduct)
	admin.PUT("/:productID", h.UpdateProduct)
	admin.DELETE("/:productID", h.DeleteProduct)
}
