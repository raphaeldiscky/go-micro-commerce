// Package handler provides HTTP handlers for product operations.
package handler

import (
	"strconv"

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

	limitParam := c.QueryParam("limit")
	if limitParam != "" {
		if limit, err := strconv.ParseInt(limitParam, 10, 64); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if req.Limit == 0 {
		req.Limit = pkgConstant.DefaultLimit
	}

	pageParam := c.QueryParam("page")
	if pageParam != "" {
		if page, err := strconv.ParseInt(pageParam, 10, 64); err == nil && page > 0 {
			req.Page = page
		}
	}

	if req.Page == 0 {
		req.Page = pkgConstant.DefaultPage
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	products, paging, err := h.productService.GetProducts(c.Request().Context(), req)
	if err != nil {
		return err
	}

	paging.Links = pageutils.NewLinks(
		c.Request(),
		int(paging.Page),
		int(paging.Size),
		int(paging.TotalPage),
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

	if err := c.Validate(&reqBody); err != nil {
		return err
	}

	req := dto.UpdateProductRequest{
		ID:       productID,
		Name:     reqBody.Name,
		Price:    reqBody.Price,
		Quantity: reqBody.Quantity,
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
