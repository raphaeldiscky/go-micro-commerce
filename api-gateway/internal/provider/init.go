package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/middleware"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/jwtutils"

	pkgConfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
)

// Providers holds all initialized providers.
type Providers struct {
	authMiddleware *middleware.AuthMiddleware
}

// SetupGlobal initializes and returns the providers.
func SetupGlobal(cfg *config.Config, appLogger logger.Logger) (*Providers, error) {
	jwtUtil := jwtutils.NewJWTUtils(&pkgConfig.JWTConfig{
		AllowedAlgs:         cfg.JWT.AllowedAlgs,
		Issuer:              cfg.JWT.Issuer,
		PublicKeyPath:       cfg.JWT.PublicKeyPath,
		JWKSUrl:             cfg.JWT.JWKSUrl,
		JWKSCacheTTL:        cfg.JWT.JWKSCacheTTL,
		JWKSRefreshInterval: cfg.JWT.JWKSRefreshInterval,
		ExpirationTime:      cfg.JWT.ExpirationTime,
		RefreshTime:         cfg.JWT.RefreshTime,
		SigningMethod:       cfg.JWT.SigningMethod,
		ContextKey:          cfg.JWT.ContextKey,
		TokenLookup:         cfg.JWT.TokenLookup,
	}, appLogger)

	return &Providers{
		authMiddleware: middleware.NewAuthMiddleware(jwtUtil),
	}, nil
}
