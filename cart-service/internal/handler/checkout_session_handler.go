// Package handler provides HTTP handlers for CheckoutSession operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
)

// CheckoutSessionHandler handles HTTP requests for CheckoutSession operations.
type CheckoutSessionHandler struct {
	checkoutSessionService service.CheckoutSessionService
	tel                    *telemetry.Telemetry
}

// NewCheckoutSessionHandler creates a new instance of CheckoutSessionHandler.
func NewCheckoutSessionHandler(
	checkoutSessionService service.CheckoutSessionService,
	tel *telemetry.Telemetry,
) *CheckoutSessionHandler {
	return &CheckoutSessionHandler{
		checkoutSessionService: checkoutSessionService,
		tel:                    tel,
	}
}

// CreateCheckoutSession handles POST /checkout-sessions.
func (h *CheckoutSessionHandler) CreateCheckoutSession(c echo.Context) error {
	ctx := echoutils.ContextWithUserInfo(c)

	ctx, end := h.tel.StartSpan(ctx, "CheckoutSessionHandler.CreateCheckoutSession")
	defer end()

	var req dto.CreateCheckoutSessionRequest

	if err := c.Bind(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	req.CustomerID = echoutils.GetUserIDFromContext(c)

	if err := c.Validate(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"http.route":      c.Path(),
		"http.method":     c.Request().Method,
		"customer.id":     req.CustomerID.String(),
		"idempotency.key": req.IdempotencyKey,
	})

	session, err := h.checkoutSessionService.CreateCheckoutSession(
		ctx,
		&req,
	)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"session.id":  session.ID,
		"cart.id":     session.CartID,
		"items.count": len(session.Items),
	})

	return echoutils.ResponseOK(c, session)
}

// GetCheckoutSessionByID retrieves a single checkout session by its ID.
//
// Route: GET /checkout-sessions/:sessionID
//
// Authentication: Requires user authentication.
func (h *CheckoutSessionHandler) GetCheckoutSessionByID(c echo.Context) error {
	ctx := echoutils.ContextWithUserInfo(c)

	ctx, end := h.tel.StartSpan(ctx, "CheckoutSessionHandler.GetCheckoutSessionByID")
	defer end()

	param := c.Param("sessionID")

	sessionID, err := uuid.Parse(param)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"http.route":  c.Path(),
		"http.method": c.Request().Method,
		"session.id":  sessionID.String(),
	})

	session, err := h.checkoutSessionService.GetCheckoutSession(
		ctx,
		sessionID,
	)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"customer.id":     session.CustomerID,
		"cart.id":         session.CartID,
		"items.count":     len(session.Items),
		"payment.gateway": session.PaymentGateway,
	})

	return echoutils.ResponseOK(c, session)
}

// UpdateCheckoutSession handles PATCH /checkout-sessions/:sessionID.
// Updates checkout session with address, carrier, or payment gateway.
func (h *CheckoutSessionHandler) UpdateCheckoutSession(c echo.Context) error {
	ctx := echoutils.ContextWithUserInfo(c)

	ctx, end := h.tel.StartSpan(ctx, "CheckoutSessionHandler.UpdateCheckoutSession")
	defer end()

	param := c.Param("sessionID")

	sessionID, err := uuid.Parse(param)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	var req dto.UpdateCheckoutSessionRequest

	if err = c.Bind(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	// Set customer info from JWT token
	req.CustomerID = echoutils.GetUserIDFromContext(c)

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"http.route":  c.Path(),
		"http.method": c.Request().Method,
		"session.id":  sessionID.String(),
		"customer.id": req.CustomerID.String(),
	})

	if req.PaymentGateway != nil {
		h.tel.AddSpanAttributes(ctx, map[string]any{
			"payment.gateway": *req.PaymentGateway,
		})
	}

	session, err := h.checkoutSessionService.UpdateCheckoutSession(
		ctx,
		sessionID,
		&req,
	)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"cart.id":     session.CartID,
		"items.count": len(session.Items),
	})

	return echoutils.ResponseOK(c, session)
}
