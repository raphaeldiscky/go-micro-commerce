package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/domain/entities"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/domain/repositories"
)

// SellerRepositoryPostgres implements the SellerRepository interface using PostgreSQL
type SellerRepositoryPostgres struct {
	db *pgxpool.Pool
}

// NewSellerRepositoryPostgres creates a new instance of SellerRepositoryPostgres
func NewSellerRepositoryPostgres(db *pgxpool.Pool) repositories.SellerRepository {
	return &SellerRepositoryPostgres{db: db}
}

// Create creates a new seller
func (r *SellerRepositoryPostgres) Create(ctx context.Context, seller *entities.Seller) (*entities.Seller, error) {
	query := `
		INSERT INTO sellers (id, name, email, phone, address, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, email, phone, address, is_active, created_at, updated_at`

	row := r.db.QueryRow(ctx, query,
		seller.Id,
		seller.Name,
		seller.Email,
		seller.Phone,
		seller.Address,
		seller.IsActive,
		seller.CreatedAt,
		seller.UpdatedAt,
	)

	var createdSeller entities.Seller
	err := row.Scan(
		&createdSeller.Id,
		&createdSeller.Name,
		&createdSeller.Email,
		&createdSeller.Phone,
		&createdSeller.Address,
		&createdSeller.IsActive,
		&createdSeller.CreatedAt,
		&createdSeller.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create seller: %w", err)
	}

	return &createdSeller, nil
}

// FindById finds a seller by ID
func (r *SellerRepositoryPostgres) FindById(ctx context.Context, id uuid.UUID) (*entities.Seller, error) {
	query := `
		SELECT id, name, email, phone, address, is_active, created_at, updated_at
		FROM sellers
		WHERE id = $1`

	row := r.db.QueryRow(ctx, query, id)

	var seller entities.Seller
	err := row.Scan(
		&seller.Id,
		&seller.Name,
		&seller.Email,
		&seller.Phone,
		&seller.Address,
		&seller.IsActive,
		&seller.CreatedAt,
		&seller.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find seller by id: %w", err)
	}

	return &seller, nil
}

// FindByEmail finds a seller by email
func (r *SellerRepositoryPostgres) FindByEmail(ctx context.Context, email string) (*entities.Seller, error) {
	query := `
		SELECT id, name, email, phone, address, is_active, created_at, updated_at
		FROM sellers
		WHERE email = $1`

	row := r.db.QueryRow(ctx, query, email)

	var seller entities.Seller
	err := row.Scan(
		&seller.Id,
		&seller.Name,
		&seller.Email,
		&seller.Phone,
		&seller.Address,
		&seller.IsActive,
		&seller.CreatedAt,
		&seller.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find seller by email: %w", err)
	}

	return &seller, nil
}

// FindAll finds all sellers with pagination
func (r *SellerRepositoryPostgres) FindAll(ctx context.Context, limit, offset int) ([]*entities.Seller, error) {
	query := `
		SELECT id, name, email, phone, address, is_active, created_at, updated_at
		FROM sellers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find all sellers: %w", err)
	}
	defer rows.Close()

	var sellers []*entities.Seller
	for rows.Next() {
		var seller entities.Seller
		err := rows.Scan(
			&seller.Id,
			&seller.Name,
			&seller.Email,
			&seller.Phone,
			&seller.Address,
			&seller.IsActive,
			&seller.CreatedAt,
			&seller.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan seller: %w", err)
		}
		sellers = append(sellers, &seller)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sellers: %w", err)
	}

	return sellers, nil
}

// FindActive finds all active sellers with pagination
func (r *SellerRepositoryPostgres) FindActive(ctx context.Context, limit, offset int) ([]*entities.Seller, error) {
	query := `
		SELECT id, name, email, phone, address, is_active, created_at, updated_at
		FROM sellers
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find active sellers: %w", err)
	}
	defer rows.Close()

	var sellers []*entities.Seller
	for rows.Next() {
		var seller entities.Seller
		err := rows.Scan(
			&seller.Id,
			&seller.Name,
			&seller.Email,
			&seller.Phone,
			&seller.Address,
			&seller.IsActive,
			&seller.CreatedAt,
			&seller.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan seller: %w", err)
		}
		sellers = append(sellers, &seller)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over active sellers: %w", err)
	}

	return sellers, nil
}

// Update updates an existing seller
func (r *SellerRepositoryPostgres) Update(ctx context.Context, seller *entities.Seller) (*entities.Seller, error) {
	query := `
		UPDATE sellers
		SET name = $2, email = $3, phone = $4, address = $5, is_active = $6, updated_at = $7
		WHERE id = $1
		RETURNING id, name, email, phone, address, is_active, created_at, updated_at`

	row := r.db.QueryRow(ctx, query,
		seller.Id,
		seller.Name,
		seller.Email,
		seller.Phone,
		seller.Address,
		seller.IsActive,
		seller.UpdatedAt,
	)

	var updatedSeller entities.Seller
	err := row.Scan(
		&updatedSeller.Id,
		&updatedSeller.Name,
		&updatedSeller.Email,
		&updatedSeller.Phone,
		&updatedSeller.Address,
		&updatedSeller.IsActive,
		&updatedSeller.CreatedAt,
		&updatedSeller.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("seller not found")
		}
		return nil, fmt.Errorf("failed to update seller: %w", err)
	}

	return &updatedSeller, nil
}

// Delete deletes a seller by ID
func (r *SellerRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sellers WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete seller: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("seller not found")
	}

	return nil
}

// Exists checks if a seller exists by ID
func (r *SellerRepositoryPostgres) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM sellers WHERE id = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check seller existence: %w", err)
	}

	return exists, nil
}

// ExistsByEmail checks if a seller exists by email
func (r *SellerRepositoryPostgres) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM sellers WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check seller email existence: %w", err)
	}

	return exists, nil
}

// Count returns the total number of sellers
func (r *SellerRepositoryPostgres) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM sellers`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sellers: %w", err)
	}

	return count, nil
}

// CountActive returns the total number of active sellers
func (r *SellerRepositoryPostgres) CountActive(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM sellers WHERE is_active = true`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active sellers: %w", err)
	}

	return count, nil
}
