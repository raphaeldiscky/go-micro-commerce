// Package handler provides HTTP handlers for Order operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/pageutils"

	pkgConstant "github.com/raphaeldiscky/go-micro-template/pkg/constant"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/service"
)

// OrderHandler handles HTTP requests for Order operations.
type OrderHandler struct {
	orderService service.OrderServiceInterface
	logger       logger.Logger
}

// NewOrderHandler creates a new instance of OrderHandler.
func NewOrderHandler(
	orderService service.OrderServiceInterface,
	appLogger logger.Logger,
) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		logger:       appLogger,
	}
}

// CreateOrder handles POST /orders.
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	req := dto.CreateOrderRequest{
		CustomerID:    echoutils.GetUserIDFromContext(c),
		CustomerEmail: echoutils.GetEmailFromContext(c),
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	order, err := h.orderService.CreateOrder(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return echoutils.ResponseCreated(c, order)
}

// GetOrder handles GET /orders/:orderID.
func (h *OrderHandler) GetOrder(c echo.Context) error {
	param := c.Param("orderID")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	order, err := h.orderService.GetOrder(c.Request().Context(), orderID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, order)
}

// GetOrdersByCustomer handles GET /orders/customer/:customerId.
func (h *OrderHandler) GetOrdersByCustomer(c echo.Context) error {
	var req dto.GetOrdersRequest

	param := c.Param("customerId")

	customerID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

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

	orders, paging, err := h.orderService.GetOrdersByCustomer(
		c.Request().Context(),
		customerID,
		req,
	)
	if err != nil {
		return err
	}

	paging.Links = pageutils.NewLinks(
		c.Request(),
		paging.Page,
		paging.Size,
		paging.TotalPage,
	)

	return echoutils.ResponseOKPagination(c, orders, paging)
}

// GetOrders handles GET /orders.
func (h *OrderHandler) GetOrders(c echo.Context) error {
	var req dto.GetOrdersRequest
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

	orders, paging, err := h.orderService.GetOrders(c.Request().Context(), req)
	if err != nil {
		return err
	}

	paging.Links = pageutils.NewLinks(
		c.Request(),
		paging.Page,
		paging.Size,
		paging.TotalPage,
	)

	return echoutils.ResponseOKPagination(c, orders, paging)
}

// UpdateOrderStatus handles PATCH /orders/:orderID/status.
func (h *OrderHandler) UpdateOrderStatus(c echo.Context) error {
	var req dto.UpdateOrderStatusRequest

	param := c.Param("orderID")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	order, err := h.orderService.UpdateOrderStatus(c.Request().Context(), orderID, req.Status)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, order)
}

// CancelOrder handles DELETE /orders/:orderID.
func (h *OrderHandler) CancelOrder(c echo.Context) error {
	param := c.Param("orderID")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	err = h.orderService.CancelOrder(c.Request().Context(), orderID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOKPlain(c)
}

// PayOrder handles POST /orders/:id/pay.
func (h *OrderHandler) PayOrder(c echo.Context) error {
	var req dto.PayOrderRequest

	param := c.Param("id")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	req.OrderID = orderID

	if err := c.Validate(&req); err != nil {
		return err
	}

	order, err := h.orderService.PayOrder(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, order)
}
