// Package routes provides the HTTP routes for the authentication service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/handler"
)

// SetupAuthRoutes sets up all authentication routes.
func SetupAuthRoutes(e *echo.Echo, h *handler.AuthHandler) {
	// Health and readiness checks
	e.GET("/health", h.Health)

	// API versioning
	v1 := e.Group("/api/v1")

	// Public routes (no authentication required)
	auth := v1.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.RefreshToken)

	// Session management
	auth.POST("/logout", h.Logout)

	// User routes (protected)
	users := v1.Group("/users")
	users.GET("/:id", h.GetProfile)
	users.PUT("/:id", h.UpdateProfile)
}
