package dto

import "github.com/google/uuid"

// UserAuthInfo holds user authentication information for the workflow.
type UserAuthInfo struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Roles    []string  `json:"roles"`
	IsActive bool      `json:"is_active"`
}
