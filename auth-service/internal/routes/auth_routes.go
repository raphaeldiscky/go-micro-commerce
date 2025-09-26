// Package routes provides the HTTP routes for the authentication service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/middleware"
)

// SetupAuthRoutes sets up all authentication routes.
func SetupAuthRoutes(e *echo.Echo, h *handler.AuthHandler) {
	// API versioning
	v1 := e.Group("/v1")

	// Public routes (no authentication required)
	public := v1.Group("")
	public.POST("/register", h.Register)
	public.POST("/login", h.Login)
	public.POST("/refresh-token", h.RefreshToken)
	public.POST("/logout", h.Logout)
	public.POST("/verify", h.VerifyUser)
	public.POST("/resend-verification", h.ResendVerification)

	// User routes (protected)
	protected := v1.Group("/users")
	protected.Use(middleware.AuthMiddleware)
	protected.GET("/whoami", h.GetLoggedInUser)
	protected.PUT("", h.UpdateLoggedInUser)

	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
}
