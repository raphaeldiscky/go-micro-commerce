// Package handler provides HTTP handlers for the auth service.
package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/service"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(
	authService service.AuthService,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
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

	// Set refresh token as HTTP-only secure cookie
	h.setRefreshTokenCookie(c, response.RefreshToken)

	// Remove refresh token from response (it's now in cookie)
	responseWithoutRefreshToken := &dto.AuthResponse{
		AccessToken: response.AccessToken,
		TokenType:   response.TokenType,
		ExpiresIn:   response.ExpiresIn,
		User:        response.User,
	}

	return echoutils.ResponseCreated(c, responseWithoutRefreshToken)
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

	// Set refresh token as HTTP-only secure cookie
	h.setRefreshTokenCookie(c, response.RefreshToken)

	// Remove refresh token from response (it's now in cookie)
	responseWithoutRefreshToken := &dto.AuthResponse{
		AccessToken: response.AccessToken,
		TokenType:   response.TokenType,
		ExpiresIn:   response.ExpiresIn,
		User:        response.User,
	}

	return echoutils.ResponseCreated(c, responseWithoutRefreshToken)
}

// RefreshToken handles token refresh.
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	// Read refresh token from HTTP-only cookie
	cookie, err := c.Cookie(constant.RefreshTokenCookieName)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token not found")
	}

	if cookie.Value == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token not found")
	}

	// Create request with refresh token from cookie
	req := &dto.RefreshTokenRequest{
		RefreshToken: cookie.Value,
	}

	response, err := h.authService.RefreshToken(c.Request().Context(), req)
	if err != nil {
		return err
	}

	// Set new refresh token as HTTP-only secure cookie
	h.setRefreshTokenCookie(c, response.RefreshToken)

	// Remove refresh token from response (it's now in cookie)
	responseWithoutRefreshToken := &dto.AuthResponse{
		AccessToken: response.AccessToken,
		TokenType:   response.TokenType,
		ExpiresIn:   response.ExpiresIn,
		User:        response.User,
	}

	return echoutils.ResponseOK(c, responseWithoutRefreshToken)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(c echo.Context) error {
	// Read refresh token from HTTP-only cookie
	cookie, err := c.Cookie(constant.RefreshTokenCookieName)
	if err == nil && cookie.Value != "" {
		// Create request with refresh token from cookie
		req := &dto.LogoutRequest{
			RefreshToken: cookie.Value,
		}

		if err = h.authService.Logout(c.Request().Context(), req); err != nil {
			return err
		}
	}

	// Clear refresh token cookie
	h.clearRefreshTokenCookie(c)

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

// setRefreshTokenCookie sets the refresh token as an HTTP-only secure cookie.
func (h *AuthHandler) setRefreshTokenCookie(c echo.Context, refreshToken string) {
	cookie := &http.Cookie{
		Name:     constant.RefreshTokenCookieName,
		Value:    refreshToken,
		Path:     constant.RefreshTokenCookiePath,
		Domain:   constant.RefreshTokenCookieDomain,
		MaxAge:   constant.RefreshTokenCookieMaxAge,
		Secure:   constant.CookieSecure,
		HttpOnly: constant.CookieHTTPOnly,
		SameSite: http.SameSiteStrictMode,
	}

	c.SetCookie(cookie)
}

// clearRefreshTokenCookie clears the refresh token cookie.
func (h *AuthHandler) clearRefreshTokenCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     constant.RefreshTokenCookieName,
		Value:    "",
		Path:     constant.RefreshTokenCookiePath,
		Domain:   constant.RefreshTokenCookieDomain,
		MaxAge:   -1,                             // Expire immediately
		Expires:  time.Now().Add(-1 * time.Hour), // Set to past time
		Secure:   constant.CookieSecure,
		HttpOnly: constant.CookieHTTPOnly,
		SameSite: http.SameSiteStrictMode,
	}

	c.SetCookie(cookie)
}
