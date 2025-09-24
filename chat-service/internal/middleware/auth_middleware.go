// Package middleware provides authentication middleware for the application.
package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
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
