// Package entity defines the domain entities for the auth service.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system.
type User struct {
	ID                      uuid.UUID  `json:"id"                db:"id"`
	Email                   string     `json:"email"             db:"email"`
	Username                string     `json:"username"          db:"username"`
	PasswordHash            string     `json:"-"                 db:"password_hash"` // Never expose password hash in JSON
	FirstName               string     `json:"first_name"        db:"first_name"`
	LastName                string     `json:"last_name"         db:"last_name"`
	Roles                   []string   `json:"roles"             db:"roles"`
	IsActive                bool       `json:"is_active"         db:"is_active"`
	IsEmailVerified         bool       `json:"is_email_verified" db:"is_email_verified"`
	EmailVerificationToken  *string    `json:"-"                 db:"email_verification_token"` // Never expose in JSON
	EmailVerificationSentAt *time.Time `json:"-"                 db:"email_verification_sent_at"`
	EmailVerifiedAt         *time.Time `json:"email_verified_at" db:"email_verified_at"`
	LastLoginAt             *time.Time `json:"last_login_at"     db:"last_login_at"`
	CreatedAt               time.Time  `json:"created_at"        db:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"        db:"updated_at"`
}

// GetFullName returns the user's full name.
func (u *User) GetFullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}

	return u.FirstName + " " + u.LastName
}

// HasRole checks if the user has a specific role.
func (u *User) HasRole(role string) bool {
	for _, userRole := range u.Roles {
		if userRole == role {
			return true
		}
	}

	return false
}

// AddRole adds a role to the user if not already present.
func (u *User) AddRole(role string) {
	if !u.HasRole(role) {
		u.Roles = append(u.Roles, role)
	}
}

// RemoveRole removes a role from the user.
func (u *User) RemoveRole(role string) {
	for i, userRole := range u.Roles {
		if userRole == role {
			u.Roles = append(u.Roles[:i], u.Roles[i+1:]...)

			break
		}
	}
}
