package constant

const (
	// RefreshTokenCookieName is the name of the HTTP-only cookie that stores the refresh token.
	RefreshTokenCookieName = "refresh_token"

	// RefreshTokenCookieMaxAge is the max age for the refresh token cookie in seconds (7 days).
	RefreshTokenCookieMaxAge = 7 * 24 * 60 * 60 // 7 days

	// RefreshTokenCookiePath is the path for the refresh token cookie.
	RefreshTokenCookiePath = "/"

	// RefreshTokenCookieDomain is the domain for the refresh token cookie.
	// Leave empty to use the current domain.
	RefreshTokenCookieDomain = ""
)

const (
	// CookieSecure indicates whether the cookie should only be sent over HTTPS.
	// Should be true in production, false in development.
	CookieSecure = false

	// CookieHTTPOnly indicates whether the cookie should be HTTP-only.
	CookieHTTPOnly = true

	// CookieSameSite defines the SameSite attribute for cookies.
	// Use "Lax" for development to allow cross-origin cookies, "Strict" for production.
	CookieSameSite = "Lax"
)
