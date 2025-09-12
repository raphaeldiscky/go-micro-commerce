// Package handler provides HTTP handlers for Payment operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// PaymentHandler handles HTTP requests for Payment operations.
type PaymentHandler struct {
	paymentService service.PaymentServiceInterface
	logger         logger.Logger
}

// NewPaymentHandler creates a new instance of PaymentHandler.
func NewPaymentHandler(
	paymentService service.PaymentServiceInterface,
	appLogger logger.Logger,
) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         appLogger,
	}
}

// ProcessPayment handles POST /payments/order/:orderID/process.
// Authentication: Requires user authentication.
func (h *PaymentHandler) ProcessPayment(c echo.Context) error {
	param := c.Param("orderID")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	req := dto.ProcessPaymentRequest{
		CustomerID:    echoutils.GetUserIDFromContext(c),
		CustomerEmail: echoutils.GetEmailFromContext(c),
	}

	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	payment, err := h.paymentService.ProcessPayment(c.Request().Context(), orderID, req)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, payment)
}

// GetPaymentByOrderID handles GET /payments/order/:orderID.
//
// Route: GET /payments/order/:orderID
//
// Authentication: Requires user authentication.
func (h *PaymentHandler) GetPaymentByOrderID(c echo.Context) error {
	param := c.Param("orderID")

	orderID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	payment, err := h.paymentService.GetPaymentByOrderID(c.Request().Context(), orderID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, payment)
}
