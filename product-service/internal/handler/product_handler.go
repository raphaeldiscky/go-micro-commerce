// Package handler provides HTTP handlers for product operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/service"
)

// ProductHandler handles HTTP requests for product operations.
type ProductHandler struct {
	productService service.ProductService
}

// NewProductHandler creates a new instance of ProductHandler.
func NewProductHandler(
	productService service.ProductService,
) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProduct handles POST /products.
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	var req dto.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	product, err := h.productService.CreateProduct(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return echoutils.ResponseCreated(c, product)
}

// GetProduct handles GET /products/:productID.
func (h *ProductHandler) GetProduct(c echo.Context) error {
	param := c.Param("productID")

	productID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	product, err := h.productService.GetProduct(c.Request().Context(), productID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, product)
}

// GetProducts handles GET /products.
func (h *ProductHandler) GetProducts(c echo.Context) error {
	var req dto.GetProductsRequest

	req.Limit = pageutils.ParseQueryInt64(
		c,
		"limit",
		pkgconstant.DefaultLimit,
		pkgconstant.DefaultMinLimit,
		pkgconstant.DefaultMaxLimit,
	)
	req.Page = pageutils.ParseQueryInt64(
		c,
		"page",
		pkgconstant.DefaultPage,
		pkgconstant.DefaultMinPage,
		pkgconstant.DefaultMaxPage,
	)

	if err := c.Validate(&req); err != nil {
		return err
	}

	products, pagination, err := h.productService.GetProducts(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return echoutils.ResponseOKOffsetPagination(c, products, pagination)
}

// UpdateProduct handles PUT /products/:productID.
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	param := c.Param("productID")

	productID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	var reqBody dto.UpdateProductRequest
	if err = c.Bind(&reqBody); err != nil {
		return err
	}

	req := dto.UpdateProductRequest{
		ID:       productID,
		Name:     reqBody.Name,
		Price:    reqBody.Price,
		Quantity: reqBody.Quantity,
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	product, err := h.productService.UpdateProduct(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, product)
}

// DeleteProduct handles DELETE /products/:productID.
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	param := c.Param("productID")

	productID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	err = h.productService.DeleteProduct(c.Request().Context(), productID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOKPlain(c)
}
