// Package repository defines the interface for product data operations.
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/entity"
)

// ProductRepositoryInterface defines the interface for product data operations.
type ProductRepositoryInterface interface {
	// Create saves a new product
	Create(ctx context.Context, product *entity.Product) (*entity.Product, error)

	// FindByID retrieves a product by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)

	// FindByIDsForUpdate retrieves products by their IDs
	FindByIDsForUpdate(ctx context.Context, ids []uuid.UUID) ([]*entity.Product, error)

	// FindByIDs retrieves products by their IDs without locking
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Product, error)

	// FindAll retrieves all products with optional pagination
	FindAll(ctx context.Context, limit, offset int64) ([]*entity.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *entity.Product) (*entity.Product, error)

	// UpdateWithOptimisticLock updates a product using optimistic locking
	UpdateWithOptimisticLock(
		ctx context.Context,
		product *entity.Product,
		expectedVersion int64,
	) (*entity.Product, error)

	// BulkUpdateQuantity updates the quantity of multiple products in the database.
	BulkUpdateQuantity(ctx context.Context, products []*entity.Product) error

	// ReserveStock reserves stock for products atomically with optimistic locking
	ReserveStock(ctx context.Context, reservations []ProductReservation) ([]*entity.Product, error)

	// Delete removes a product by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if a product exists by ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// Count returns the total number of products
	Count(ctx context.Context) (int64, error)
}

// ProductReservation represents a stock reservation request.
type ProductReservation struct {
	ProductID       uuid.UUID
	Quantity        int64
	ExpectedVersion int64
}

// ProductRepositoryPostgres implements the ProductRepository interface for PostgreSQL.
type ProductRepositoryPostgres struct {
	db DBTX
}

// NewProductRepositoryPostgres creates a new instance of ProductRepositoryPostgres.
func NewProductRepositoryPostgres(db DBTX) ProductRepositoryInterface {
	return &ProductRepositoryPostgres{
		db: db,
	}
}

// Create creates a new product in the database.
func (r *ProductRepositoryPostgres) Create(
	ctx context.Context,
	product *entity.Product,
) (*entity.Product, error) {
	query := `
		INSERT INTO products (id, name, price, quantity, version, reserved_quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, price, quantity, version, reserved_quantity, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		product.ID,
		product.Name,
		product.Price,
		product.Quantity,
		product.Version,
		product.ReservedQuantity,
		product.CreatedAt,
		product.UpdatedAt,
	)

	var savedProduct entity.Product

	err := row.Scan(
		&savedProduct.ID,
		&savedProduct.Name,
		&savedProduct.Price,
		&savedProduct.Quantity,
		&savedProduct.Version,
		&savedProduct.ReservedQuantity,
		&savedProduct.CreatedAt,
		&savedProduct.UpdatedAt,
	)
	if err != nil {
		return nil, err
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
		SET name = $2, price = $3, quantity = $4, version = $5, reserved_quantity = $6, updated_at = $7
		WHERE id = $1
		RETURNING id, name, price, quantity, version, reserved_quantity, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		product.ID,
		product.Name,
		product.Price,
		product.Quantity,
		product.Version,
		product.ReservedQuantity,
		product.UpdatedAt,
	)

	var updatedProduct entity.Product

	err := row.Scan(
		&updatedProduct.ID,
		&updatedProduct.Name,
		&updatedProduct.Price,
		&updatedProduct.Quantity,
		&updatedProduct.Version,
		&updatedProduct.ReservedQuantity,
		&updatedProduct.CreatedAt,
		&updatedProduct.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		return nil, err
	}

	return &updatedProduct, nil
}

// BulkUpdateQuantity updates the quantity of multiple products in the database.
func (r *ProductRepositoryPostgres) BulkUpdateQuantity(
	ctx context.Context,
	products []*entity.Product,
) error {
	if len(products) == 0 {
		return nil
	}

	query := `
		INSERT INTO
			products (id, name, price, quantity)
		VALUES
			%s
		ON CONFLICT (id) DO UPDATE
		SET
			quantity = EXCLUDED.quantity,
			updated_at = CURRENT_TIMESTAMP
	`

	valueStrings := make([]string, 0, len(products))
	valueArgs := make([]any, 0, len(products)*4)

	i := 1
	for _, product := range products {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i, i+1, i+2, i+3))
		valueArgs = append(valueArgs, product.ID, product.Name, product.Price, product.Quantity)
		i += 4
	}

	query = fmt.Sprintf(query, strings.Join(valueStrings, ","))

	_, err := r.db.Exec(ctx, query, valueArgs...)

	return err
}

// Delete deletes a product from the database.
func (r *ProductRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return err
	}

	return nil
}

// FindByID finds a product by its ID.
func (r *ProductRepositoryPostgres) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Product, error) {
	query := `
		SELECT id, name, price, quantity, version, reserved_quantity, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	var product entity.Product

	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Quantity,
		&product.Version,
		&product.ReservedQuantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &product, nil
}

