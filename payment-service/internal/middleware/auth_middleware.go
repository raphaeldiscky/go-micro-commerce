// Package middleware provides authentication middleware for the application.
package middleware

import (
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
	"github.com/raphaeldiscky/go-micro-template/pkg/httperror"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/echoutils"
)

// AuthMiddleware is a middleware function that checks for the presence of user information in the context.
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		userID, ok := echoutils.GetXUserID(ctx)
		if !ok {
			return httperror.NewMissingXUserIDError()
		}

		email, ok := echoutils.GetXEmail(ctx)
		if !ok {
			return httperror.NewMissingXEmailError()
		}

		roles, ok := echoutils.GetXRoles(ctx)
		if !ok {
			return httperror.NewMissingXRolesError()
		}

		isActive, ok := echoutils.GetXIsActive(ctx)
		if !ok {
			return httperror.NewMissingXIsActiveError()
		}

		ctx.Set(string(constant.CtxUserID), userID)
		ctx.Set(string(constant.CtxEmail), email)
		ctx.Set(string(constant.CtxRoles), roles)
		ctx.Set(string(constant.CtxIsActive), isActive)

		return next(ctx)
	}
}

// RequireAdminRole is a middleware that checks if the user has a specific role.
func RequireAdminRole(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roles := echoutils.GetRolesFromContext(c)

		log.Infof("user roles: %v", roles)
		// Check if user has the admin role
		if slices.Contains(roles, constant.RoleAdmin) {
			return next(c)
		}

		return echo.NewHTTPError(
			http.StatusForbidden,
			"access denied: insufficient permissions",
		)
	}
}
