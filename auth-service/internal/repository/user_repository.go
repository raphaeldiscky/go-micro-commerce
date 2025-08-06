// Package repository defines the repository interfaces for the auth service.
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/entity"
)

// UserRepositoryInterface defines the methods for user repository.
type UserRepositoryInterface interface {
	// User CRUD operations
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByEmailVerificationToken(ctx context.Context, token string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entity.User, error)
	Count(ctx context.Context) (int64, error)

	// User status operations
	ActivateUser(ctx context.Context, id uuid.UUID) error
	DeactivateUser(ctx context.Context, id uuid.UUID) error
	VerifyEmail(ctx context.Context, id uuid.UUID) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// Role operations
	UpdateRoles(ctx context.Context, id uuid.UUID, roles []string) error

	// Email verification operations
	SetEmailVerificationToken(ctx context.Context, id uuid.UUID, token string) error
	ClearEmailVerificationToken(ctx context.Context, id uuid.UUID) error

	// Existence checks
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
}
