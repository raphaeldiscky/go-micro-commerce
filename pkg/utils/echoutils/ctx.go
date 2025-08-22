// Package echoutils provides utility functions for working with Echo context.
package echoutils

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
)

// GetUserIDFromContext retrieves the user ID (UUID) from the context safely.
func GetUserIDFromContext(ctx echo.Context) uuid.UUID {
	val, ok := ctx.Get(constant.CtxUserID).(string)
	if !ok {
		return uuid.Nil
	}

	return uuid.MustParse(val)
}

// GetEmailFromContext retrieves the user email from the context safely.
func GetEmailFromContext(ctx echo.Context) string {
	val, ok := ctx.Get(constant.CtxEmail).(string)
	if !ok {
		return ""
	}

	return val
}

// GetRolesFromContext retrieves the user roles from the context safely.
func GetRolesFromContext(ctx echo.Context) []string {
	val, ok := ctx.Get(constant.CtxRoles).([]string)
	if !ok {
		return nil
	}

	return val
}

// GetIsActiveFromContext checks if the user is active based on the context safely.
func GetIsActiveFromContext(ctx echo.Context) bool {
	val, ok := ctx.Get(constant.CtxIsActive).(bool)
	if !ok {
		return false
	}

	return val
}
