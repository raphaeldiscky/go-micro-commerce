// Package handler provides HTTP handlers for Cart operations.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
)

// CartHandler handles HTTP requests for Cart operations.
type CartHandler struct {
	cartService service.CartService
}

// NewCartHandler creates a new instance of CartHandler.
func NewCartHandler(
	cartService service.CartService,
) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// GetCartByID retrieves a single cart by its ID.
//
// Route: GET /carts/:cartID
//
// Authentication: Requires admin privileges.
func (h *CartHandler) GetCartByID(c echo.Context) error {
	param := c.Param("cartID")

	cartID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	cart, err := h.cartService.GetCart(c.Request().Context(), cartID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, cart)
}

// GetMyCart retrieves the logged-in user's active cart.
//
// Route: GET /carts/me
//
// Authentication: Requires user authentication.
func (h *CartHandler) GetMyCart(c echo.Context) error {
	customerID := echoutils.GetUserIDFromContext(c)

	cart, err := h.cartService.GetCartByUserID(c.Request().Context(), customerID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, cart)
}

// CreateCart creates a new cart for the logged-in user.
//
// Route: POST /carts
//
// Authentication: Requires user authentication.
func (h *CartHandler) CreateCart(c echo.Context) error {
	var req dto.CreateCartRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	req.CustomerID = echoutils.GetUserIDFromContext(c)

	cart, err := h.cartService.CreateCart(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	return echoutils.ResponseCreated(c, cart)
}

// AddItemToCart adds an item to the user's cart.
//
// Route: POST /carts/:cartID/items
//
// Authentication: Requires user authentication.
func (h *CartHandler) AddItemToCart(c echo.Context) error {
	param := c.Param("cartID")

	cartID, err := uuid.Parse(param)
	if err != nil {
		return err
	}

	var req dto.AddCartItemRequest

	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	cart, errAdd := h.cartService.AddItemToCart(c.Request().Context(), cartID, &req)
	if errAdd != nil {
		return errAdd
	}

	return echoutils.ResponseOK(c, cart)
}

// RemoveItemFromCart removes an item from the user's cart.
//
// Route: DELETE /carts/:cartID/items/:itemID
//
// Authentication: Requires user authentication.
func (h *CartHandler) RemoveItemFromCart(c echo.Context) error {
	cartParam := c.Param("cartID")

	cartID, err := uuid.Parse(cartParam)
	if err != nil {
		return err
	}

	itemParam := c.Param("itemID")

	itemID, err := uuid.Parse(itemParam)
	if err != nil {
		return err
	}

	cart, err := h.cartService.RemoveItemFromCart(c.Request().Context(), cartID, itemID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, cart)
}

// UpdateItemQuantity updates the quantity of a cart item.
//
// Route: PATCH /carts/:cartID/items/:itemID/quantity
//
// Authentication: Requires user authentication.
func (h *CartHandler) UpdateItemQuantity(c echo.Context) error {
	cartParam := c.Param("cartID")

	cartID, err := uuid.Parse(cartParam)
	if err != nil {
		return err
	}

	itemParam := c.Param("itemID")

	itemID, err := uuid.Parse(itemParam)
	if err != nil {
		return err
	}

	var req dto.UpdateCartItemQuantityRequest

	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	cart, errUpdate := h.cartService.UpdateItemQuantity(
		c.Request().Context(),
		cartID,
		itemID,
		req.Quantity,
	)
	if errUpdate != nil {
		return errUpdate
	}

	return echoutils.ResponseOK(c, cart)
}

// SelectItemForCheckout marks an item as selected for checkout.
//
// Route: PATCH /carts/:cartID/items/:itemID/select
//
// Authentication: Requires user authentication.
func (h *CartHandler) SelectItemForCheckout(c echo.Context) error {
	cartParam := c.Param("cartID")

	cartID, err := uuid.Parse(cartParam)
	if err != nil {
		return err
	}

	itemParam := c.Param("itemID")

	itemID, errParse := uuid.Parse(itemParam)
	if errParse != nil {
		return errParse
	}

	var req dto.SelectItemForCheckoutRequest

	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	cart, err := h.cartService.SelectItemForCheckout(
		c.Request().Context(),
		cartID,
		itemID,
		req.Selected,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, cart)
}
