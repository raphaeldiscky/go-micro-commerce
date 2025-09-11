package echoutils

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// GetXUserID retrieves the X-User-ID header from the context as UUID.
func GetXUserID(ctx echo.Context) (uuid.UUID, bool) {
	xUserID := ctx.Request().Header.Get(constant.XUserID)
	if xUserID == "" {
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(xUserID)
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}

// GetXEmail retrieves the X-Email header from the context.
func GetXEmail(ctx echo.Context) (string, bool) {
	xEmail := ctx.Request().Header.Get(constant.XEmail)
	if xEmail == "" {
		return "", false
	}

	return xEmail, true
}

// GetXRoles retrieves the X-Roles header from the context.
func GetXRoles(ctx echo.Context) ([]string, bool) {
	xRoles := ctx.Request().Header.Get(constant.XRoles)
	if xRoles == "" {
		return nil, false
	}

	// Handle single role or comma-separated roles
	roles := strings.Split(xRoles, ",")
	// Trim whitespace from each role
	for i, role := range roles {
		roles[i] = strings.TrimSpace(role)
	}

	return roles, true
}

// GetXIsActive retrieves the X-Is-Active header from the context.
func GetXIsActive(ctx echo.Context) (bool, bool) {
	xIsActive := ctx.Request().Header.Get(constant.XIsActive)
	if xIsActive == "" {
		return false, false
	}

	isActive, err := strconv.ParseBool(xIsActive)
	if err != nil {
		return false, false
	}

	return isActive, true
}

// GetXRequestID retrieves the X-RequestID header from the context.
func GetXRequestID(ctx echo.Context) (string, bool) {
	xRequestID := ctx.Request().Header.Get(constant.XRequestID)
	if xRequestID == "" {
		return "", false
	}

	return xRequestID, true
}
