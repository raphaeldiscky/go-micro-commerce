// Package jwtutils provides JWT token generation and validation utilities.
package jwtutils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-template/pkg/config"
)

// Interface defines the methods for JWT utilities.
type Interface interface {
	GenerateAccessToken(userID, email string, roles []string, isActive bool) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error)
	ValidateAccessToken(tokenString string) (*AccessTokenClaims, error)
}

// RefreshTokenClaims represents the claims in a refresh token.
type RefreshTokenClaims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

// AccessTokenClaims represents the claims in an access token.
type AccessTokenClaims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IsActive bool     `json:"is_active"`
	jwt.RegisteredClaims
}

// JWTUtils implements JWTUtilsInterface.
type JWTUtils struct {
	config *config.JWTConfig
}

// NewJWTUtils creates a new JWTUtils instance.
func NewJWTUtils(cfg *config.JWTConfig) Interface {
	return &JWTUtils{
		config: cfg,
	}
}

// GenerateAccessToken generates a JWT access token for the given user.
func (j *JWTUtils) GenerateAccessToken(
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
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.ExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.config.Secret))
}

// GenerateRefreshToken generates a JWT refresh token for the given user.
func (j *JWTUtils) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()

	claims := &RefreshTokenClaims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.RefreshTime)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.config.Secret))
}

// ValidateRefreshToken validates and parses a refresh token.
func (j *JWTUtils) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&RefreshTokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}

			return []byte(j.config.Secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
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
func (j *JWTUtils) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&AccessTokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}

			return []byte(j.config.Secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GetUserIDFromRefreshToken extracts user ID from a refresh token string.
func (j *JWTUtils) GetUserIDFromRefreshToken(tokenString string) (uuid.UUID, error) {
	claims, err := j.ValidateRefreshToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	return userID, nil
}

// GetUserIDFromAccessToken extracts user ID from an access token string.
func (j *JWTUtils) GetUserIDFromAccessToken(tokenString string) (uuid.UUID, error) {
	claims, err := j.ValidateAccessToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	return userID, nil
}
