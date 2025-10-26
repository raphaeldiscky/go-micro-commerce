// Package handler provides HTTP handlers for CheckoutSession operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
)

// CheckoutSessionHandler handles HTTP requests for CheckoutSession operations.
type CheckoutSessionHandler struct {
	checkoutSessionService service.CheckoutSessionService
}

// NewCheckoutSessionHandler creates a new instance of CheckoutSessionHandler.
func NewCheckoutSessionHandler(
	checkoutSessionService service.CheckoutSessionService,
) *CheckoutSessionHandler {
	return &CheckoutSessionHandler{
		checkoutSessionService: checkoutSessionService,
	}
}

// CreateCheckoutSession handles POST /checkout-sessions.
func (h *CheckoutSessionHandler) CreateCheckoutSession(c echo.Context) error {
	var req dto.CreateCheckoutSessionRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	req.CustomerID = echoutils.GetUserIDFromContext(c)

	if err := c.Validate(&req); err != nil {
		return err
	}

	session, err := h.checkoutSessionService.CreateCheckoutSession(
		echoutils.ContextWithUserInfo(c),
		&req,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, session)
}

// GetCheckoutSessionByID retrieves a single checkout session by its ID.
//
// Route: GET /checkout-sessions/:sessionID
//
// Authentication: Requires user authentication.
func (h *CheckoutSessionHandler) GetCheckoutSessionByID(c echo.Context) error {
	param := c.Param("sessionID")

	sessionID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	session, err := h.checkoutSessionService.GetCheckoutSession(
		echoutils.ContextWithUserInfo(c),
		sessionID,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, session)
}

// UpdateCheckoutSession handles PATCH /checkout-sessions/:sessionID.
// Updates checkout session with address, carrier, or payment gateway.
func (h *CheckoutSessionHandler) UpdateCheckoutSession(c echo.Context) error {
	param := c.Param("sessionID")

	sessionID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	var req dto.UpdateCheckoutSessionRequest

	if err = c.Bind(&req); err != nil {
		return err
	}

	// Set customer info from JWT token
	req.CustomerID = echoutils.GetUserIDFromContext(c)

	session, err := h.checkoutSessionService.UpdateCheckoutSession(
		echoutils.ContextWithUserInfo(c),
		sessionID,
		&req,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, session)
}

// PlaceOrder handles POST /checkout-sessions/:sessionID/place-order.
func (h *CheckoutSessionHandler) PlaceOrder(c echo.Context) error {
	param := c.Param("sessionID")

	sessionID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	var req dto.PlaceOrderRequest

	req.CheckoutSessionID = sessionID
	if err = c.Bind(&req); err != nil {
		return err
	}

	// Set customer info from JWT token
	req.CustomerID = echoutils.GetUserIDFromContext(c)

	if err = c.Validate(&req); err != nil {
		return err
	}

	session, err := h.checkoutSessionService.PlaceOrder(
		echoutils.ContextWithUserInfo(c),
		&req,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, session)
}
