package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/handler"
)

// SetupAuthRoutes sets up all authentication routes.
func SetupAuthRoutes(e *echo.Echo, authHandler *handler.AuthHandler) {
	// Health and readiness checks
	e.GET("/health", authHandler.Health)
	e.GET("/ready", authHandler.Health)

	// API versioning
	v1 := e.Group("/api/v1")

	// Public routes (no authentication required)
	auth := v1.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.RefreshToken)

	// Session management
	auth.POST("/logout", authHandler.Logout)

	// Add CORS middleware for all routes
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.DELETE,
			echo.PATCH,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))

	// Request logging
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
}
