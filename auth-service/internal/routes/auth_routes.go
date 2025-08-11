// Package routes provides the HTTP routes for the authentication service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/handler"
)

// SetupAuthRoutes sets up all authentication routes.
func SetupAuthRoutes(e *echo.Echo, h *handler.AuthHandler) {
	// API versioning
	v1 := e.Group("/v1")

	// Public routes (no authentication required)
	auth := v1.Group("")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh-token", h.RefreshToken)
	auth.POST("/logout", h.Logout)
	auth.POST("/verify", h.VerifyEmail)
	auth.POST("/resend-verification", h.ResendVerification)

	// User routes (protected)
	users := v1.Group("/users")
	users.GET("/:id", h.GetUser)
	users.PUT("/:id", h.UpdateUser)
}
