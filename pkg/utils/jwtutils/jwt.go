// Package jwtutils provides JWT token generation and validation utilities.
package jwtutils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/config"
)

// JWT defines the methods for JWT utilities.
type JWT interface {
	GenerateAccessToken(userID, email string, roles []string, isActive bool) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateRefreshToken(tokenString string) (*refreshTokenClaims, error)
	ValidateAccessToken(tokenString string) (*accessTokenClaims, error)
	GetExpirationTime(tokenString string) (int64, error)
}

// refreshTokenClaims represents the claims in a refresh token.
type refreshTokenClaims struct {
	jwt.RegisteredClaims

	UserID string `json:"user_id"`
	Type   string `json:"type"`
}

// accessTokenClaims represents the claims in an access token.
type accessTokenClaims struct {
	jwt.RegisteredClaims

	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IsActive bool     `json:"is_active"`
}

// jwtUtils implements JWTUtils.
type jwtUtils struct {
	config *config.JWTConfig
}

// NewJWTUtils creates a new jwtUtils instance.
func NewJWTUtils(cfg *config.JWTConfig) JWT {
	return &jwtUtils{
		config: cfg,
	}
}

// GenerateAccessToken generates a JWT access token for the given user.
func (j *jwtUtils) GenerateAccessToken(
	userID, email string,
	roles []string,
	isActive bool,
) (string, error) {
	now := time.Now()

	claims := &accessTokenClaims{
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
func (j *jwtUtils) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()

	claims := &refreshTokenClaims{
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
func (j *jwtUtils) ValidateRefreshToken(tokenString string) (*refreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&refreshTokenClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}

			return []byte(j.config.Secret), nil
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
func (j *jwtUtils) ValidateAccessToken(tokenString string) (*accessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&accessTokenClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}

			return []byte(j.config.Secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*accessTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token test 2")
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
