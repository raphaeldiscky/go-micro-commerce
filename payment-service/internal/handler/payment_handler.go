// Package handler provides HTTP handlers for Payment operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/service"
)

// PaymentHandler handles HTTP requests for Payment operations.
type PaymentHandler struct {
	orderService service.PaymentServiceInterface
	logger       logger.Logger
}

// NewPaymentHandler creates a new instance of PaymentHandler.
func NewPaymentHandler(
	orderService service.PaymentServiceInterface,
	appLogger logger.Logger,
) *PaymentHandler {
	return &PaymentHandler{
		orderService: orderService,
		logger:       appLogger,
	}
}

// PayOrder handles POST /orders/pay/:orderID.
//
// Route: POST /orders/pay/:orderID
//
// Authentication: Requires user authentication.
func (h *PaymentHandler) PayOrder(c echo.Context) error {
	param := c.Param("orderID")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	req := dto.PaymentRequest{
		CustomerID:    echoutils.GetUserIDFromContext(c),
		CustomerEmail: echoutils.GetEmailFromContext(c),
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	order, err := h.orderService.PayPayment(c.Request().Context(), req, orderID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, order)
}
