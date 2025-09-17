// Package handler provides HTTP handlers for Order operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
)

// OrderHandler handles HTTP requests for Order operations.
type OrderHandler struct {
	orderService service.OrderService
}

// NewOrderHandler creates a new instance of OrderHandler.
func NewOrderHandler(
	orderService service.OrderService,
) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrderWithSaga handles POST /orders/saga with saga pattern processing.
func (h *OrderHandler) CreateOrderWithSaga(c echo.Context) error {
	req := &dto.CreateOrderRequest{
		CustomerID:    echoutils.GetUserIDFromContext(c),
		CustomerEmail: echoutils.GetEmailFromContext(c),
	}

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := echoutils.ContextWithUserInfo(c)

	order, err := h.orderService.CreateOrderWithSaga(ctx, req)
	if err != nil {
		return err
	}

	if order.Status == constant.OrderStatusProcessing {
		mapped := mapper.MapToOrderSagaResponse(order)

		return echoutils.ResponseCreated(c, mapped)
	}

	return echoutils.ResponseCreated(c, order)
}

// CreateOrderWithTemporal handles POST /orders/temporal with Temporal processing.
func (h *OrderHandler) CreateOrderWithTemporal(c echo.Context) error {
	req := &dto.CreateOrderRequest{
		CustomerID:    echoutils.GetUserIDFromContext(c),
		CustomerEmail: echoutils.GetEmailFromContext(c),
	}

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := echoutils.ContextWithUserInfo(c)

	order, err := h.orderService.CreateOrderWithTemporal(ctx, req)
	if err != nil {
		return err
	}

	if order.Status == constant.OrderStatusPending {
		mapped := mapper.MapToOrderSagaResponse(order)

		return echoutils.ResponseCreated(c, mapped)
	}

	return echoutils.ResponseCreated(c, order)
}

// GetOrderByID retrieves a single order by its ID.
//
// Route: GET /orders/:orderID
//
// Authentication: Requires admin privileges.
func (h *OrderHandler) GetOrderByID(c echo.Context) error {
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

// GetOrdersByCustomer retrieves a list of orders by customer ID.
//
// Route: GET /orders/customer/:customerId
//
// Authentication: Requires admin privileges.
func (h *OrderHandler) GetOrdersByCustomer(c echo.Context) error {
	var req dto.GetOrdersRequest

	param := c.Param("customerID")

	customerID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

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

	if err = c.Validate(&req); err != nil {
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

// GetOrders retrieves a list of orders with pagination.
//
// Route: GET /orders
//
// Authentication: Requires admin privileges.
func (h *OrderHandler) GetOrders(c echo.Context) error {
	var req dto.GetOrdersRequest

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

// GetLoggedInOrders retrieves a list of orders for the logged-in user with pagination.
//
// Route: GET /orders/user
//
// Authentication: Requires user authentication.
func (h *OrderHandler) GetLoggedInOrders(c echo.Context) error {
	var req dto.GetOrdersRequest

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

	customerID := echoutils.GetUserIDFromContext(c)

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

// CancelOrder handles DELETE /orders/:orderID.
func (h *OrderHandler) CancelOrder(c echo.Context) error {
	param := c.Param("orderID")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	req := &dto.CancelOrderRequest{
		CustomerID:    echoutils.GetUserIDFromContext(c),
		CustomerEmail: echoutils.GetEmailFromContext(c),
		OrderID:       orderID,
	}

	err = h.orderService.CancelOrder(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return echoutils.ResponseOKPlain(c)
}

// RequestPaymentOrder handles POST /orders/pay/:orderID.
//
// Route: POST /orders/pay/:orderID
//
// Authentication: Requires user authentication.
func (h *OrderHandler) RequestPaymentOrder(c echo.Context) error {
	param := c.Param("orderID")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	req := dto.PayOrderRequest{
		CustomerID:    echoutils.GetUserIDFromContext(c),
		CustomerEmail: echoutils.GetEmailFromContext(c),
	}

	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	order, err := h.orderService.RequestPaymentOrder(c.Request().Context(), req, orderID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, order)
}
