// Package routes provides the HTTP routes for the product service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/handler"
)

// SetupProductRoutes sets up all product routes.
func SetupProductRoutes(e *echo.Echo, h *handler.ProductHandler) {
	v1 := e.Group("/v1")

	products := v1.Group("")
	products.POST("", h.CreateProduct)
	products.GET("", h.GetProducts)
	products.GET("/:productID", h.GetProduct)
	products.PUT("/:productID", h.UpdateProduct)
	products.DELETE("/:productID", h.DeleteProduct)
}
