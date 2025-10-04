package middleware

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/jwtutils"
)

// GraphQLContextMiddleware extracts HTTP headers and adds them to GraphQL context.
// This middleware should be used with gqlgen's AroundOperations.
func GraphQLContextMiddleware() graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		// Get the HTTP request from GraphQL context
		requestContext := graphql.GetOperationContext(ctx)
		if requestContext != nil && requestContext.Headers != nil {
			if clientIP := requestContext.Headers.Get(constant.XClientIP); clientIP != "" {
				ctx = context.WithValue(ctx, constant.CtxKeyClientIP, clientIP)
			}

			if userAgent := requestContext.Headers.Get(constant.XUserAgent); userAgent != "" {
				ctx = context.WithValue(ctx, constant.CtxKeyUserAgent, userAgent)
			}
		}

		return next(ctx)
	}
}

// ExtractClientIP extracts client IP from GraphQL context with fallback.
func ExtractClientIP(ctx context.Context) string {
	if clientIP, ok := ctx.Value(constant.CtxKeyClientIP).(string); ok && clientIP != "" {
		return clientIP
	}

	return ""
}

// ExtractUserAgent extracts user agent from GraphQL context with fallback.
func ExtractUserAgent(ctx context.Context) string {
	if userAgent, ok := ctx.Value(constant.CtxKeyUserAgent).(string); ok && userAgent != "" {
		return userAgent
	}

	return ""
}

// GraphQLAuthMiddleware extracts client metadata and validates JWT token from Authorization header.
// This middleware should be used with gqlgen's AroundOperations.
func GraphQLAuthMiddleware(jwtUtils jwtutils.JWT) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		ctx = extractClientMetadata(ctx)
		ctx = extractAndValidateJWT(ctx, jwtUtils)

		return next(ctx)
	}
}

// extractClientMetadata extracts client IP and user agent from GraphQL context.
func extractClientMetadata(ctx context.Context) context.Context {
	requestContext := graphql.GetOperationContext(ctx)
	if requestContext == nil || requestContext.Headers == nil {
		return ctx
	}

	if clientIP := requestContext.Headers.Get(constant.XClientIP); clientIP != "" {
		ctx = context.WithValue(ctx, constant.CtxKeyClientIP, clientIP)
	}

	if userAgent := requestContext.Headers.Get(constant.XUserAgent); userAgent != "" {
		ctx = context.WithValue(ctx, constant.CtxKeyUserAgent, userAgent)
	}

	return ctx
}

// extractAndValidateJWT extracts and validates JWT token from Authorization header.
func extractAndValidateJWT(ctx context.Context, jwtUtils jwtutils.JWT) context.Context {
	requestContext := graphql.GetOperationContext(ctx)
	if requestContext == nil || requestContext.Headers == nil {
		return ctx
	}

	authHeader := requestContext.Headers.Get("Authorization")
	if authHeader == "" {
		return ctx
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != constant.BearerPrefix {
		return ctx
	}

	token := parts[1]

	// Validate token
	claims, err := jwtUtils.ValidateAccessToken(token)
	if err != nil || claims.UserID == "" {
		return ctx
	}

	// Parse user ID to UUID
	userID, parseErr := uuid.Parse(claims.UserID)
	if parseErr != nil {
		return ctx
	}

	// Set user information in context
	ctx = context.WithValue(ctx, constant.CtxKeyUserID, userID)
	ctx = context.WithValue(ctx, constant.CtxKeyEmail, claims.Email)
	ctx = context.WithValue(ctx, constant.CtxKeyRoles, claims.Roles)
	ctx = context.WithValue(ctx, constant.CtxKeyIsActive, claims.IsActive)

	return ctx
}

// GraphQLRequireAuth ensures a user is authenticated for GraphQL operations.
// Returns a GraphQL error if user is not authenticated.
func GraphQLRequireAuth() graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		// Check if user ID exists in context
		if _, ok := ctx.Value(constant.CtxKeyUserID).(uuid.UUID); !ok {
			return func(_ context.Context) *graphql.Response {
				return &graphql.Response{
					Errors: gqlerror.List{
						&gqlerror.Error{
							Message: "unauthorized: missing or invalid authentication token",
							Extensions: map[string]any{
								"code": "UNAUTHENTICATED",
							},
						},
					},
				}
			}
		}

		return next(ctx)
	}
}
