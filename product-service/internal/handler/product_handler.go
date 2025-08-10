// Package handler provides HTTP handlers for product operations.
package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/service"
)

// ProductHandler handles HTTP requests for product operations.
type ProductHandler struct {
	productService service.ProductServiceInterface
	logger         logger.Logger
}

// NewProductHandler creates a new instance of ProductHandler.
func NewProductHandler(
	productService service.ProductServiceInterface,
	appLogger logger.Logger,
) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		logger:         appLogger,
	}
}

// CreateProduct handles POST /products.
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	var req dto.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		httpErr := httperror.NewInvalidRequestBodyError()

		return c.JSON(httpErr.GetCode(), httpErr)
	}

	// Validate request using go-playground validator
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	product, err := h.productService.CreateProduct(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, product)
}

// GetProduct handles GET /products/:id.
func (h *ProductHandler) GetProduct(c echo.Context) error {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		httpErr := httperror.NewInvalidProductIDError()

		return c.JSON(httpErr.GetCode(), httpErr)
	}

	product, err := h.productService.GetProduct(c.Request().Context(), id)
	if err != nil {
		if err.Error() == constant.ProductNotFoundErrorMessage {
			httpErr := httperror.NewProductNotFoundError()

			return c.JSON(httpErr.GetCode(), httpErr)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, product)
}

// GetProducts handles GET /products.
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

	// Validate query parameters using go-playground validator
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	products, err := h.productService.GetProducts(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, products)
}

// UpdateProduct handles PUT /products/:id.
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		httpErr := httperror.NewInvalidProductIDError()

		return c.JSON(httpErr.GetCode(), httpErr)
	}

	var reqBody dto.UpdateProductRequest

	if err := c.Bind(&reqBody); err != nil {
		httpErr := httperror.NewInvalidRequestBodyError()

		return c.JSON(httpErr.GetCode(), httpErr)
	}

	// Validate request using go-playground validator
	if err := c.Validate(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	req := dto.UpdateProductRequest{
		ID:       id,
		Name:     reqBody.Name,
		Price:    reqBody.Price,
		Quantity: reqBody.Quantity,
	}

	product, err := h.productService.UpdateProduct(c.Request().Context(), req)
	if err != nil {
		if err.Error() == constant.ProductNotFoundErrorMessage {
			httpErr := httperror.NewProductNotFoundError()

			return c.JSON(httpErr.GetCode(), httpErr)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, product)
}

// DeleteProduct handles DELETE /products/:id.
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		httpErr := httperror.NewInvalidProductIDError()

		return c.JSON(httpErr.GetCode(), httpErr)
	}

	err = h.productService.DeleteProduct(c.Request().Context(), id)
	if err != nil {
		if err.Error() == constant.ProductNotFoundErrorMessage {
			httpErr := httperror.NewProductNotFoundError()

			return c.JSON(httpErr.GetCode(), httpErr)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
