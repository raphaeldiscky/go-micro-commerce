package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/app/dto"
	service "github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/app/service"
)

// ProductHandler handles HTTP requests for product operations
type ProductHandler struct {
	productService service.ProductServiceInterface
}

// NewProductHandler creates a new instance of ProductHandler
func NewProductHandler(productService service.ProductServiceInterface) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProduct handles POST /products
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	var req dto.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate request
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}
	if req.Price <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Price must be greater than 0"})
	}
	if req.SellerId == uuid.Nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Seller ID is required"})
	}

	product, err := h.productService.CreateProduct(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, product)
}

// GetProduct handles GET /products/:id
func (h *ProductHandler) GetProduct(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	product, err := h.productService.GetProduct(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, product)
}

// GetProducts handles GET /products
func (h *ProductHandler) GetProducts(c echo.Context) error {
	var req dto.GetProductsRequest

	// Parse query parameters
	limitParam := c.QueryParam("limit")
	if limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 {
			req.Limit = limit
		}
	}
	if req.Limit == 0 {
		req.Limit = 10 // Default limit
	}

	offsetParam := c.QueryParam("offset")
	if offsetParam != "" {
		if offset, err := strconv.Atoi(offsetParam); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	sellerIdParam := c.QueryParam("seller_id")
	if sellerIdParam != "" {
		if sellerId, err := uuid.Parse(sellerIdParam); err == nil {
			req.SellerId = &sellerId
		} else {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid seller ID"})
		}
	}

	products, err := h.productService.GetProducts(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, products)
}

// UpdateProduct handles PUT /products/:id
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	var reqBody struct {
		Name     string    `json:"name"`
		Price    float64   `json:"price"`
		SellerId uuid.UUID `json:"seller_id"`
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate request
	if reqBody.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}
	if reqBody.Price <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Price must be greater than 0"})
	}
	if reqBody.SellerId == uuid.Nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Seller ID is required"})
	}

	req := dto.UpdateProductRequest{
		Id:       id,
		Name:     reqBody.Name,
		Price:    reqBody.Price,
		SellerId: reqBody.SellerId,
	}

	product, err := h.productService.UpdateProduct(c.Request().Context(), req)
	if err != nil {
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, product)
}

// DeleteProduct handles DELETE /products/:id
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	err = h.productService.DeleteProduct(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// RegisterRoutes registers all product routes
func (h *ProductHandler) RegisterRoutes(e *echo.Echo) {
	productGroup := e.Group("/api/v1/products")

	productGroup.POST("", h.CreateProduct)
	productGroup.GET("", h.GetProducts)
	productGroup.GET("/:id", h.GetProduct)
	productGroup.PUT("/:id", h.UpdateProduct)
	productGroup.DELETE("/:id", h.DeleteProduct)
}
