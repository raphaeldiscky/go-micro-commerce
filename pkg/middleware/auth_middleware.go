// Package middleware provides authentication and authorization middleware.
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/jwtutils"
)

// AuthMiddleware handles authentication-related middleware.
type AuthMiddleware struct {
	jwtUtils jwtutils.JWT
}

// NewAuthMiddleware creates a new AuthMiddleware instance.
func NewAuthMiddleware(jwtUtils jwtutils.JWT) *AuthMiddleware {
	return &AuthMiddleware{
		jwtUtils: jwtUtils,
	}
}

// Authorization validates the access token and extracts user information.
func (m *AuthMiddleware) Authorization() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken, err := m.parseAccessToken(c)
			if err != nil {
				return echo.NewHTTPError(
					http.StatusUnauthorized,
					"invalid or missing authorization header",
				)
			}

			claims, err := m.jwtUtils.ValidateAccessToken(accessToken)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			// Additional validation
			if claims.UserID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid userID in token")
			}

			if claims.Email == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid email in token")
			}

			// Parse user ID to UUID
			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid userID format in token")
			}

			// Additional validation for active users only
			if !claims.IsActive {
				return echo.NewHTTPError(http.StatusUnauthorized, "user account is inactive")
			}

			// Set user information in context
			c.Set(string(constant.CtxKeyUserID), userID)
			c.Set(string(constant.CtxKeyEmail), claims.Email)
			c.Set(string(constant.CtxKeyRoles), claims.Roles)
			c.Set(string(constant.CtxKeyIsActive), claims.IsActive)

			return next(c)
		}
	}
}

// parseAccessToken extracts the access token from the request context.
func (m *AuthMiddleware) parseAccessToken(c echo.Context) (string, error) {
	accessToken := c.Request().Header.Get("Authorization")
	if accessToken == "" {
		return "", errors.New("missing Authorization header")
	}

	splitToken := strings.Split(accessToken, " ")
	if len(splitToken) != 2 || splitToken[0] != constant.BearerPrefix {
		return "", errors.New("invalid Authorization header format")
	}

	return splitToken[1], nil
}

// OptionalAuthorization validates the access token if present but doesn't return error if missing.
// This is useful for GraphQL endpoints where some queries require auth and others don't.
func (m *AuthMiddleware) OptionalAuthorization() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken, err := m.parseAccessToken(c)
			if err != nil {
				// Token missing or invalid format - continue without setting user context
				return next(c)
			}

			claims, err := m.jwtUtils.ValidateAccessToken(accessToken)
			if err != nil {
				// Token validation failed - continue without setting user context
				return next(c)
			}

			// Validate user ID
			if claims.UserID == "" {
				return next(c)
			}

			// Parse user ID to UUID
			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				return next(c)
			}

			// Set user information in context if validation succeeds
			c.Set(string(constant.CtxKeyUserID), userID)

			if claims.Email != "" {
				c.Set(string(constant.CtxKeyEmail), claims.Email)
			}

			c.Set(string(constant.CtxKeyRoles), claims.Roles)
			c.Set(string(constant.CtxKeyIsActive), claims.IsActive)

			return next(c)
		}
	}
}

// RequireRole is a middleware that checks if the user has a specific role.
func (m *AuthMiddleware) RequireRole(requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles := echoutils.GetRolesFromContext(c)

			// Check if user has the required role
			for _, role := range roles {
				if role == requiredRole {
					return next(c)
				}
			}

			return echo.NewHTTPError(
				http.StatusForbidden,
				"access denied: insufficient permissions",
			)
		}
	}
}

// RequireAnyRole is a middleware that checks if the user has any of the specified roles.
func (m *AuthMiddleware) RequireAnyRole(requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles := echoutils.GetRolesFromContext(c)

			// Check if user has any of the required roles
			for _, userRole := range roles {
				for _, requiredRole := range requiredRoles {
					if userRole == requiredRole {
						return next(c)
					}
				}
			}

			return echo.NewHTTPError(
				http.StatusForbidden,
				"access denied: insufficient permissions",
			)
		}
	}
}

// RequireActiveUser is a middleware that ensures the user account is active.
func (m *AuthMiddleware) RequireActiveUser() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			isActive := echoutils.GetIsActiveFromContext(c)
			if !isActive {
				return echo.NewHTTPError(
					http.StatusForbidden,
					"access denied: user account is inactive",
				)
			}

			return next(c)
		}
	}
}
