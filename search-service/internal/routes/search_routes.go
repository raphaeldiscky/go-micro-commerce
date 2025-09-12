package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/handler"
)

// SetupSearchRoutes sets up search-related routes.
func SetupSearchRoutes(e *echo.Echo, searchHandler *handler.SearchHandler) {
	// Search routes
	search := e.Group("/search")

	// Product search and management
	search.GET("/products", searchHandler.SearchProducts)
	search.GET("/product/:id", searchHandler.GetProduct)
	search.POST("/index/product", searchHandler.IndexProduct)
	search.PUT("/index/product", searchHandler.UpdateProduct)
	search.DELETE("/index/product/:id", searchHandler.DeleteProduct)

	// Autocomplete and suggestions
	search.GET("/autocomplete", searchHandler.AutoComplete)
	search.GET("/suggestions", searchHandler.GetSuggestions)

	// Admin routes
	admin := search.Group("/admin")
	admin.POST("/init-indices", searchHandler.InitializeIndices)
	admin.POST("/refresh-indices", searchHandler.RefreshIndices)
}
