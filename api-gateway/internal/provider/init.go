package provider

import (
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/jwtutils"

	pkgConfig "github.com/raphaeldiscky/go-micro-template/pkg/config"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
)

// Providers holds all initialized providers.
type Providers struct {
	jwtUtil jwtutils.JWTUtil
}

// SetupGlobal initializes and returns the providers.
func SetupGlobal(cfg *config.Config) (*Providers, error) {
	jwtUtil := jwtutils.NewJWTUtil(&pkgConfig.JWTConfig{
		AllowedAlgs:   cfg.JWT.AllowedAlgs,
		Issuer:        cfg.JWT.Issuer,
		SecretKey:     cfg.JWT.SecretKey,
		TokenDuration: cfg.JWT.TokenDuration,
	})

	return &Providers{
		jwtUtil: jwtUtil,
	}, nil
}
