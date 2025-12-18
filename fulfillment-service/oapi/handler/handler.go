// Package handler provides the OpenAPI server implementation.
package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/oapi"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/oapi/mapper"
)

// Ensure Handler implements ServerInterface.
var _ oapi.ServerInterface = (*Handler)(nil)

// Handler implements the generated ServerInterface.
type Handler struct {
	fulfillmentService service.FulfillmentService
}

// NewHandler creates a new Handler instance.
func NewHandler(fulfillmentService service.FulfillmentService) *Handler {
	return &Handler{fulfillmentService: fulfillmentService}
}

// CalculateShippingRates implements ServerInterface.
func (h *Handler) CalculateShippingRates(ctx echo.Context) error {
	var req oapi.CalculateShippingRatesRequest
	if err := ctx.Bind(&req); err != nil {
		return err
	}

	if err := ctx.Validate(&req); err != nil {
		return err
	}

	serviceReq := mapper.ToCalculateShippingRatesDTO(&req)

	rates, err := h.fulfillmentService.CalculateShippingRates(ctx.Request().Context(), serviceReq)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(ctx, rates)
}

// GetFulfillmentByOrderID implements ServerInterface.
func (h *Handler) GetFulfillmentByOrderID(ctx echo.Context, orderID oapi.OrderID) error {
	fulfillment, err := h.fulfillmentService.GetFulfillmentByOrderID(
		ctx.Request().Context(),
		orderID,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(ctx, fulfillment)
}

// UpdateFulfillmentStatus implements ServerInterface.
func (h *Handler) UpdateFulfillmentStatus(ctx echo.Context, orderID oapi.OrderID) error {
	var req oapi.UpdateFulfillmentStatusRequest
	if err := ctx.Bind(&req); err != nil {
		return err
	}

	if err := ctx.Validate(&req); err != nil {
		return err
	}

	serviceReq := dto.UpdateFulfillmentStatusRequest{
		Status: constant.FulfillmentStatus(req.Status),
	}

	fulfillment, err := h.fulfillmentService.UpdateFulfillmentStatusByOrderID(
		ctx.Request().Context(),
		orderID,
		serviceReq,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(ctx, fulfillment)
}
