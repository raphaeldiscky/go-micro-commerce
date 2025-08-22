package echoutils

import (
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
)

// GetXUserID retrieves the XUserID header from the context.
func GetXUserID(ctx echo.Context) (int64, bool) {
	xUserID := ctx.Request().Header.Get(constant.XUserID)

	userID, err := strconv.ParseInt(xUserID, 10, 64)
	if err != nil {
		return 0, false
	}

	return userID, true
}

// GetXEmail retrieves the XEmail header from the context.
func GetXEmail(ctx echo.Context) (string, bool) {
	xEmail := ctx.Request().Header.Get(constant.XEmail)
	if xEmail == "" {
		return "", false
	}

	return xEmail, true
}

// GetXRoles retrieves the XRoles header from the context.
func GetXRoles(ctx echo.Context) ([]string, bool) {
	xRoles := ctx.Request().Header.Get(constant.XRoles)
	if xRoles == "" {
		return nil, false
	}

	return strings.Split(xRoles, ","), true
}
