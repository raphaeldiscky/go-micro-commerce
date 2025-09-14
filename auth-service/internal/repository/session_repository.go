// Package repository defines repository interfaces for the auth service.
package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/entity"
)

// SessionRepositoryInterface defines the contract for session repository.
type SessionRepositoryInterface interface {
	Create(ctx context.Context, session *entity.Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*entity.Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Session, error)
	GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error)
	Update(ctx context.Context, session *entity.Session) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeactivateSession(ctx context.Context, id uuid.UUID) error
	DeactivateAllUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
	TouchSession(ctx context.Context, id uuid.UUID) error
}

// sessionRepository implements SessionRepositoryInterface using PostgreSQL.
type sessionRepository struct {
	db DBTX
}

// NewSessionRepository creates a new session repository.
func NewSessionRepository(db DBTX) SessionRepositoryInterface {
	return &sessionRepository{db: db}
}

// Create creates a new session.
func (r *sessionRepository) Create(ctx context.Context, session *entity.Session) error {
	query := `
		INSERT INTO sessions (user_id, refresh_token, is_active, ip_address, user_agent, expires_at, created_at, updated_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		session.UserID,
		session.RefreshToken,
		session.IsActive,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.CreatedAt,
		session.UpdatedAt,
		session.LastUsedAt,
	)

	return err
}

// GetByRefreshToken retrieves a session by refresh token.
func (r *sessionRepository) GetByRefreshToken(
	ctx context.Context,
	refreshToken string,
) (*entity.Session, error) {
	session := &entity.Session{}
	query := `
		SELECT id, user_id, refresh_token, is_active, ip_address, user_agent, expires_at, created_at, updated_at, last_used_at
		FROM sessions
		WHERE refresh_token = $1
	`

	row := r.db.QueryRow(ctx, query, refreshToken)
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.IsActive,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.LastUsedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return session, err
}

// GetByID retrieves a session by ID.
func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Session, error) {
	session := &entity.Session{}
	query := `
		SELECT id, user_id, refresh_token, is_active, ip_address, user_agent, expires_at, created_at, updated_at, last_used_at
		FROM sessions
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.IsActive,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.LastUsedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return session, err
}

// GetActiveSessionsByUserID retrieves all active sessions for a user.
func (r *sessionRepository) GetActiveSessionsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*entity.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, is_active, ip_address, user_agent, expires_at, created_at, updated_at, last_used_at
		FROM sessions
		WHERE user_id = $1 AND is_active = true AND expires_at > NOW()
		ORDER BY last_used_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*entity.Session

	for rows.Next() {
		session := &entity.Session{}

		err = rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshToken,
			&session.IsActive,
			&session.IPAddress,
			&session.UserAgent,
			&session.ExpiresAt,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.LastUsedAt,
		)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// Update updates a session.
func (r *sessionRepository) Update(ctx context.Context, session *entity.Session) error {
	query := `
		UPDATE sessions
		SET is_active = $2, ip_address = $3, user_agent = $4, expires_at = $5, updated_at = $6, last_used_at = $7
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.IsActive,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.UpdatedAt,
		session.LastUsedAt,
	)

	return err
}

// Delete deletes a session.
func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)

	return err
}

// DeactivateSession deactivates a session.
func (r *sessionRepository) DeactivateSession(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)

	return err
}

// DeactivateAllUserSessions deactivates all sessions for a user.
func (r *sessionRepository) DeactivateAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE sessions SET is_active = false, updated_at = NOW() WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)

	return err
}

// CleanupExpiredSessions removes expired sessions.
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	_, err := r.db.Exec(ctx, query)

	return err
}

// TouchSession updates the last_used_at timestamp.
func (r *sessionRepository) TouchSession(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET last_used_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)

	return err
}
