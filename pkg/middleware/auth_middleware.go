// Package middleware provides authentication and authorization middleware.
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
	"github.com/raphaeldiscky/go-micro-template/pkg/httperror"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/jwtutils"
)

// AuthMiddleware handles authentication-related middleware.
type AuthMiddleware struct {
	jwtUtil jwtutils.JWTUtil
}

// NewAuthMiddleware creates a new AuthMiddleware instance.
func NewAuthMiddleware(jwtUtil jwtutils.JWTUtil) *AuthMiddleware {
	return &AuthMiddleware{
		jwtUtil: jwtUtil,
	}
}

// Authorization middleware.
func (m *AuthMiddleware) Authorization() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken, err := m.parseAccessToken(c)
			if err != nil {
				return err
			}

			claims, err := m.jwtUtil.Parse(accessToken)
			if err != nil {
				return httperror.NewUnauthorizedError()
			}

			// Set user information in context
			c.Set(constant.CtxUserID, claims.UserID)
			c.Set(constant.CtxEmail, claims.Email)

			return next(c)
		}
	}
}

// parseAccessToken extracts the access token from the request context.
func (m *AuthMiddleware) parseAccessToken(c echo.Context) (string, error) {
	accessToken := c.Request().Header.Get("Authorization")
	if accessToken == "" {
		return "", httperror.NewUnauthorizedError()
	}

	splitToken := strings.Split(accessToken, " ")
	if len(splitToken) != 2 || splitToken[0] != "Bearer" {
		return "", httperror.NewUnauthorizedError()
	}

	return splitToken[1], nil
}

// GetUserIDFromContext extracts the user ID from the request context.
func GetUserIDFromContext(c echo.Context) (int64, bool) {
	if userID := c.Get(constant.CtxUserID); userID != nil {
		if id, ok := userID.(int64); ok {
			return id, true
		}
	}

	return 0, false
}

// GetEmailFromContext extracts the email from the request context.
func GetEmailFromContext(c echo.Context) (string, bool) {
	if email := c.Get(constant.CtxEmail); email != nil {
		if emailStr, ok := email.(string); ok {
			return emailStr, true
		}
	}

	return "", false
}

// AuthorizationWithValidation validates the access token and extracts user information.
func (m *AuthMiddleware) AuthorizationWithValidation() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken, err := m.parseAccessToken(c)
			if err != nil {
				return echo.NewHTTPError(
					http.StatusUnauthorized,
					"Invalid or missing authorization header",
				)
			}

			claims, err := m.jwtUtil.Parse(accessToken)
			if err != nil {
				// Handle specific JWT errors
				switch {
				case errors.Is(err, jwtutils.ErrTokenExpired):
					return echo.NewHTTPError(http.StatusUnauthorized, "Token has expired")
				case errors.Is(err, jwtutils.ErrInvalidSignature):
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token signature")
				case errors.Is(err, jwtutils.ErrMissingClaims):
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
				default:
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
				}
			}

			// Additional validation
			if claims.UserID <= 0 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID in token")
			}

			if claims.Email == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email in token")
			}

			// Set user information in context
			c.Set(constant.CtxUserID, claims.UserID)
			c.Set(constant.CtxEmail, claims.Email)

			return next(c)
		}
	}
}

// OptionalAuthorization is a middleware for optional authentication (doesn't fail if no token).
func (m *AuthMiddleware) OptionalAuthorization() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken, err := m.parseAccessToken(c)
			if err != nil {
				// Continue without authentication
				return next(c)
			}

			claims, err := m.jwtUtil.Parse(accessToken)
			if err != nil {
				// Continue without authentication
				return next(c)
			}

			// Set user information in context if token is valid
			c.Set(constant.CtxUserID, claims.UserID)
			c.Set(constant.CtxEmail, claims.Email)

			return next(c)
		}
	}
}
