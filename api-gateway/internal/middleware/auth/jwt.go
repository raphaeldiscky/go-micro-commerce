// Package auth provides JWT authentication middleware for Echo framework.
package auth

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	echojwt "github.com/labstack/echo-jwt/v4"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
)

// Claims represents JWT claims.
type Claims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IsActive bool     `json:"is_active"`
	jwt.RegisteredClaims
}

// JWTMiddleware creates JWT authentication middleware.
func JWTMiddleware(cfg *config.JWTConfig) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(cfg.Secret),
		SigningMethod: cfg.SigningMethod,
		TokenLookup:   cfg.TokenLookup,
		ContextKey:    cfg.ContextKey,
		NewClaimsFunc: func(_ echo.Context) jwt.Claims {
			return &Claims{}
		},
		ErrorHandler: func(_ echo.Context, _ error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
		},
		SuccessHandler: func(c echo.Context) {
			token, ok := c.Get(cfg.ContextKey).(*jwt.Token)
			if !ok {
				return
			}

			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
				// Check if user is active
				if !claims.IsActive {
					panic(echo.NewHTTPError(http.StatusForbidden, "user account is inactive"))
				}

				// Add user info to context
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
				c.Set("user_roles", claims.Roles)
			}
		},
	})
}

// CustomJWTMiddleware creates a custom JWT authentication middleware without external dependencies.
func CustomJWTMiddleware(cfg *config.JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from request
			tokenString := ExtractToken(c, cfg)
			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid token")
			}

			// Validate token
			claims, err := ValidateToken(tokenString, cfg)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			// Check if user is active
			if !claims.IsActive {
				return echo.NewHTTPError(http.StatusForbidden, "user account is inactive")
			}

			// Add user info to context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_roles", claims.Roles)
			c.Set(cfg.ContextKey, claims)

			return next(c)
		}
	}
}

// RequireRole creates middleware that requires specific roles.
func RequireRole(requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRoles, ok := c.Get("user_roles").([]string)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "user roles not found")
			}

			hasRole := false

			for _, requiredRole := range requiredRoles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRole = true

						break
					}
				}

				if hasRole {
					break
				}
			}

			if !hasRole {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			return next(c)
		}
	}
}

// ExtractToken extracts token from request.
func ExtractToken(c echo.Context, cfg *config.JWTConfig) string {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], cfg.AuthScheme) {
		return ""
	}

	return parts[1]
}

// ValidateToken validates a JWT token.
func ValidateToken(tokenString string, cfg *config.JWTConfig) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(_ *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenMalformed
}
