// Package jwtutils provides JWT token generation and validation utilities.
package jwtutils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/config"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/jwks"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// JWT defines the methods for JWT utilities.
type JWT interface {
	GenerateAccessToken(userID, email string, roles []string, isActive bool) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateRefreshToken(tokenString string) (*refreshTokenClaims, error)
	ValidateAccessToken(tokenString string) (*AccessTokenClaims, error)
	GetExpirationTime(tokenString string) (int64, error)
	GetPublicKey() *rsa.PublicKey
}

// refreshTokenClaims represents the claims in a refresh token.
type refreshTokenClaims struct {
	jwt.RegisteredClaims

	UserID string `json:"user_id"`
	Type   string `json:"type"`
}

// AccessTokenClaims represents the claims in an access token.
type AccessTokenClaims struct {
	jwt.RegisteredClaims

	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IsActive bool     `json:"is_active"`
}

// jwtUtils implements JWTUtils.
type jwtUtils struct {
	config      *config.JWTConfig
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
	jwksFetcher *jwks.Fetcher
	useJWKS     bool
	logger      logger.Logger
}

// NewJWTUtils creates a new jwtUtils instance.
func NewJWTUtils(cfg *config.JWTConfig, logger logger.Logger) JWT {
	var (
		publicKey   *rsa.PublicKey
		privateKey  *rsa.PrivateKey
		jwksFetcher *jwks.Fetcher
		useJWKS     bool
	)

	// Priority 1: Try JWKS if URL is provided
	if cfg.JWKSUrl != "" {
		jwksFetcher = jwks.NewFetcher(
			cfg.JWKSUrl,
			cfg.JWKSCacheTTL,
			cfg.JWKSRefreshInterval,
			logger,
		)

		// Try to get initial key from JWKS
		key, err := jwksFetcher.GetPublicKey()
		if err == nil {
			publicKey = key
			useJWKS = true
		}
		// If JWKS fails, fall through to file-based keys
	}

	// Priority 2: File-based keys (fallback or when JWKS not configured)
	if !useJWKS && cfg.PublicKeyPath != "" {
		var err error

		publicKey, err = loadPublicKey(cfg.PublicKeyPath)
		if err != nil {
			panic("failed to load public key: " + err.Error())
		}
	}

	// Load RSA private key only if path is provided (required for signing)
	if cfg.PrivateKeyPath != "" {
		var err error

		privateKey, err = loadPrivateKey(cfg.PrivateKeyPath)
		if err != nil {
			panic("failed to load private key: " + err.Error())
		}
	}

	// Ensure we have at least a public key
	if publicKey == nil && privateKey == nil {
		panic("no public or private key available")
	}

	return &jwtUtils{
		config:      cfg,
		privateKey:  privateKey,
		publicKey:   publicKey,
		jwksFetcher: jwksFetcher,
		useJWKS:     useJWKS,
		logger:      logger,
	}
}

// loadPrivateKey loads an RSA private key from a PEM file in PKCS8 format.
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsaKey, nil
}

// loadPublicKey loads an RSA public key from a PEM file.
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPublicKey, nil
}

// GenerateAccessToken generates a JWT access token for the given user.
func (j *jwtUtils) GenerateAccessToken(
	userID, email string,
	roles []string,
	isActive bool,
) (string, error) {
	now := time.Now()

	claims := &AccessTokenClaims{
		UserID:   userID,
		Email:    email,
		Roles:    roles,
		IsActive: isActive,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.ExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(j.privateKey)
}

// GenerateRefreshToken generates a JWT refresh token for the given user.
func (j *jwtUtils) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()

	claims := &refreshTokenClaims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.RefreshTime)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(j.privateKey)
}

// ValidateRefreshToken validates and parses a refresh token.
func (j *jwtUtils) ValidateRefreshToken(tokenString string) (*refreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&refreshTokenClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("invalid signing method")
			}

			// Use GetPublicKey() to leverage JWKS cache
			return j.GetPublicKey(), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*refreshTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Ensure it's actually a refresh token
	if claims.Type != "refresh" {
		return nil, errors.New("not a refresh token")
	}

	return claims, nil
}

// ValidateAccessToken validates and parses an access token.
func (j *jwtUtils) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&AccessTokenClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("invalid signing method")
			}

			// Use GetPublicKey() to leverage JWKS cache
			return j.GetPublicKey(), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		j.logger.Warnf("Invalid access token: %v", err)
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GetUserIDFromRefreshToken extracts user ID from a refresh token string.
func (j *jwtUtils) GetUserIDFromRefreshToken(tokenString string) (uuid.UUID, error) {
	claims, err := j.ValidateRefreshToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		j.logger.Warnf("Invalid user ID in refresh token: %v", err)
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	return userID, nil
}

// GetUserIDFromAccessToken extracts user ID from an access token string.
func (j *jwtUtils) GetUserIDFromAccessToken(tokenString string) (uuid.UUID, error) {
	claims, err := j.ValidateAccessToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		j.logger.Warnf("Invalid user ID in access token: %v", err)
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	return userID, nil
}

// GetExpirationTime extracts the expiration time from a token string.
func (j *jwtUtils) GetExpirationTime(tokenString string) (int64, error) {
	claims, err := j.ValidateAccessToken(tokenString)
	if err != nil {
		return 0, err
	}

	return int64(time.Until(claims.ExpiresAt.Time).Seconds()), nil
}

// GetPublicKey returns the RSA public key.
func (j *jwtUtils) GetPublicKey() *rsa.PublicKey {
	// If using JWKS, try to get fresh key from cache
	if j.useJWKS && j.jwksFetcher != nil {
		j.logger.Info("Fetching public key from JWKS")

		if key, err := j.jwksFetcher.GetPublicKey(); err == nil {
			return key
		}
		// If JWKS fetch fails, fall back to initial cached key
	}

	return j.publicKey
}
