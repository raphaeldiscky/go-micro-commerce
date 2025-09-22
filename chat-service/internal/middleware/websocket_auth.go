package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// WebSocketAuth represents authentication information extracted from the request.
type WebSocketAuth struct {
	UserID   uuid.UUID
	Email    string
	UserType constant.UserType
	IsActive bool
}

// AuthenticateWebSocket extracts authentication information from WebSocket request headers.
func AuthenticateWebSocket(r *http.Request) (*WebSocketAuth, error) {
	userID, err := extractUserID(r)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	email := extractEmail(r)
	if email == "" {
		return nil, errors.New("missing email header")
	}

	userType, err := extractUserType(r)
	if err != nil {
		return nil, fmt.Errorf("invalid user type: %w", err)
	}

	isActive, err := extractIsActive(r)
	if err != nil {
		return nil, fmt.Errorf("invalid is_active header: %w", err)
	}

	if !isActive {
		return nil, errors.New("user account is inactive")
	}

	return &WebSocketAuth{
		UserID:   userID,
		Email:    email,
		UserType: userType,
		IsActive: isActive,
	}, nil
}

// extractUserID extracts and validates the user ID from request headers.
func extractUserID(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.Header.Get(pkgconstant.XUserID)
	if userIDStr == "" {
		return uuid.Nil, fmt.Errorf("missing %s header", pkgconstant.XUserID)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	return userID, nil
}

// extractEmail extracts the email from request headers.
func extractEmail(r *http.Request) string {
	return r.Header.Get(pkgconstant.XEmail)
}

// extractUserType extracts and validates the user type from request headers.
func extractUserType(r *http.Request) (constant.UserType, error) {
	rolesStr := r.Header.Get(pkgconstant.XRoles)
	if rolesStr == "" {
		return "", fmt.Errorf("missing %s header", pkgconstant.XRoles)
	}

	roles := strings.Split(rolesStr, ",")
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "admin" || role == "support" {
			return constant.UserTypeAdmin, nil
		}
	}

	return constant.UserTypeUser, nil
}

// extractIsActive extracts and validates the is_active flag from request headers.
func extractIsActive(r *http.Request) (bool, error) {
	isActiveStr := r.Header.Get(pkgconstant.XIsActive)
	if isActiveStr == "" {
		return false, fmt.Errorf("missing %s header", pkgconstant.XIsActive)
	}

	return isActiveStr == "true", nil
}

// RequireUserType returns a middleware function that checks if the user has the required type.
func RequireUserType(requiredType constant.UserType) func(*WebSocketAuth) error {
	return func(auth *WebSocketAuth) error {
		if auth.UserType != requiredType {
			return fmt.Errorf(
				"access denied: user type %s required, got %s",
				requiredType,
				auth.UserType,
			)
		}

		return nil
	}
}

// RequireActiveUser returns a middleware function that checks if the user is active.
func RequireActiveUser() func(*WebSocketAuth) error {
	return func(auth *WebSocketAuth) error {
		if !auth.IsActive {
			return errors.New("access denied: user account is inactive")
		}

		return nil
	}
}
