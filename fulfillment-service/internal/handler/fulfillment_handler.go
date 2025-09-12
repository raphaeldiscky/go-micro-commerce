package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/service"
)

// FulfillmentHandler handles HTTP requests for Fulfillment operations.
type FulfillmentHandler struct {
	fulfillmentService service.FulfillmentServiceInterface
	logger             logger.Logger
}

// NewFulfillmentHandler creates a new instance of FulfillmentHandler.
func NewFulfillmentHandler(
	fulfillmentService service.FulfillmentServiceInterface,
	appLogger logger.Logger,
) *FulfillmentHandler {
	return &FulfillmentHandler{
		fulfillmentService: fulfillmentService,
		logger:             appLogger,
	}
}

// UpdateFulfillmentStatusByOrderID handles PUT /fulfillments/order/:orderID/status.
//
// Route: PUT /fulfillments/:fulfillmentID/status
//
// Authentication: Requires user authentication.
func (h *FulfillmentHandler) UpdateFulfillmentStatusByOrderID(c echo.Context) error {
	param := c.Param("orderID")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	var req dto.UpdateFulfillmentStatusRequest
	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	fulfillment, err := h.fulfillmentService.UpdateFulfillmentStatusByOrderID(
		c.Request().Context(),
		orderID,
		req,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, fulfillment)
}

// GetFulfillmentByOrderID handles GET /fulfillments/order/:orderID.
//
// Route: GET /fulfillments/order/:orderID
//
// Authentication: Requires user authentication.
func (h *FulfillmentHandler) GetFulfillmentByOrderID(c echo.Context) error {
	param := c.Param("orderID")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	fulfillment, err := h.fulfillmentService.GetFulfillmentByOrderID(c.Request().Context(), orderID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, fulfillment)
}

// CalculateShippingRates handles POST /fulfillments/shipping-rates.
//
// Route: POST /fulfillments
//
// Authentication: Requires user authentication.
func (h *FulfillmentHandler) CalculateShippingRates(c echo.Context) error {
	var req *dto.CalculateShippingRatesRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	rates, err := h.fulfillmentService.CalculateShippingRates(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, rates)
}
