package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// ClientMetadataMiddleware extracts client IP and User-Agent from incoming requests
// and ensures they're available in request headers for GraphQL processing.
func ClientMetadataMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Extract client IP with fallback logic
			clientIP := extractClientIP(c)
			if clientIP != "" && req.Header.Get(constant.XClientIP) == "" {
				req.Header.Set(constant.XClientIP, clientIP)
			}

			// Extract user agent
			userAgent := req.UserAgent()
			if userAgent != "" && req.Header.Get(constant.XUserAgent) == "" {
				req.Header.Set(constant.XUserAgent, userAgent)
			}

			return next(c)
		}
	}
}

// extractClientIP extracts the real client IP from various headers with fallback logic.
func extractClientIP(c echo.Context) string {
	req := c.Request()

	// Priority 1: X-Client-IP (set by GraphQL gateway or API gateway)
	if ip := req.Header.Get(constant.XClientIP); ip != "" {
		return ip
	}

	// Priority 2: X-Real-IP (common reverse proxy header)
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Priority 3: X-Forwarded-For (take the first IP)
	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain multiple IPs: client, proxy1, proxy2
		// We want the first one (the original client)
		for idx := range len(forwarded) {
			if forwarded[idx] == ',' {
				return forwarded[:idx]
			}
		}

		return forwarded
	}

	// Priority 4: Use Echo's built-in RealIP() which checks multiple sources
	return c.RealIP()
}
