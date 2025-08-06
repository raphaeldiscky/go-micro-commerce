// Package repository defines repository interfaces for the auth service.
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/entity"
)

// SessionRepositoryInterface defines the contract for session repository.
type SessionRepositoryInterface interface {
	// Create creates a new session
	Create(ctx context.Context, session *entity.Session) error

	// GetByRefreshToken retrieves a session by refresh token
	GetByRefreshToken(ctx context.Context, refreshToken string) (*entity.Session, error)

	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Session, error)

	// GetActiveSessionsByUserID retrieves all active sessions for a user
	GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error)

	// Update updates a session
	Update(ctx context.Context, session *entity.Session) error

	// Delete deletes a session
	Delete(ctx context.Context, id uuid.UUID) error

	// DeactivateSession deactivates a session
	DeactivateSession(ctx context.Context, id uuid.UUID) error

	// DeactivateAllUserSessions deactivates all sessions for a user
	DeactivateAllUserSessions(ctx context.Context, userID uuid.UUID) error

	// CleanupExpiredSessions removes expired sessions
	CleanupExpiredSessions(ctx context.Context) error

	// TouchSession updates the last_used_at timestamp
	TouchSession(ctx context.Context, id uuid.UUID) error
}
