// Package grpc provides gRPC interceptors for authentication and authorization.
package grpc

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
)

// AuthInterceptor provides authentication interceptor for gRPC services.
type AuthInterceptor struct{}

// NewAuthInterceptor creates a new authentication interceptor.
func NewAuthInterceptor() *AuthInterceptor {
	return &AuthInterceptor{}
}

// ServiceToServiceAuth creates a unary interceptor that validates user headers from API Gateway.
func (a *AuthInterceptor) ServiceToServiceAuth() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Skip auth for health checks
		if strings.HasSuffix(info.FullMethod, "Health") {
			return handler(ctx, req)
		}

		// Extract user information from metadata (forwarded from API Gateway)
		userInfo, err := a.extractUserInfoFromMetadata(ctx)
		if err != nil {
			return nil, status.Errorf(
				codes.Unauthenticated,
				"missing or invalid user information: %v",
				err,
			)
		}

		// Add user information to context for downstream use
		newCtx := context.WithValue(ctx, constant.CtxKeyUserID, userInfo.UserID)
		newCtx = context.WithValue(newCtx, constant.CtxKeyEmail, userInfo.Email)
		newCtx = context.WithValue(newCtx, constant.CtxKeyRoles, userInfo.Roles)
		newCtx = context.WithValue(newCtx, constant.CtxKeyIsActive, userInfo.IsActive)

		return handler(newCtx, req)
	}
}

// extractUserInfoFromMetadata extracts user information from gRPC metadata.
func (a *AuthInterceptor) extractUserInfoFromMetadata(
	ctx context.Context,
) (*dto.UserAuthInfo, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, constant.MissingMetadataErrorMessage)
	}

	// Extract X-User-ID (gRPC metadata keys are lowercased)
	userIDValues := md[strings.ToLower(constant.XUserID)]
	if len(userIDValues) == 0 {
		return nil, status.Error(codes.Unauthenticated, constant.MissingXUserIDErrorMessage)
	}

	userID, err := uuid.Parse(userIDValues[0])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, constant.InvalidXuserIDFormateErrorMessage)
	}

	// Extract X-Email
	emailValues := md[strings.ToLower(constant.XEmail)]
	if len(emailValues) == 0 {
		return nil, status.Error(codes.Unauthenticated, constant.MissingXEmailErrorMessage)
	}

	email := emailValues[0]

	// Extract X-Roles
	rolesValues := md[strings.ToLower(constant.XRoles)]
	if len(rolesValues) == 0 {
		return nil, status.Error(codes.Unauthenticated, constant.MissingXRolesErrorMessage)
	}

	roles := strings.Split(rolesValues[0], ",")

	// Extract X-Is-Active
	isActiveValues := md[strings.ToLower(constant.XIsActive)]
	if len(isActiveValues) == 0 {
		return nil, status.Error(codes.Unauthenticated, constant.MissingXIsActiveErrorMessage)
	}

	isActive, err := strconv.ParseBool(isActiveValues[0])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, constant.InvalidXIsActiveFormatErrorMessage)
	}

	return &dto.UserAuthInfo{
		UserID:   userID,
		Email:    email,
		Roles:    roles,
		IsActive: isActive,
	}, nil
}
