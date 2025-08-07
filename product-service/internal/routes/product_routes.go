// Package routes provides the HTTP routes for the product service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/handler"
)

// SetupProductRoutes sets up all product routes.
func SetupProductRoutes(e *echo.Echo, h *handler.ProductHandler) {
	// Health and readiness checks
	e.GET("/health", h.Health)

	v1 := e.Group("/api/v1")

	products := v1.Group("/products")
	products.POST("", h.CreateProduct)
	products.GET("", h.GetProducts)
	products.GET("/:id", h.GetProduct)
	products.PUT("/:id", h.UpdateProduct)
	products.DELETE("/:id", h.DeleteProduct)
}
