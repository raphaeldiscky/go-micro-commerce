// Package middleware provides authentication and authorization middleware.
package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/jwtutils"
)

// AuthMiddleware handles authentication-related middleware.
type AuthMiddleware struct {
	jwtUtils jwtutils.JWTInterface
}

// NewAuthMiddleware creates a new AuthMiddleware instance.
func NewAuthMiddleware(jwtUtils jwtutils.JWTInterface) *AuthMiddleware {
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
				return err
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
			c.Set(constant.CtxUserID, userID)
			c.Set(constant.CtxEmail, claims.Email)
			c.Set(constant.CtxRoles, claims.Roles)
			c.Set(constant.CtxIsActive, claims.IsActive)

			return next(c)
		}
	}
}

// parseAccessToken extracts the access token from the request context.
func (m *AuthMiddleware) parseAccessToken(c echo.Context) (string, error) {
	accessToken := c.Request().Header.Get("Authorization")
	if accessToken == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	log.Printf("Authorization header: %s", accessToken)

	splitToken := strings.Split(accessToken, " ")
	if len(splitToken) != 2 || splitToken[0] != "Bearer" {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	return splitToken[1], nil
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
