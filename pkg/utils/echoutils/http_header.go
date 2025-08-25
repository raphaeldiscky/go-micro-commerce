package echoutils

import (
	"log"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
)

// GetXUserID retrieves the X-UserID header from the context as UUID.
func GetXUserID(ctx echo.Context) (uuid.UUID, bool) {
	xUserID := ctx.Request().Header.Get(constant.XUserID)
	if xUserID == "" {
		log.Printf("failed to get X-User-ID from headers")

		return uuid.Nil, false
	}

	userID, err := uuid.Parse(xUserID)
	if err != nil {
		log.Printf("failed to parse X-User-ID as UUID: %v", err)

		return uuid.Nil, false
	}

	return userID, true
}

// GetXEmail retrieves the X-Email header from the context.
func GetXEmail(ctx echo.Context) (string, bool) {
	xEmail := ctx.Request().Header.Get(constant.XEmail)
	if xEmail == "" {
		log.Printf("failed to get X-Email from headers")

		return "", false
	}

	return xEmail, true
}

// GetXRoles retrieves the X-Roles header from the context.
func GetXRoles(ctx echo.Context) ([]string, bool) {
	xRoles := ctx.Request().Header.Get(constant.XRoles)
	if xRoles == "" {
		log.Printf("failed to get X-Roles from headers")

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

// GetXIsActive retrieves the X-IsActive header from the context.
func GetXIsActive(ctx echo.Context) (isActive, ok bool) {
	xIsActive := ctx.Request().Header.Get(constant.XIsActive)
	if xIsActive == "" {
		log.Printf("failed to get X-IsActive from headers")

		return false, false
	}

	isActive, err := strconv.ParseBool(xIsActive)
	if err != nil {
		log.Printf("failed to parse X-IsActive: %v", err)

		return false, false
	}

	return isActive, true
}

// GetXRequestID retrieves the X-RequestID header from the context.
func GetXRequestID(ctx echo.Context) (string, bool) {
	xRequestID := ctx.Request().Header.Get(constant.XRequestID)
	if xRequestID == "" {
		log.Printf("failed to get X-Request-ID from headers")

		return "", false
	}

	return xRequestID, true
}
