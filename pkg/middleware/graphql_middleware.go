package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// GraphQLContextMiddleware extracts HTTP headers and adds them to GraphQL context.
// This middleware should be used with gqlgen's AroundOperations.
// It extracts both client metadata (IP, user agent) and user authentication headers
// (forwarded by Apollo Router or API Gateway from JWT claims).
// Note: This middleware is permissive - it doesn't enforce authentication.
func GraphQLContextMiddleware() graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		// Get the HTTP request from GraphQL context
		requestContext := graphql.GetOperationContext(ctx)

		// If requestContext or headers are nil (e.g., during WebSocket connection_init),
		// continue without setting context values. This allows WebSocket connections to
		// establish first, then auth can be enforced on actual operations.
		if requestContext == nil || requestContext.Headers == nil {
			return next(ctx)
		}
		// Extract client metadata
		if clientIP := requestContext.Headers.Get(constant.XClientIP); clientIP != "" {
			ctx = context.WithValue(ctx, constant.CtxKeyClientIP, clientIP)
		}

		if userAgent := requestContext.Headers.Get(constant.XUserAgent); userAgent != "" {
			ctx = context.WithValue(ctx, constant.CtxKeyUserAgent, userAgent)
		}

		// Extract user authentication headers (forwarded from Apollo Router JWT claims)
		if userIDHeader := requestContext.Headers.Get(constant.XUserID); userIDHeader != "" {
			if userID, err := uuid.Parse(userIDHeader); err == nil {
				ctx = context.WithValue(ctx, constant.CtxKeyUserID, userID)
			}
		}

		if email := requestContext.Headers.Get(constant.XEmail); email != "" {
			ctx = context.WithValue(ctx, constant.CtxKeyEmail, email)
		}

		if rolesHeader := requestContext.Headers.Get(constant.XRoles); rolesHeader != "" {
			// Parse comma-separated roles string into []string
			roles := strings.Split(rolesHeader, ",")
			// Trim whitespace from each role
			for i, role := range roles {
				roles[i] = strings.TrimSpace(role)
			}

			ctx = context.WithValue(ctx, constant.CtxKeyRoles, roles)
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

// GraphQLLoggingMiddleware logs GraphQL operations with operation name, duration, and errors.
// This middleware should be used with gqlgen's AroundOperations.
func GraphQLLoggingMiddleware(appLogger logger.Logger) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		// Get operation context for operation name and other metadata
		opCtx := graphql.GetOperationContext(ctx)
		if opCtx == nil {
			return next(ctx)
		}

		operationName := opCtx.OperationName
		if operationName == "" {
			operationName = "<unnamed>"
		}

		// Execute the operation and capture response
		start := time.Now()
		responseHandler := next(ctx)

		return func(ctx context.Context) *graphql.Response {
			response := responseHandler(ctx)
			duration := time.Since(start)

			// Build log fields
			fields := map[string]any{
				"operation": operationName,
				"duration":  duration.String(),
			}

			// Check for errors in response
			if response != nil && len(response.Errors) > 0 {
				fields["status"] = "error"
				fields["error_count"] = len(response.Errors)
				fields["first_error"] = response.Errors[0].Message

				appLogger.WithFields(fields).Info("GraphQL operation completed with errors")
			} else {
				fields["status"] = "success"
				appLogger.WithFields(fields).Info("GraphQL operation completed")
			}

			return response
		}
	}
}
