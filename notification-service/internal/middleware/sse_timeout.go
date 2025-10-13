// Package middleware provides custom HTTP middleware for the notification service.
package middleware

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
)

// SSETimeout creates a middleware that extends the request context timeout
// specifically for Server-Sent Events (SSE) connections.
func SSETimeout(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create a new context with extended timeout for SSE connections
			ctx, cancel := context.WithTimeout(c.Request().Context(), cfg.SSEServer.Timeout)

			// Replace the request context with the extended timeout context
			c.SetRequest(c.Request().WithContext(ctx))

			// Ensure cancellation function is called when request completes
			defer cancel()

			return next(c)
		}
	}
}
