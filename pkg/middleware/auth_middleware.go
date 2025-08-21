// Package middleware provides authentication and authorization middleware.
package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
	"github.com/raphaeldiscky/go-micro-template/pkg/httperror"
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

// Authorization middleware.
func (m *AuthMiddleware) Authorization() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken, err := m.parseAccessToken(c)
			if err != nil {
				return err
			}

			claims, err := m.jwtUtils.ValidateAccessToken(accessToken)
			if err != nil {
				return httperror.NewUnauthorizedError()
			}

			// Parse user ID to UUID
			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				return httperror.NewUnauthorizedError()
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
		return "", httperror.NewUnauthorizedError()
	}

	splitToken := strings.Split(accessToken, " ")
	if len(splitToken) != 2 || splitToken[0] != "Bearer" {
		return "", httperror.NewUnauthorizedError()
	}

	return splitToken[1], nil
}

// GetUserIDFromContext extracts the user ID from the request context.
func GetUserIDFromContext(c echo.Context) (uuid.UUID, bool) {
	if userID := c.Get(constant.CtxUserID); userID != nil {
		if id, ok := userID.(uuid.UUID); ok {
			return id, true
		}
	}

	return uuid.Nil, false
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

// GetRolesFromContext extracts the roles from the request context.
func GetRolesFromContext(c echo.Context) ([]string, bool) {
	if roles := c.Get(constant.CtxRoles); roles != nil {
		if rolesSlice, ok := roles.([]string); ok {
			return rolesSlice, true
		}
	}

	return nil, false
}

// GetIsActiveFromContext retrieves the "is active" status from the Echo context.
func GetIsActiveFromContext(c echo.Context) (isActive, ok bool) {
	if isActive := c.Get(constant.CtxIsActive); isActive != nil {
		if active, ok := isActive.(bool); ok {
			return active, true
		}
	}

	return false, false
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

			claims, err := m.jwtUtils.ValidateAccessToken(accessToken)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			// Additional validation
			if claims.UserID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID in token")
			}

			if claims.Email == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email in token")
			}

			// Parse user ID to UUID
			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID format in token")
			}

			// Additional validation for active users only
			if !claims.IsActive {
				return echo.NewHTTPError(http.StatusUnauthorized, "User account is inactive")
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

// OptionalAuthorization is a middleware for optional authentication (doesn't fail if no token).
func (m *AuthMiddleware) OptionalAuthorization() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken, err := m.parseAccessToken(c)
			if err != nil {
				// Continue without authentication
				return next(c)
			}

			claims, err := m.jwtUtils.ValidateAccessToken(accessToken)
			if err != nil {
				// Continue without authentication
				return next(c)
			}

			// Parse user ID to UUID
			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				// Continue without authentication
				return next(c)
			}

			// Set user information in context if token is valid
			c.Set(constant.CtxUserID, userID)
			c.Set(constant.CtxEmail, claims.Email)
			c.Set(constant.CtxRoles, claims.Roles)
			c.Set(constant.CtxIsActive, claims.IsActive)

			return next(c)
		}
	}
}

// RequireRole is a middleware that checks if the user has a specific role.
func (m *AuthMiddleware) RequireRole(requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles, ok := GetRolesFromContext(c)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "Access denied: no roles found")
			}

			// Check if user has the required role
			for _, role := range roles {
				if role == requiredRole {
					return next(c)
				}
			}

			return echo.NewHTTPError(
				http.StatusForbidden,
				"Access denied: insufficient permissions",
			)
		}
	}
}

// RequireAnyRole is a middleware that checks if the user has any of the specified roles.
func (m *AuthMiddleware) RequireAnyRole(requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles, ok := GetRolesFromContext(c)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "Access denied: no roles found")
			}

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
				"Access denied: insufficient permissions",
			)
		}
	}
}

// RequireActiveUser is a middleware that ensures the user account is active.
func (m *AuthMiddleware) RequireActiveUser() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			isActive, ok := GetIsActiveFromContext(c)
			if !ok || !isActive {
				return echo.NewHTTPError(
					http.StatusForbidden,
					"Access denied: user account is inactive",
				)
			}

			return next(c)
		}
	}
}
