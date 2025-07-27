package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/domain/entities"
	"github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/domain/repositories"
)

// ProductRepositoryPostgres implements the ProductRepository interface using PostgreSQL
type ProductRepositoryPostgres struct {
	db *pgxpool.Pool
}

// NewProductRepositoryPostgres creates a new instance of ProductRepositoryPostgres
func NewProductRepositoryPostgres(db *pgxpool.Pool) repositories.ProductRepository {
	return &ProductRepositoryPostgres{
		db: db,
	}
}

// Create creates a new product in the database
func (r *ProductRepositoryPostgres) Create(ctx context.Context, product *entities.Product) (*entities.Product, error) {
	query := `
		INSERT INTO products (id, name, price, seller_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, price, seller_id, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		product.Id,
		product.Name,
		product.Price,
		product.SellerId,
		product.CreatedAt,
		product.UpdatedAt,
	)

	var savedProduct entities.Product
	err := row.Scan(
		&savedProduct.Id,
		&savedProduct.Name,
		&savedProduct.Price,
		&savedProduct.SellerId,
		&savedProduct.CreatedAt,
		&savedProduct.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &savedProduct, nil
}

// Update updates an existing product in the database
func (r *ProductRepositoryPostgres) Update(ctx context.Context, product *entities.Product) (*entities.Product, error) {
	query := `
		UPDATE products 
		SET name = $2, price = $3, seller_id = $4, updated_at = $5
		WHERE id = $1
		RETURNING id, name, price, seller_id, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		product.Id,
		product.Name,
		product.Price,
		product.SellerId,
		product.UpdatedAt,
	)

	var updatedProduct entities.Product
	err := row.Scan(
		&updatedProduct.Id,
		&updatedProduct.Name,
		&updatedProduct.Price,
		&updatedProduct.SellerId,
		&updatedProduct.CreatedAt,
		&updatedProduct.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &updatedProduct, nil
}

// Delete deletes a product from the database
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

// FindById finds a product by its ID
func (r *ProductRepositoryPostgres) FindById(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	query := `
		SELECT id, name, price, seller_id, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	var product entities.Product
	err := row.Scan(
		&product.Id,
		&product.Name,
		&product.Price,
		&product.SellerId,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	return &product, nil
}

// FindAll finds all products with pagination
func (r *ProductRepositoryPostgres) FindAll(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	query := `
		SELECT id, name, price, seller_id, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find products: %w", err)
	}
	defer rows.Close()

	var products []*entities.Product
	for rows.Next() {
		var product entities.Product
		err := rows.Scan(
			&product.Id,
			&product.Name,
			&product.Price,
			&product.SellerId,
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

// FindBySellerId finds products by seller ID with pagination
func (r *ProductRepositoryPostgres) FindBySellerId(ctx context.Context, sellerId uuid.UUID, limit, offset int) ([]*entities.Product, error) {
	query := `
		SELECT id, name, price, seller_id, created_at, updated_at
		FROM products
		WHERE seller_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, sellerId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find products by seller: %w", err)
	}
	defer rows.Close()

	var products []*entities.Product
	for rows.Next() {
		var product entities.Product
		err := rows.Scan(
			&product.Id,
			&product.Name,
			&product.Price,
			&product.SellerId,
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

// Count returns the total number of products
func (r *ProductRepositoryPostgres) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM products`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}

// CountBySellerId returns the total number of products for a specific seller
func (r *ProductRepositoryPostgres) CountBySellerId(ctx context.Context, sellerId uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM products WHERE seller_id = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, sellerId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count products by seller: %w", err)
	}

	return count, nil
}

// Exists checks if a product exists by ID
func (r *ProductRepositoryPostgres) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}

	return exists, nil
}
