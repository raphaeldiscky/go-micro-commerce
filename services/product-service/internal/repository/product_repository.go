// Package repository defines the interface for product data operations.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	entity "github.com/raphaeldiscky/go-micro-template/services/product-service/internal/entity"
)

type ProductRepositoryPostgres struct {
	db *pgxpool.Pool
}

// NewProductRepository creates a new instance of ProductRepositoryPostgres.
func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &ProductRepositoryPostgres{
		db: db,
	}
}

// ProductRepository defines the interface for product data operations.
type ProductRepository interface {
	// Create saves a new product
	Create(ctx context.Context, product *entity.Product) (*entity.Product, error)

	// FindByID retrieves a product by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)

	// FindAll retrieves all products with optional pagination
	FindAll(ctx context.Context, limit, offset int) ([]*entity.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *entity.Product) (*entity.Product, error)

	// Delete removes a product by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if a product exists by ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// Count returns the total number of products
	Count(ctx context.Context) (int64, error)
}

// Create creates a new product in the database.
func (r *ProductRepositoryPostgres) Create(
	ctx context.Context,
	product *entity.Product,
) (*entity.Product, error) {
	query := `
		INSERT INTO products (id, name, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, price, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		product.ID,
		product.Name,
		product.Price,
		product.CreatedAt,
		product.UpdatedAt,
	)

	var savedProduct entity.Product

	err := row.Scan(
		&savedProduct.ID,
		&savedProduct.Name,
		&savedProduct.Price,
		&savedProduct.CreatedAt,
		&savedProduct.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &savedProduct, nil
}

// Update updates an existing product in the database.
func (r *ProductRepositoryPostgres) Update(
	ctx context.Context,
	product *entity.Product,
) (*entity.Product, error) {
	query := `
		UPDATE products 
		SET name = $2, price = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, name, price, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		product.ID,
		product.Name,
		product.Price,
		product.UpdatedAt,
	)

	var updatedProduct entity.Product

	err := row.Scan(
		&updatedProduct.ID,
		&updatedProduct.Name,
		&updatedProduct.Price,
		&updatedProduct.CreatedAt,
		&updatedProduct.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("product not found")
		}

		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &updatedProduct, nil
}

// Delete deletes a product from the database.
func (r *ProductRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// FindByID finds a product by its ID.
func (r *ProductRepositoryPostgres) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Product, error) {
	query := `
		SELECT id, name, price, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	var product entity.Product

	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	return &product, nil
}

// FindAll finds all products with pagination.
func (r *ProductRepositoryPostgres) FindAll(
	ctx context.Context,
	limit, offset int,
) ([]*entity.Product, error) {
	query := `
		SELECT id, name, price, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find products: %w", err)
	}
	defer rows.Close()

	var products []*entity.Product

	for rows.Next() {
		var product entity.Product

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading products: %w", err)
	}

	return products, nil
}

// Count returns the total number of products.
func (r *ProductRepositoryPostgres) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM products`

	var count int64

	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}

// Exists checks if a product exists by ID.
func (r *ProductRepositoryPostgres) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`

	var exists bool

	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}

	return exists, nil
}
