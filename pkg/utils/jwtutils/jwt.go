// Package jwtutils provides utilities for working with JSON Web Tokens (JWT).
package jwtutils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-template/pkg/config"
)

var (
	// ErrInvalidToken indicates that the token is not valid.
	ErrInvalidToken = errors.New("token is not valid")
	// ErrTokenExpired indicates that the token has expired.
	ErrTokenExpired = errors.New("token has expired")
	// ErrInvalidSignature indicates that the token signature is invalid.
	ErrInvalidSignature = errors.New("token signature is invalid")
	// ErrMissingClaims indicates that the token claims are missing or invalid.
	ErrMissingClaims = errors.New("token claims are missing or invalid")
)

// JWTUtil is an interface for JWT token generation and parsing.
type JWTUtil interface {
	Sign(payload *JWTPayload) (string, error)
	Parse(token string) (*JWTClaims, error)
	Validate(token string) error // Additional validation method
}

// JWTClaims represents the claims contained within a JWT.
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}

// JWTPayload represents the payload contained within a JWT.
type JWTPayload struct {
	UserID int64
	Email  string
}

// jwtUtil is a concrete implementation of the JWTUtil interface.
type jwtUtil struct {
	config *config.JWTConfig
}

// NewJWTUtil creates a new instance of JWTUtil.
func NewJWTUtil(cfg *config.JWTConfig) JWTUtil {
	// Add basic validation for config
	if cfg == nil {
		panic("JWT config cannot be nil")
	}

	if cfg.SecretKey == "" {
		panic("JWT secret key cannot be empty")
	}

	if cfg.TokenDuration <= 0 {
		panic("JWT token duration must be positive")
	}

	return &jwtUtil{
		config: cfg,
	}
}

// Sign generates a new JWT token for the given payload.
func (j *jwtUtil) Sign(payload *JWTPayload) (string, error) {
	if payload == nil {
		return "", errors.New("payload cannot be nil")
	}

	currentTime := time.Now()
	expirationTime := currentTime.Add(time.Duration(j.config.TokenDuration) * time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		UserID: payload.UserID,
		Email:  payload.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			NotBefore: jwt.NewNumericDate(currentTime), // Token is valid from now
			Issuer:    j.config.Issuer,
		},
	})

	signedToken, err := token.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return signedToken, nil
}

// Parse parses the JWT token and returns the claims.
func (j *jwtUtil) Parse(token string) (*JWTClaims, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods(j.config.AllowedAlgs),
		jwt.WithIssuer(j.config.Issuer),
		jwt.WithIssuedAt(),
	)

	return j.parseClaims(parser, token)
}

// Validate validates the JWT token without returning claims.
func (j *jwtUtil) Validate(token string) error {
	_, err := j.Parse(token)

	return err
}

// parseClaims parses the JWT token and returns the claims.
func (j *jwtUtil) parseClaims(parser *jwt.Parser, token string) (*JWTClaims, error) {
	parsedToken, err := parser.ParseWithClaims(
		token,
		&JWTClaims{},
		func(t *jwt.Token) (interface{}, error) {
			// Verify the signing method
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(j.config.SecretKey), nil
		},
	)
	if err != nil {
		// Provide more specific error messages
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrInvalidSignature
		}

		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(*JWTClaims)
	if !ok || !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	// Additional validation
	if claims.UserID <= 0 {
		return nil, ErrMissingClaims
	}

	if claims.Email == "" {
		return nil, ErrMissingClaims
	}

	return claims, nil
}
