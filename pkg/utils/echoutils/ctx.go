// Package echoutils provides utility functions for working with Echo context.
package echoutils

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
)

// GetUserID retrieves the user ID (UUID) from the context safely.
func GetUserID(ctx echo.Context) (userID string, ok bool) {
	val, ok := ctx.Get(constant.CtxUserID).(string)

	return val, ok
}

// GetEmail retrieves the user email from the context safely.
func GetEmail(ctx echo.Context) (email string, ok bool) {
	val, ok := ctx.Get(constant.CtxEmail).(string)

	return val, ok
}

// GetRoles retrieves the user roles from the context safely.
func GetRoles(ctx echo.Context) (roles []string, ok bool) {
	val, ok := ctx.Get(constant.CtxRoles).([]string)

	return val, ok
}

// IsActive checks if the user is active based on the context safely.
func IsActive(ctx echo.Context) (isActive, ok bool) {
	val, ok := ctx.Get(constant.CtxIsActive).(bool)

	return val, ok
}
