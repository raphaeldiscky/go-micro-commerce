// Package handler provides HTTP handlers for Cart operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
)

// CartHandler handles HTTP requests for Cart operations.
type CartHandler struct {
	cartService service.CartService
	tel         *telemetry.Telemetry
}

// NewCartHandler creates a new instance of CartHandler.
func NewCartHandler(
	cartService service.CartService,
	tel *telemetry.Telemetry,
) *CartHandler {
	return &CartHandler{
		cartService: cartService,
		tel:         tel,
	}
}

// GetCartByID retrieves a single cart by its ID.
//
// Route: GET /carts/:cartID
//
// Authentication: Requires admin privileges.
func (h *CartHandler) GetCartByID(c echo.Context) error {
	ctx, end := h.tel.StartSpan(c.Request().Context(), "CartHandler.GetCartByID")
	defer end()

	param := c.Param("cartID")

	cartID, err := uuid.Parse(param)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"http.route":  c.Path(),
		"http.method": c.Request().Method,
		"cart.id":     cartID.String(),
	})

	cart, err := h.cartService.GetCart(ctx, cartID)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"customer.id": cart.CustomerID,
		"items.count": len(cart.Items),
	})

	return echoutils.ResponseOK(c, cart)
}

// GetMyCart retrieves the logged-in user's active cart.
//
// Route: GET /carts/me
//
// Authentication: Requires user authentication.
func (h *CartHandler) GetMyCart(c echo.Context) error {
	ctx, end := h.tel.StartSpan(c.Request().Context(), "CartHandler.GetMyCart")
	defer end()

	customerID := echoutils.GetUserIDFromContext(c)

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"http.route":  c.Path(),
		"http.method": c.Request().Method,
		"customer.id": customerID.String(),
	})

	cart, err := h.cartService.GetUserActiveCart(ctx, customerID)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"cart.id":     cart.ID,
		"items.count": len(cart.Items),
	})

	return echoutils.ResponseOK(c, cart)
}

// CreateCart creates a new cart for the logged-in user.
//
// Route: POST /carts
//
// Authentication: Requires user authentication.
func (h *CartHandler) CreateCart(c echo.Context) error {
	ctx, end := h.tel.StartSpan(c.Request().Context(), "CartHandler.CreateCart")
	defer end()

	var req dto.CreateCartRequest

	if err := c.Bind(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	if err := c.Validate(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	req.CustomerID = echoutils.GetUserIDFromContext(c)

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"http.route":  c.Path(),
		"http.method": c.Request().Method,
		"customer.id": req.CustomerID.String(),
	})

	cart, err := h.cartService.CreateCart(ctx, &req)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"cart.id": cart.ID,
	})

	return echoutils.ResponseCreated(c, cart)
}

// AddItemToActiveCart adds an item to the user's active cart.
//
// Route: POST /carts/:cartID/items
//
// Authentication: Requires user authentication.
func (h *CartHandler) AddItemToActiveCart(c echo.Context) error {
	ctx, end := h.tel.StartSpan(c.Request().Context(), "CartHandler.AddItemToActiveCart")
	defer end()

	var req dto.AddCartItemRequest

	req.CustomerID = echoutils.GetUserIDFromContext(c)

	if err := c.Bind(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	if err := c.Validate(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"http.route":  c.Path(),
		"http.method": c.Request().Method,
		"customer.id": req.CustomerID.String(),
		"product.id":  req.ProductID.String(),
		"quantity":    req.Quantity,
	})

	cart, errAdd := h.cartService.AddItemToActiveCart(ctx, &req)
	if errAdd != nil {
		h.tel.SetSpanError(ctx, errAdd)
		return errAdd
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"cart.id":     cart.ID,
		"items.count": len(cart.Items),
	})

	return echoutils.ResponseOK(c, cart)
}

// RemoveItemFromCart removes an item from the user's active cart.
//
// Route: DELETE /items/:itemID
//
// Authentication: Requires user authentication.
func (h *CartHandler) RemoveItemFromCart(c echo.Context) error {
	ctx, end := h.tel.StartSpan(c.Request().Context(), "CartHandler.RemoveItemFromCart")
	defer end()

	customerID := echoutils.GetUserIDFromContext(c)

	itemParam := c.Param("itemID")

	itemID, err := uuid.Parse(itemParam)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"http.route":  c.Path(),
		"http.method": c.Request().Method,
		"customer.id": customerID.String(),
		"item.id":     itemID.String(),
	})

	cart, err := h.cartService.RemoveItemFromActiveCart(ctx, customerID, itemID)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"cart.id":     cart.ID,
		"items.count": len(cart.Items),
	})

	return echoutils.ResponseOK(c, cart)
}

// UpdateItemQuantity updates the quantity of a cart item.
//
// Route: PATCH /items/:itemID/quantity
//
// Authentication: Requires user authentication.
func (h *CartHandler) UpdateItemQuantity(c echo.Context) error {
	ctx, end := h.tel.StartSpan(c.Request().Context(), "CartHandler.UpdateItemQuantity")
	defer end()

	customerID := echoutils.GetUserIDFromContext(c)

	itemParam := c.Param("itemID")

	itemID, err := uuid.Parse(itemParam)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	var req dto.UpdateCartItemQuantityRequest

	if err = c.Bind(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	if err = c.Validate(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"http.route":  c.Path(),
		"http.method": c.Request().Method,
		"customer.id": customerID.String(),
		"item.id":     itemID.String(),
		"quantity":    req.Quantity,
	})

	cart, errUpdate := h.cartService.UpdateActiveCartItemQuantity(
		ctx,
		customerID,
		itemID,
		req.Quantity,
	)
	if errUpdate != nil {
		h.tel.SetSpanError(ctx, errUpdate)
		return errUpdate
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"cart.id":     cart.ID,
		"items.count": len(cart.Items),
	})

	return echoutils.ResponseOK(c, cart)
}

// SelectItemForCheckout marks an item as selected for checkout.
//
// Route: PATCH /items/:itemID/select
//
// Authentication: Requires user authentication.
func (h *CartHandler) SelectItemForCheckout(c echo.Context) error {
	ctx, end := h.tel.StartSpan(c.Request().Context(), "CartHandler.SelectItemForCheckout")
	defer end()

	customerID := echoutils.GetUserIDFromContext(c)

	itemParam := c.Param("itemID")

	itemID, err := uuid.Parse(itemParam)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	var req dto.SelectItemForCheckoutRequest

	if err = c.Bind(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	if err = c.Validate(&req); err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"http.route":  c.Path(),
		"http.method": c.Request().Method,
		"customer.id": customerID.String(),
		"item.id":     itemID.String(),
		"selected":    req.Selected,
	})

	cart, err := h.cartService.SelectActiveCartItemForCheckout(
		ctx,
		customerID,
		itemID,
		req.Selected,
	)
	if err != nil {
		h.tel.SetSpanError(ctx, err)
		return err
	}

	h.tel.AddSpanAttributes(ctx, map[string]any{
		"cart.id":     cart.ID,
		"items.count": len(cart.Items),
	})

	return echoutils.ResponseOK(c, cart)
}
