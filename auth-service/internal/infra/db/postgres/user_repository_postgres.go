// Package postgres provides PostgreSQL implementation of repositories.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/repository"
)

// UserRepositoryPostgres implements UserRepositoryInterface using PostgreSQL.
type UserRepositoryPostgres struct {
	db *pgxpool.Pool
}

// NewUserRepositoryPostgres creates a new UserRepositoryPostgres.
func NewUserRepositoryPostgres(db *pgxpool.Pool) repository.UserRepositoryInterface {
	return &UserRepositoryPostgres{db: db}
}

// Create creates a new user.
func (r *UserRepositoryPostgres) Create(ctx context.Context, user *entity.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (
			id, email, username, password_hash, first_name, last_name, 
			roles, is_active, is_email_verified, email_verification_token,
			email_verification_sent_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash,
		user.FirstName, user.LastName, pq.Array(user.Roles), user.IsActive,
		user.IsEmailVerified, user.EmailVerificationToken,
		user.EmailVerificationSentAt, user.CreatedAt, user.UpdatedAt,
	)

	return err
}

// GetByID retrieves a user by ID.
func (r *UserRepositoryPostgres) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT id, email, username, password_hash, first_name, last_name,
		       roles, is_active, is_email_verified, email_verification_token,
		       email_verification_sent_at, email_verified_at, last_login_at,
		       created_at, updated_at
		FROM users WHERE id = $1`

	var roles pq.StringArray

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &roles, &user.IsActive,
		&user.IsEmailVerified, &user.EmailVerificationToken,
		&user.EmailVerificationSentAt, &user.EmailVerifiedAt,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, err
	}

	user.Roles = []string(roles)

	return user, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepositoryPostgres) GetByEmail(
	ctx context.Context,
	email string,
) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT id, email, username, password_hash, first_name, last_name,
		       roles, is_active, is_email_verified, email_verification_token,
		       email_verification_sent_at, email_verified_at, last_login_at,
		       created_at, updated_at
		FROM users WHERE email = $1`

	var roles pq.StringArray

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &roles, &user.IsActive,
		&user.IsEmailVerified, &user.EmailVerificationToken,
		&user.EmailVerificationSentAt, &user.EmailVerifiedAt,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, err
	}

	user.Roles = []string(roles)

	return user, nil
}

// GetByUsername retrieves a user by username.
func (r *UserRepositoryPostgres) GetByUsername(
	ctx context.Context,
	username string,
) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT id, email, username, password_hash, first_name, last_name,
		       roles, is_active, is_email_verified, email_verification_token,
		       email_verification_sent_at, email_verified_at, last_login_at,
		       created_at, updated_at
		FROM users WHERE username = $1`

	var roles pq.StringArray

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &roles, &user.IsActive,
		&user.IsEmailVerified, &user.EmailVerificationToken,
		&user.EmailVerificationSentAt, &user.EmailVerifiedAt,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, err
	}

	user.Roles = []string(roles)

	return user, nil
}

// GetByEmailVerificationToken retrieves a user by email verification token.
func (r *UserRepositoryPostgres) GetByEmailVerificationToken(
	ctx context.Context,
	token string,
) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT id, email, username, password_hash, first_name, last_name,
		       roles, is_active, is_email_verified, email_verification_token,
		       email_verification_sent_at, email_verified_at, last_login_at,
		       created_at, updated_at
		FROM users WHERE email_verification_token = $1`

	var roles pq.StringArray

	err := r.db.QueryRow(ctx, query, token).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &roles, &user.IsActive,
		&user.IsEmailVerified, &user.EmailVerificationToken,
		&user.EmailVerificationSentAt, &user.EmailVerifiedAt,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, err
	}

	user.Roles = []string(roles)

	return user, nil
}

// Update updates a user.
func (r *UserRepositoryPostgres) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET
			email = $2, username = $3, password_hash = $4,
			first_name = $5, last_name = $6, roles = $7,
			is_active = $8, is_email_verified = $9,
			email_verification_token = $10, email_verification_sent_at = $11,
			email_verified_at = $12, last_login_at = $13, updated_at = $14
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash,
		user.FirstName, user.LastName, pq.Array(user.Roles),
		user.IsActive, user.IsEmailVerified, user.EmailVerificationToken,
		user.EmailVerificationSentAt, user.EmailVerifiedAt,
		user.LastLoginAt, user.UpdatedAt,
	)

	return err
}

// Delete deletes a user by ID.
func (r *UserRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)

	return err
}

// List retrieves a list of users with pagination.
func (r *UserRepositoryPostgres) List(
	ctx context.Context,
	limit, offset int,
) ([]*entity.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, last_name,
		       roles, is_active, is_email_verified, email_verification_token,
		       email_verification_sent_at, email_verified_at, last_login_at,
		       created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User

	for rows.Next() {
		user := &entity.User{}

		var roles pq.StringArray

		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.PasswordHash,
			&user.FirstName, &user.LastName, &roles, &user.IsActive,
			&user.IsEmailVerified, &user.EmailVerificationToken,
			&user.EmailVerificationSentAt, &user.EmailVerifiedAt,
			&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		user.Roles = []string(roles)
		users = append(users, user)
	}

	return users, nil
}

// Count returns the total number of users.
func (r *UserRepositoryPostgres) Count(ctx context.Context) (int64, error) {
	var count int64

	query := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRow(ctx, query).Scan(&count)

	return count, err
}

// ActivateUser activates a user.
func (r *UserRepositoryPostgres) ActivateUser(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET is_active = true, updated_at = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, time.Now())

	return err
}

// DeactivateUser deactivates a user.
func (r *UserRepositoryPostgres) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET is_active = false, updated_at = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, time.Now())

	return err
}

// VerifyEmail verifies a user's email.
func (r *UserRepositoryPostgres) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users SET 
			is_email_verified = true, 
			email_verified_at = $2, 
			email_verification_token = NULL,
			updated_at = $2
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, time.Now())

	return err
}

// UpdateLastLogin updates the last login time.
func (r *UserRepositoryPostgres) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $2, updated_at = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, time.Now())

	return err
}

// UpdateRoles updates user roles.
func (r *UserRepositoryPostgres) UpdateRoles(
	ctx context.Context,
	id uuid.UUID,
	roles []string,
) error {
	query := `UPDATE users SET roles = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, pq.Array(roles), time.Now())

	return err
}

// SetEmailVerificationToken sets the email verification token.
func (r *UserRepositoryPostgres) SetEmailVerificationToken(
	ctx context.Context,
	id uuid.UUID,
	token string,
) error {
	query := `
		UPDATE users SET 
			email_verification_token = $2, 
			email_verification_sent_at = $3,
			updated_at = $3
		WHERE id = $1`
	now := time.Now()
	_, err := r.db.Exec(ctx, query, id, token, now)

	return err
}

// ClearEmailVerificationToken clears the email verification token.
func (r *UserRepositoryPostgres) ClearEmailVerificationToken(
	ctx context.Context,
	id uuid.UUID,
) error {
	query := `
		UPDATE users SET 
			email_verification_token = NULL,
			email_verification_sent_at = NULL,
			updated_at = $2
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, time.Now())

	return err
}

// EmailExists checks if an email already exists.
func (r *UserRepositoryPostgres) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)

	return exists, err
}

// UsernameExists checks if a username already exists.
func (r *UserRepositoryPostgres) UsernameExists(
	ctx context.Context,
	username string,
) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	err := r.db.QueryRow(ctx, query, username).Scan(&exists)

	return exists, err
}
