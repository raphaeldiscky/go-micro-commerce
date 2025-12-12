// Package routes provides the HTTP routes for the authentication service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/middleware"
)

// SetupAuthRoutes sets up all authentication routes.
func SetupAuthRoutes(e *echo.Echo, h *handler.AuthHandler, jwksHandler *handler.JWKSHandler) {
	// JWKS endpoint (must be at root path for standard compliance)
	e.GET("/.well-known/jwks.json", jwksHandler.GetJWKS)

	// Public routes (no authentication required)
	public := e.Group("")
	public.POST("/register", h.Register)
	public.POST("/login", h.Login)
	public.POST("/refresh-token", h.RefreshToken)
	public.POST("/logout", h.Logout)
	public.POST("/verify", h.VerifyUser)
	public.POST("/resend-verification", h.ResendVerification)

	// User routes (protected)
	protected := e.Group("/users")
	protected.Use(middleware.AuthMiddleware)
	protected.GET("/whoami", h.GetLoggedInUser)
	protected.PUT("", h.UpdateLoggedInUser)

	admin := protected.Group("")
	admin.Use(middleware.RequireAdminRole)
}
