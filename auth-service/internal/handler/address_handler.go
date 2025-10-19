// Package handler provides HTTP handlers for the auth service.
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/service"
)

// AddressHandler handles HTTP requests for user addresses.
type AddressHandler struct {
	addressService service.AddressService
}

// NewAddressHandler creates a new address handler.
func NewAddressHandler(addressService service.AddressService) *AddressHandler {
	return &AddressHandler{
		addressService: addressService,
	}
}

// CreateAddress handles address creation.
func (h *AddressHandler) CreateAddress(c echo.Context) error {
	// Get authenticated user ID from context
	userID := echoutils.GetUserIDFromContext(c)

	var req dto.CreateAddressRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	response, err := h.addressService.CreateAddress(c.Request().Context(), userID, &req)
	if err != nil {
		return err
	}

	return echoutils.ResponseCreated(c, response)
}

// ListAddresses handles listing all addresses for the authenticated user.
func (h *AddressHandler) ListAddresses(c echo.Context) error {
	// Get authenticated user ID from context
	userID := echoutils.GetUserIDFromContext(c)

	limit := pageutils.ParseQueryInt64(
		c,
		"limit",
		pkgconstant.DefaultLimit,
		pkgconstant.DefaultMinLimit,
		pkgconstant.DefaultMaxLimit,
	)

	nextCursor := c.QueryParam("next_cursor")

	responses, pagination, err := h.addressService.ListUserAddresses(
		c.Request().Context(),
		userID,
		limit,
		nextCursor,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOKCursorPagination(c, responses, pagination)
}

// GetAddress handles getting a single address by ID.
func (h *AddressHandler) GetAddress(c echo.Context) error {
	// Get authenticated user ID from context
	userID := echoutils.GetUserIDFromContext(c)

	// Get address ID from path parameter
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest
	}

	response, err := h.addressService.GetAddress(c.Request().Context(), userID, addressID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, response)
}

// GetDefaultAddress handles getting the default address for the authenticated user.
func (h *AddressHandler) GetDefaultAddress(c echo.Context) error {
	// Get authenticated user ID from context
	userID := echoutils.GetUserIDFromContext(c)

	response, err := h.addressService.GetDefaultAddress(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, response)
}

// UpdateAddress handles updating an address.
func (h *AddressHandler) UpdateAddress(c echo.Context) error {
	// Get authenticated user ID from context
	userID := echoutils.GetUserIDFromContext(c)

	// Get address ID from path parameter
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest
	}

	var req dto.UpdateAddressRequest
	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	response, err := h.addressService.UpdateAddress(c.Request().Context(), userID, addressID, &req)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, response)
}

// DeleteAddress handles deleting an address.
func (h *AddressHandler) DeleteAddress(c echo.Context) error {
	// Get authenticated user ID from context
	userID := echoutils.GetUserIDFromContext(c)

	// Get address ID from path parameter
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest
	}

	if err = h.addressService.DeleteAddress(c.Request().Context(), userID, addressID); err != nil {
		return err
	}

	return echoutils.ResponseOKPlain(c)
}

// SetDefaultAddress handles setting an address as default.
func (h *AddressHandler) SetDefaultAddress(c echo.Context) error {
	// Get authenticated user ID from context
	userID := echoutils.GetUserIDFromContext(c)

	// Get address ID from path parameter
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest
	}

	response, err := h.addressService.SetDefaultAddress(c.Request().Context(), userID, addressID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, response)
}
