// Package repository defines the repository interfaces for the auth service.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/entity"
)

// AddressRepository defines the methods for address repository.
type AddressRepository interface {
	// Address CRUD operations
	Create(ctx context.Context, address *entity.Address) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Address, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Address, error)
	GetByUserIDWithCursor(
		ctx context.Context,
		userID uuid.UUID,
		limit int64,
		cursorID string,
		cursorTimestamp int64,
		cursorIsDefault string,
	) ([]*entity.Address, error)
	GetDefaultByUserID(ctx context.Context, userID uuid.UUID) (*entity.Address, error)
	Update(ctx context.Context, address *entity.Address) (*entity.Address, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Default address operations
	SetDefault(ctx context.Context, userID, addressID uuid.UUID) error
	UnsetAllDefaults(ctx context.Context, userID uuid.UUID) error

	// Count and existence checks
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}

// addressRepository implements AddressRepository using PostgreSQL.
type addressRepository struct {
	db DBTX
}

// NewAddressRepository creates a new AddressRepository.
func NewAddressRepository(db DBTX) AddressRepository {
	return &addressRepository{db: db}
}

// Create creates a new address.
func (r *addressRepository) Create(ctx context.Context, address *entity.Address) error {
	query := `
		INSERT INTO user_addresses (
			id, user_id, receiver_name, address_line1, address_line2,
			city, state, postal_code, country_code, latitude, longitude,
			is_default, note, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)`

	_, err := r.db.Exec(ctx, query,
		address.ID, address.UserID, address.ReceiverName, address.AddressLine1,
		address.AddressLine2, address.City, address.State, address.PostalCode,
		address.CountryCode, address.Latitude, address.Longitude, address.IsDefault,
		address.Note, address.CreatedAt, address.UpdatedAt,
	)

	return err
}