// FindByIDsForUpdate finds products by their IDs.
func (r *ProductRepositoryPostgres) FindByIDsForUpdate(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*entity.Product, error) {
	if len(ids) == 0 {
		return []*entity.Product{}, nil
	}

	// Build the SQL query with the correct number of placeholders
	query := `
		SELECT id, name, price, quantity, version, reserved_quantity, created_at, updated_at
		FROM products
		WHERE id = ANY($1)
		FOR UPDATE
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*entity.Product

	for rows.Next() {
		var product entity.Product

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Quantity,
			&product.Version,
			&product.ReservedQuantity,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// FindByIDs finds products by their IDs without locking.
func (r *ProductRepositoryPostgres) FindByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*entity.Product, error) {
	if len(ids) == 0 {
		return []*entity.Product{}, nil
	}

	// Build the SQL query with the correct number of placeholders
	query := `
		SELECT id, name, price, quantity, version, reserved_quantity, created_at, updated_at
		FROM products
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*entity.Product

	for rows.Next() {
		var product entity.Product

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Quantity,
			&product.Version,
			&product.ReservedQuantity,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// FindAll finds all products with pagination.
func (r *ProductRepositoryPostgres) FindAll(
	ctx context.Context,
	limit, offset int64,
) ([]*entity.Product, error) {
	query := `
		SELECT id, name, price, quantity, version, reserved_quantity, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*entity.Product

	for rows.Next() {
		var product entity.Product

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Quantity,
			&product.Version,
			&product.ReservedQuantity,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// Count returns the total number of products.
func (r *ProductRepositoryPostgres) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM products`

	var count int64

	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Exists checks if a product exists by ID.
func (r *ProductRepositoryPostgres) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`

	var exists bool

	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// UpdateWithOptimisticLock updates a product using optimistic locking.
func (r *ProductRepositoryPostgres) UpdateWithOptimisticLock(
	ctx context.Context,
	product *entity.Product,
	expectedVersion int64,
) (*entity.Product, error) {
	query := `
		UPDATE products 
		SET name = $2, price = $3, quantity = $4, version = $5, reserved_quantity = $6, updated_at = $7
		WHERE id = $1 AND version = $8
		RETURNING id, name, price, quantity, version, reserved_quantity, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		product.ID,
		product.Name,
		product.Price,
		product.Quantity,
		product.Version,
		product.ReservedQuantity,
		product.UpdatedAt,
		expectedVersion,
	)

	var updatedProduct entity.Product

	err := row.Scan(
		&updatedProduct.ID,
		&updatedProduct.Name,
		&updatedProduct.Price,
		&updatedProduct.Quantity,
		&updatedProduct.Version,
		&updatedProduct.ReservedQuantity,
		&updatedProduct.CreatedAt,
		&updatedProduct.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("optimistic lock conflict: product version mismatch")
		}

		return nil, err
	}

	return &updatedProduct, nil
}

// ReserveStock reserves stock for products atomically with optimistic locking.
func (r *ProductRepositoryPostgres) ReserveStock(
	ctx context.Context,
	reservations []ProductReservation,
) ([]*entity.Product, error) {
	if len(reservations) == 0 {
		return []*entity.Product{}, nil
	}

	var reservedProducts []*entity.Product

	// Process each reservation atomically
	for _, reservation := range reservations {
		query := `
			UPDATE products 
			SET reserved_quantity = reserved_quantity + $2, 
				version = version + 1, 
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 
			  AND version = $3 
			  AND (quantity - reserved_quantity) >= $2
			RETURNING id, name, price, quantity, version, reserved_quantity, created_at, updated_at
		`

		row := r.db.QueryRow(ctx, query,
			reservation.ProductID,
			reservation.Quantity,
			reservation.ExpectedVersion,
		)

		var product entity.Product

		err := row.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Quantity,
			&product.Version,
			&product.ReservedQuantity,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf(
					"reservation failed for product %s: insufficient stock or version conflict",
					reservation.ProductID,
				)
			}

			return nil, err
		}

		reservedProducts = append(reservedProducts, &product)
	}

	return reservedProducts, nil
}
