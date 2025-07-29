// Package rest provides the REST API implementation for product management.
package rest

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/raphaeldiscky/go-ddd-template/internal/app/interfaces"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/http/dto/mapper"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/http/dto/request"
)

// ProductController handles HTTP requests related to products.
type ProductController struct {
	service interfaces.ProductService
}

// NewProductController initializes a new ProductController.
func NewProductController(e *echo.Echo, service interfaces.ProductService) *ProductController {
	controller := &ProductController{
		service: service,
	}

	e.POST("/api/v1/products", controller.CreateProductController)
	e.GET("/api/v1/products", controller.GetAllProductsController)
	e.GET("/api/v1/products/:id", controller.GetProductByIDController)
	e.Use(middleware.Recover())

	return controller
}

// CreateProductController handles the creation of a new product.
func (pc *ProductController) CreateProductController(c echo.Context) error {
	var createProductRequest request.CreateProductRequest

	if err := c.Bind(&createProductRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse request body",
		})
	}

	productCommand, err := createProductRequest.ToCreateProductCommand()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid product Id format",
		})
	}

	result, err := pc.service.CreateProduct(productCommand)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create product",
		})
	}

	response := mapper.ToProductResponse(result.Result)

	return c.JSON(http.StatusCreated, response)
}

// GetAllProductsController retrieves all products.
func (pc *ProductController) GetAllProductsController(c echo.Context) error {
	products, err := pc.service.FindAllProducts()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch products",
		})
	}

	response := mapper.ToProductListResponse(products.Result)

	return c.JSON(http.StatusOK, response)
}

// GetProductByIDController retrieves a product by ID.
func (pc *ProductController) GetProductByIDController(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid product Id format",
		})
	}

	product, err := pc.service.FindProductByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch product",
		})
	}

	if product == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Product not found",
		})
	}

	response := mapper.ToProductResponse(product.Result)

	return c.JSON(http.StatusOK, response)
}