// GetByID retrieves an address by ID.
func (r *addressRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Address, error) {
	address := &entity.Address{}
	query := `
		SELECT id, user_id, receiver_name, address_line1, address_line2,
		       city, state, postal_code, country_code, latitude, longitude,
		       is_default, note, created_at, updated_at
		FROM user_addresses
		WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&address.ID, &address.UserID, &address.ReceiverName, &address.AddressLine1,
		&address.AddressLine2, &address.City, &address.State, &address.PostalCode,
		&address.CountryCode, &address.Latitude, &address.Longitude, &address.IsDefault,
		&address.Note, &address.CreatedAt, &address.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, err
	}

	return address, nil
}

// GetByUserID retrieves all addresses for a user.
func (r *addressRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*entity.Address, error) {
	query := `
		SELECT id, user_id, receiver_name, address_line1, address_line2,
		       city, state, postal_code, country_code, latitude, longitude,
		       is_default, note, created_at, updated_at
		FROM user_addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []*entity.Address

	for rows.Next() {
		address := &entity.Address{}

		err = rows.Scan(
			&address.ID, &address.UserID, &address.ReceiverName, &address.AddressLine1,
			&address.AddressLine2, &address.City, &address.State, &address.PostalCode,
			&address.CountryCode, &address.Latitude, &address.Longitude, &address.IsDefault,
			&address.Note, &address.CreatedAt, &address.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetByUserIDWithCursor retrieves addresses for a user with cursor-based pagination.
func (r *addressRepository) GetByUserIDWithCursor(
	ctx context.Context,
	userID uuid.UUID,
	limit int64,
	cursorID string,
	cursorTimestamp int64,
	cursorIsDefault string,
) ([]*entity.Address, error) {
	var (
		query string
		rows  pgx.Rows
		err   error
	)

	if cursorID == "" {
		// First page: just order and limit
		query = `
			SELECT id, user_id, receiver_name, address_line1, address_line2,
			       city, state, postal_code, country_code, latitude, longitude,
			       is_default, note, created_at, updated_at
			FROM user_addresses
			WHERE user_id = $1
			ORDER BY is_default DESC, created_at DESC, id DESC
			LIMIT $2`

		rows, err = r.db.Query(ctx, query, userID, limit)
	} else {
		// Subsequent pages: use keyset pagination with composite ordering
		cursorUUID, parseErr := uuid.Parse(cursorID)
		if parseErr != nil {
			return nil, errors.New("invalid cursor ID")
		}

		// Convert cursor is_default string to boolean
		cursorBool := cursorIsDefault == "true"

		// Keyset pagination with composite key (is_default, created_at, id)
		// We need to handle the case where is_default might change between pages
		query = `
			SELECT id, user_id, receiver_name, address_line1, address_line2,
			       city, state, postal_code, country_code, latitude, longitude,
			       is_default, note, created_at, updated_at
			FROM user_addresses
			WHERE user_id = $1
			  AND (
			    is_default < $2
			    OR (is_default = $2 AND created_at < to_timestamp($3))
			    OR (is_default = $2 AND created_at = to_timestamp($3) AND id < $4)
			  )
			ORDER BY is_default DESC, created_at DESC, id DESC
			LIMIT $5`

		rows, err = r.db.Query(ctx, query, userID, cursorBool, cursorTimestamp, cursorUUID, limit)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	addresses := make([]*entity.Address, 0, limit)

	for rows.Next() {
		address := &entity.Address{}

		err = rows.Scan(
			&address.ID, &address.UserID, &address.ReceiverName, &address.AddressLine1,
			&address.AddressLine2, &address.City, &address.State, &address.PostalCode,
			&address.CountryCode, &address.Latitude, &address.Longitude, &address.IsDefault,
			&address.Note, &address.CreatedAt, &address.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, address)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

// GetDefaultByUserID retrieves the default address for a user.
func (r *addressRepository) GetDefaultByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*entity.Address, error) {
	address := &entity.Address{}
	query := `
		SELECT id, user_id, receiver_name, address_line1, address_line2,
		       city, state, postal_code, country_code, latitude, longitude,
		       is_default, note, created_at, updated_at
		FROM user_addresses
		WHERE user_id = $1 AND is_default = true
		LIMIT 1`

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&address.ID, &address.UserID, &address.ReceiverName, &address.AddressLine1,
		&address.AddressLine2, &address.City, &address.State, &address.PostalCode,
		&address.CountryCode, &address.Latitude, &address.Longitude, &address.IsDefault,
		&address.Note, &address.CreatedAt, &address.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, err
	}

	return address, nil
}

// Update updates an address.
func (r *addressRepository) Update(
	ctx context.Context,
	address *entity.Address,
) (*entity.Address, error) {
	query := `
		UPDATE user_addresses SET
			receiver_name = $2, address_line1 = $3, address_line2 = $4,
			city = $5, state = $6, postal_code = $7, country_code = $8,
			latitude = $9, longitude = $10, note = $11, updated_at = $12
		WHERE id = $1
		RETURNING id, user_id, receiver_name, address_line1, address_line2,
				  city, state, postal_code, country_code, latitude, longitude,
				  is_default, note, created_at, updated_at`

	updatedAddress := &entity.Address{}

	err := r.db.QueryRow(ctx, query,
		address.ID, address.ReceiverName, address.AddressLine1,
		address.AddressLine2, address.City, address.State, address.PostalCode,
		address.CountryCode, address.Latitude, address.Longitude,
		address.Note, address.UpdatedAt,
	).Scan(
		&updatedAddress.ID, &updatedAddress.UserID, &updatedAddress.ReceiverName,
		&updatedAddress.AddressLine1, &updatedAddress.AddressLine2,
		&updatedAddress.City, &updatedAddress.State, &updatedAddress.PostalCode,
		&updatedAddress.CountryCode, &updatedAddress.Latitude, &updatedAddress.Longitude,
		&updatedAddress.IsDefault, &updatedAddress.Note,
		&updatedAddress.CreatedAt, &updatedAddress.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, err
	}

	return updatedAddress, nil
}

// Delete deletes an address by ID.
func (r *addressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM user_addresses WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// SetDefault atomically sets an address as default and unsets all other defaults for the user.
// Uses a single UPDATE with CASE to ensure atomicity without explicit transaction handling.
func (r *addressRepository) SetDefault(ctx context.Context, userID, addressID uuid.UUID) error {
	now := time.Now()

	// Use a single UPDATE with CASE to atomically set/unset defaults
	// This is atomic because it's a single query
	query := `
		UPDATE user_addresses
		SET is_default = CASE WHEN id = $1 THEN true ELSE false END,
		    updated_at = $2
		WHERE user_id = $3
		RETURNING id`

	rows, err := r.db.Query(ctx, query, addressID, now, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Check if the target address was found and updated
	found := false

	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return err
		}

		if id == addressID {
			found = true
		}
	}

	if !found {
		return sql.ErrNoRows
	}

	return nil
}

// UnsetAllDefaults unsets all default addresses for a user.
func (r *addressRepository) UnsetAllDefaults(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE user_addresses
		SET is_default = false, updated_at = $1
		WHERE user_id = $2`

	_, err := r.db.Exec(ctx, query, time.Now(), userID)

	return err
}

// CountByUserID returns the number of addresses for a user.
func (r *addressRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64

	query := `SELECT COUNT(*) FROM user_addresses WHERE user_id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)

	return count, err
}

// ExistsByID checks if an address exists by ID.
func (r *addressRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM user_addresses WHERE id = $1)`
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)

	return exists, err
}
