package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/middleware"
)

// SetupSearchRoutes sets up search-related routes.
func SetupSearchRoutes(e *echo.Echo, searchHandler *handler.SearchHandler) {
	v1 := e.Group("/v1")

	public := v1.Group("")
	// Product search and management
	public.GET("/products", searchHandler.SearchProducts)
	public.GET("/product/:id", searchHandler.GetProduct)

	// Autocomplete and suggestions
	public.GET("/autocomplete", searchHandler.AutoComplete)
	public.GET("/suggestions", searchHandler.GetSuggestions)

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware)
	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
	admin.POST("/init-indices", searchHandler.InitializeIndices)
	admin.POST("/refresh-indices", searchHandler.RefreshIndices)
	admin.POST("/index/product", searchHandler.IndexProduct)
	admin.PUT("/index/product", searchHandler.UpdateProduct)
	admin.DELETE("/index/product/:id", searchHandler.DeleteProduct)
}
