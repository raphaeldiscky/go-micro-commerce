// Package middleware provides authentication middleware for the application.
package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	chatConstant "github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
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

// Auth represents authentication context information.
type Auth struct {
	UserID   uuid.UUID
	Email    string
	UserType chatConstant.UserType
	IsActive bool
}
