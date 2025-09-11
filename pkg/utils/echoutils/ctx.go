// Package echoutils provides utility functions for working with Echo context.
package echoutils

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
)

// GetUserIDFromContext retrieves the user ID (UUID) from the context safely.
func GetUserIDFromContext(ctx echo.Context) uuid.UUID {
	if val, ok := ctx.Get(string(constant.CtxUserID)).(uuid.UUID); ok {
		return val
	}

	return uuid.Nil
}

// GetEmailFromContext retrieves the user email from the context safely.
func GetEmailFromContext(ctx echo.Context) string {
	val, ok := ctx.Get(string(constant.CtxEmail)).(string)
	if !ok {
		return ""
	}

	return val
}

// GetRolesFromContext retrieves the user roles from the context safely.
func GetRolesFromContext(ctx echo.Context) []string {
	val, ok := ctx.Get(string(constant.CtxRoles)).([]string)
	if !ok {
		return nil
	}

	return val
}

// GetIsActiveFromContext checks if the user is active based on the context safely.
func GetIsActiveFromContext(ctx echo.Context) bool {
	val, ok := ctx.Get(string(constant.CtxIsActive)).(bool)
	if !ok {
		return false
	}

	return val
}

// ContextWithUserInfo creates a Go context with user information from Echo context.
// This is useful for passing user info to gRPC clients that need authentication headers.
func ContextWithUserInfo(c echo.Context) context.Context {
	ctx := c.Request().Context()

	// Extract user info from Echo context and add to Go context
	if userID := GetUserIDFromContext(c); userID != uuid.Nil {
		ctx = context.WithValue(ctx, constant.CtxUserID, userID)
	}

	if email := GetEmailFromContext(c); email != "" {
		ctx = context.WithValue(ctx, constant.CtxEmail, email)
	}

	if roles := GetRolesFromContext(c); len(roles) > 0 {
		ctx = context.WithValue(ctx, constant.CtxRoles, roles)
	}

	if isActive := GetIsActiveFromContext(c); isActive {
		ctx = context.WithValue(ctx, constant.CtxIsActive, isActive)
	}

	return ctx
}

// PropagateUserContextToBackground creates a new background context with user information
// from the original context. This is useful for async operations that need to preserve
// authentication context for gRPC calls.
func PropagateUserContextToBackground(ctx context.Context) context.Context {
	bgCtx := context.Background()

	// Copy user authentication information
	if userID := ctx.Value(constant.CtxUserID); userID != nil {
		bgCtx = context.WithValue(bgCtx, constant.CtxUserID, userID)
	}

	if email := ctx.Value(constant.CtxEmail); email != nil {
		bgCtx = context.WithValue(bgCtx, constant.CtxEmail, email)
	}

	if roles := ctx.Value(constant.CtxRoles); roles != nil {
		bgCtx = context.WithValue(bgCtx, constant.CtxRoles, roles)
	}

	if isActive := ctx.Value(constant.CtxIsActive); isActive != nil {
		bgCtx = context.WithValue(bgCtx, constant.CtxIsActive, isActive)
	}

	return bgCtx
}

// GetUserAuthContexts retrieves user information from Go context.
func GetUserAuthContexts(ctx context.Context) (dto.UserAuthInfo, error) {
	var uc dto.UserAuthInfo

	userID, ok := ctx.Value(constant.CtxUserID).(uuid.UUID)
	if !ok {
		return uc, errors.New("failed to get user ID from context")
	}

	uc.UserID = userID

	email, ok := ctx.Value(constant.CtxEmail).(string)
	if !ok {
		return uc, errors.New("failed to get email from context")
	}

	uc.Email = email

	roles, ok := ctx.Value(constant.CtxRoles).([]string)
	if !ok {
		return uc, errors.New("failed to get roles from context")
	}

	uc.Roles = roles

	isActive, ok := ctx.Value(constant.CtxIsActive).(bool)
	if !ok {
		return uc, errors.New("failed to get is active from context")
	}

	uc.IsActive = isActive

	return uc, nil
}

// AddUserAuthToContexts adds user authentication information to the contexts.
func AddUserAuthToContexts(ctx context.Context, uc dto.UserAuthInfo) context.Context {
	ctx = context.WithValue(ctx, constant.CtxUserID, uc.UserID)
	ctx = context.WithValue(ctx, constant.CtxEmail, uc.Email)
	ctx = context.WithValue(ctx, constant.CtxRoles, uc.Roles)
	ctx = context.WithValue(ctx, constant.CtxIsActive, uc.IsActive)

	return ctx
}
