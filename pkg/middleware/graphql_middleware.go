package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// GraphQLContextMiddleware extracts HTTP headers and adds them to GraphQL context.
// This middleware should be used with gqlgen's AroundOperations.
func GraphQLContextMiddleware() graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		// Get the HTTP request from GraphQL context
		requestContext := graphql.GetOperationContext(ctx)
		if requestContext != nil && requestContext.Headers != nil {
			// Extract client IP from header (set by API Gateway)
			if clientIP := requestContext.Headers.Get(constant.XClientIP); clientIP != "" {
				ctx = context.WithValue(ctx, constant.CtxKeyClientIP, clientIP)
			}

			// Extract user agent from header (set by API Gateway)
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
