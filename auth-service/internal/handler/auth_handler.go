// Package handler provides HTTP handlers for the auth service.
package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/service"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	authService service.AuthService
	logger      logger.Logger
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(
	authService service.AuthService,
	appLogger logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      appLogger,
	}
}

// Register handles user registration.
func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	clientIP := c.RealIP()
	userAgent := c.Request().UserAgent()

	response, err := h.authService.Register(c.Request().Context(), &req, clientIP, userAgent)
	if err != nil {
		return err
	}

	return echoutils.ResponseCreated(c, response)
}

// Login handles user login.
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	clientIP := c.RealIP()
	userAgent := c.Request().UserAgent()

	response, err := h.authService.Login(c.Request().Context(), &req, clientIP, userAgent)
	if err != nil {
		return err
	}

	return echoutils.ResponseCreated(c, response)
}

// RefreshToken handles token refresh.
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req dto.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	response, err := h.authService.RefreshToken(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, response)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(c echo.Context) error {
	var req dto.LogoutRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.authService.Logout(c.Request().Context(), &req); err != nil {
		return err
	}

	return echoutils.ResponseOKPlain(c)
}

// GetLoggedInUser retrieves the user's user.
func (h *AuthHandler) GetLoggedInUser(c echo.Context) error {
	userID := echoutils.GetUserIDFromContext(c)

	user, err := h.authService.GetUser(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, user)
}

// UpdateLoggedInUser updates the user's user.
func (h *AuthHandler) UpdateLoggedInUser(c echo.Context) error {
	userID := echoutils.GetUserIDFromContext(c)

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.authService.UpdateUser(c.Request().Context(), userID, &req)
	if err != nil {
		return err
	}

	return echoutils.ResponseCreated(c, user)
}

// VerifyUser verify user.
func (h *AuthHandler) VerifyUser(c echo.Context) error {
	req := dto.VerifyEmailRequest{
		Token: c.QueryParam("token"),
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.authService.VerifyUser(c.Request().Context(), &req); err != nil {
		return err
	}

	return echoutils.ResponseOKPlain(c)
}

// ResendVerification handles resend email verification.
func (h *AuthHandler) ResendVerification(c echo.Context) error {
	var req dto.ResendVerificationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.authService.ResendVerification(c.Request().Context(), &req); err != nil {
		return err
	}

	return echoutils.ResponseOKPlain(c)
}
