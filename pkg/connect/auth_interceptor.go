// Package connect provides Connect-RPC interceptors for authentication and authorization.
package connect

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
)

// AuthInterceptor provides authentication interceptor for Connect-RPC services.
type AuthInterceptor struct{}

// NewAuthInterceptor creates a new authentication interceptor.
func NewAuthInterceptor() *AuthInterceptor {
	return &AuthInterceptor{}
}

// ServiceToServiceAuth creates a Connect interceptor that validates user headers from API Gateway.
func (a *AuthInterceptor) ServiceToServiceAuth() connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(
			func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				// Skip auth for health checks
				if strings.HasSuffix(req.Spec().Procedure, "Health") {
					return next(ctx, req)
				}

				// Extract user information from HTTP headers (forwarded from API Gateway)
				userInfo, err := a.extractUserInfoFromHeaders(req.Header())
				if err != nil {
					return nil, connect.NewError(
						connect.CodeUnauthenticated,
						err,
					)
				}

				// Add user information to context for downstream use
				newCtx := context.WithValue(ctx, constant.CtxUserID, userInfo.UserID)
				newCtx = context.WithValue(newCtx, constant.CtxEmail, userInfo.Email)
				newCtx = context.WithValue(newCtx, constant.CtxRoles, userInfo.Roles)
				newCtx = context.WithValue(newCtx, constant.CtxIsActive, userInfo.IsActive)

				return next(newCtx, req)
			},
		)
	})
}

// extractUserInfoFromHeaders extracts user information from HTTP headers.
func (a *AuthInterceptor) extractUserInfoFromHeaders(
	headers http.Header,
) (*dto.UserAuthInfo, error) {
	// Extract X-User-ID
	userIDValue := headers.Get(constant.XUserID)
	if userIDValue == "" {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New(constant.MissingXUserIDErrorMessage),
		)
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New(constant.InvalidXuserIDFormateErrorMessage),
		)
	}

	// Extract X-Email
	email := headers.Get(constant.XEmail)
	if email == "" {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New(constant.MissingXEmailErrorMessage),
		)
	}

	// Extract X-Roles
	rolesValue := headers.Get(constant.XRoles)
	if rolesValue == "" {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New(constant.MissingXRolesErrorMessage),
		)
	}

	roles := strings.Split(rolesValue, ",")

	// Extract X-Is-Active
	isActiveValue := headers.Get(constant.XIsActive)
	if isActiveValue == "" {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New(constant.MissingXIsActiveErrorMessage),
		)
	}

	isActive, err := strconv.ParseBool(isActiveValue)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New(constant.InvalidXIsActiveFormatErrorMessage),
		)
	}

	return &dto.UserAuthInfo{
		UserID:   userID,
		Email:    email,
		Roles:    roles,
		IsActive: isActive,
	}, nil
}
