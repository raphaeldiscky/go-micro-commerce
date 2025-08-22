// Package handler provides HTTP handlers for product operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/pageutils"

	pkgConstant "github.com/raphaeldiscky/go-micro-template/pkg/constant"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/dto"
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
		pkgConstant.DefaultLimit,
		1,
		100,
	) // min=1, max=100
	req.Page = pageutils.ParseQueryInt64(
		c,
		"page",
		pkgConstant.DefaultPage,
		1,
		0,
	) // min=1, max=0 (no max)

	if err := c.Validate(&req); err != nil {
		return err
	}

	products, paging, err := h.productService.GetProducts(c.Request().Context(), req)
	if err != nil {
		return err
	}

	paging.Links = pageutils.NewLinks(
		c.Request(),
		paging.Page,
		paging.Size,
		paging.TotalPage,
	)

	return echoutils.ResponseOKPagination(c, products, paging)
}

// UpdateProduct handles PUT /products/:productID.
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	param := c.Param("productID")

	productID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	var reqBody dto.UpdateProductRequest
	if err := c.Bind(&reqBody); err != nil {
		return err
	}

	req := dto.UpdateProductRequest{
		ID:       productID,
		Name:     reqBody.Name,
		Price:    reqBody.Price,
		Quantity: reqBody.Quantity,
	}

	if err := c.Validate(&req); err != nil {
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
