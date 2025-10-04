// Package cookieutils provides utilities for auth-service cookies
package cookieutils

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
)

// SetRefreshTokenCookie sets the refresh token as an HTTP-only secure cookie in GraphQL context.
func SetRefreshTokenCookie(ctx context.Context, refreshToken string) {
	w, ok := ctx.Value(pkgconstant.CtxKeyResponseWriter).(http.ResponseWriter)
	if !ok {
		// Fallback: response writer not in context
		return
	}

	sameSite := getSameSiteMode(constant.CookieSameSite)

	cookie := &http.Cookie{
		Name:     constant.RefreshTokenCookieName,
		Value:    refreshToken,
		Path:     constant.RefreshTokenCookiePath,
		Domain:   constant.RefreshTokenCookieDomain,
		MaxAge:   constant.RefreshTokenCookieMaxAge,
		Secure:   constant.CookieSecure,
		HttpOnly: constant.CookieHTTPOnly,
		SameSite: sameSite,
	}

	http.SetCookie(w, cookie)
}

// ClearRefreshTokenCookie clears the refresh token cookie in GraphQL context.
func ClearRefreshTokenCookie(ctx context.Context) {
	w, ok := ctx.Value(pkgconstant.CtxKeyResponseWriter).(http.ResponseWriter)
	if !ok {
		// Fallback: response writer not in context
		return
	}

	sameSite := getSameSiteMode(constant.CookieSameSite)

	cookie := &http.Cookie{
		Name:     constant.RefreshTokenCookieName,
		Value:    "",
		Path:     constant.RefreshTokenCookiePath,
		Domain:   constant.RefreshTokenCookieDomain,
		MaxAge:   -1, // Delete cookie
		Secure:   constant.CookieSecure,
		HttpOnly: constant.CookieHTTPOnly,
		SameSite: sameSite,
	}

	http.SetCookie(w, cookie)
}

// getSameSiteMode converts a string SameSite value to http.SameSite constant.
func getSameSiteMode(sameSite string) http.SameSite {
	switch sameSite {
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// GetRefreshTokenFromCookie gets the refresh token from HTTP-only cookie in GraphQL context.
func GetRefreshTokenFromCookie(ctx context.Context) (string, error) {
	// Get the HTTP request from GraphQL context
	requestContext := graphql.GetOperationContext(ctx)
	if requestContext == nil {
		return "", http.ErrNoCookie
	}

	// Extract cookie from request headers
	cookies := requestContext.Headers.Values("Cookie")
	if len(cookies) == 0 {
		return "", http.ErrNoCookie
	}

	// Parse cookies
	header := http.Header{}
	for _, cookie := range cookies {
		header.Add("Cookie", cookie)
	}

	req := &http.Request{Header: header}

	cookie, err := req.Cookie(constant.RefreshTokenCookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}
