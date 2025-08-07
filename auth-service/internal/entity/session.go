// Package entity defines the domain entities for the auth service.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a user session in the system.
type Session struct {
	ID           uuid.UUID `json:"id"           db:"id"`
	UserID       uuid.UUID `json:"user_id"      db:"user_id"`
	RefreshToken string    `json:"-"            db:"refresh_token"` // Never expose in JSON
	IsActive     bool      `json:"is_active"    db:"is_active"`
	IPAddress    string    `json:"ip_address"   db:"ip_address"`
	UserAgent    string    `json:"user_agent"   db:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at"   db:"expires_at"`
	CreatedAt    time.Time `json:"created_at"   db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"   db:"updated_at"`
	LastUsedAt   time.Time `json:"last_used_at" db:"last_used_at"`
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (active and not expired).
func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// Touch updates the last used timestamp.
func (s *Session) Touch() {
	s.LastUsedAt = time.Now()
}

// Deactivate marks the session as inactive.
func (s *Session) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}
