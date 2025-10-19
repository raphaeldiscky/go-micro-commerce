// Package repository defines the interface for product data operations.
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/entity"
)

// ProductRepository defines the interface for product data operations.
type ProductRepository interface {
	// Create saves a new product
	Create(ctx context.Context, product *entity.Product) (*entity.Product, error)

	// FindByID retrieves a product by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)

	// FindByIDsForUpdate retrieves products by their IDs
	FindByIDsForUpdate(ctx context.Context, ids []uuid.UUID) ([]*entity.Product, error)

	// FindByIDs retrieves products by their IDs without locking
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Product, error)

	// FindAllWithCursor retrieves all products with cursor-based pagination
	FindAllWithCursor(
		ctx context.Context,
		limit int64,
		cursorID string,
		cursorTimestamp int64,
	) ([]*entity.Product, error)

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

	// ReserveProducts reserves stock for products atomically with optimistic locking
	ReserveProducts(
		ctx context.Context,
		reservations []entity.ProductReservation,
	) ([]*entity.Product, error)

	// ReleaseProducts releases reserved stock for products atomically without version checking
	ReleaseProducts(
		ctx context.Context,
		releases []entity.ProductRestoration,
	) ([]*entity.Product, error)

	// ConfirmProductsDeduction confirms stock deduction for products atomically without version checking
	ConfirmProductsDeduction(
		ctx context.Context,
		confirmations []entity.ProductRestoration,
	) ([]*entity.Product, error)

	// Delete removes a product by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if a product exists by ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// Count returns the total number of products
	Count(ctx context.Context) (int64, error)
}

// productRepository implements the ProductRepository interface for PostgreSQL.
type productRepository struct {
	db DBTX
}

// NewProductRepository creates a new instance of productRepository.
func NewProductRepository(db DBTX) ProductRepository {
	return &productRepository{
		db: db,
	}
}

// Create creates a new product in the database.
func (r *productRepository) Create(
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
func (r *productRepository) Update(
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
			return nil, errors.New(constant.ProductNotFoundErrorMessage)
		}

		return nil, err
	}

	return &updatedProduct, nil
}

const (
	defaultArgsLenMultiplier = 4
	paramOffset1             = 1
	paramOffset2             = 2
	paramOffset3             = 3
)

// BulkUpdateQuantity updates the quantity of multiple products in the database.
func (r *productRepository) BulkUpdateQuantity(
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
	valueArgs := make([]any, 0, len(products)*defaultArgsLenMultiplier)

	i := 1
	for _, product := range products {
		valueStrings = append(
			valueStrings,
			fmt.Sprintf("($%d, $%d, $%d, $%d)", i, i+paramOffset1, i+paramOffset2, i+paramOffset3),
		)
		valueArgs = append(valueArgs, product.ID, product.Name, product.Price, product.Quantity)
		i += 4
	}

	query = fmt.Sprintf(query, strings.Join(valueStrings, ","))

	_, err := r.db.Exec(ctx, query, valueArgs...)

	return err
}

// Delete deletes a product from the database.
func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
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
func (r *productRepository) FindByID(
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
			return nil, errors.New(constant.ProductNotFoundErrorMessage)
		}

		return nil, err
	}

	return &product, nil
}

// FindByIDsForUpdate finds products by their IDs.
func (r *productRepository) FindByIDsForUpdate(
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

		err = rows.Scan(
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// FindByIDs finds products by their IDs without locking.
func (r *productRepository) FindByIDs(
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

		err = rows.Scan(
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// FindAll finds all products with pagination.
func (r *productRepository) FindAll(
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

		err = rows.Scan(
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// FindAllWithCursor finds all products with cursor-based pagination.
func (r *productRepository) FindAllWithCursor(
	ctx context.Context,
	limit int64,
	cursorID string,
	cursorTimestamp int64,
) ([]*entity.Product, error) {
	var (
		query string
		rows  pgx.Rows
		err   error
	)

	if cursorID == "" {
		query = `
			SELECT id, name, price, quantity, version, reserved_quantity, created_at, updated_at
			FROM products
			ORDER BY created_at DESC, id DESC
			LIMIT $1
		`
		rows, err = r.db.Query(ctx, query, limit)
	} else {
		cursorUUID, parseErr := uuid.Parse(cursorID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid cursor ID: %w", parseErr)
		}

		query = `
			SELECT id, name, price, quantity, version, reserved_quantity, created_at, updated_at
			FROM products
			WHERE created_at < to_timestamp($2 / 1000.0)
			   OR (created_at = to_timestamp($2 / 1000.0) AND id < $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $1
		`
		rows, err = r.db.Query(ctx, query, limit, cursorTimestamp, cursorUUID)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	products := make([]*entity.Product, 0, limit)

	for rows.Next() {
		var product entity.Product

		err = rows.Scan(
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// Count returns the total number of products.
func (r *productRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM products`

	var count int64

	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Exists checks if a product exists by ID.
func (r *productRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`

	var exists bool

	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// UpdateWithOptimisticLock updates a product using optimistic locking.
func (r *productRepository) UpdateWithOptimisticLock(
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

// ReserveProducts reserves stock for multiple products atomically.
func (r *productRepository) ReserveProducts(
	ctx context.Context,
	reservations []entity.ProductReservation,
) ([]*entity.Product, error) {
	if len(reservations) == 0 {
		return []*entity.Product{}, nil
	}

	var reservedProducts []*entity.Product

	// Use a single query for all reservations to maintain atomicity
	batch := &pgx.Batch{}

	for _, reservation := range reservations {
		query := `
            UPDATE products 
            SET quantity = quantity - $2,
                reserved_quantity = reserved_quantity + $2, 
                version = version + 1, 
                updated_at = CURRENT_TIMESTAMP
            WHERE id = $1 
              AND version = $3 
              AND quantity >= $2
            RETURNING id, name, price, quantity, version, reserved_quantity, created_at, updated_at
        `

		batch.Queue(query, reservation.ProductID, reservation.Quantity, reservation.ExpectedVersion)
	}

	// Execute all reservations in single batch
	results := r.db.SendBatch(ctx, batch)

	defer func() {
		if err := results.Close(); err != nil {
			return // ignore
		}
	}()

	for i := range batch.Len() {
		row := results.QueryRow()

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
					reservations[i].ProductID,
				)
			}

			return nil, err
		}

		reservedProducts = append(reservedProducts, &product)
	}

	return reservedProducts, nil
}

// ReleaseProducts releases reserved stock atomically.
func (r *productRepository) ReleaseProducts(
	ctx context.Context,
	releases []entity.ProductRestoration,
) ([]*entity.Product, error) {
	var updatedProducts []*entity.Product

	batch := &pgx.Batch{}

	for _, release := range releases {
		query := `
            UPDATE products
            SET quantity = quantity + $2,
                reserved_quantity = reserved_quantity - $2,
                version = version + 1,
                updated_at = CURRENT_TIMESTAMP
            WHERE id = $1
              AND reserved_quantity >= $2
            RETURNING id, name, price, quantity, version, reserved_quantity, created_at, updated_at
        `

		batch.Queue(query, release.ProductID, release.Quantity)
	}

	results := r.db.SendBatch(ctx, batch)

	defer func() {
		if err := results.Close(); err != nil {
			return // ignore
		}
	}()

	for i := range batch.Len() {
		row := results.QueryRow()

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
					"release failed for product %s: insufficient reserved quantity or version conflict",
					releases[i].ProductID,
				)
			}

			return nil, err
		}

		updatedProducts = append(updatedProducts, &product)
	}

	return updatedProducts, nil
}

// ConfirmProductsDeduction confirms stock deduction.
func (r *productRepository) ConfirmProductsDeduction(
	ctx context.Context,
	confirmations []entity.ProductRestoration,
) ([]*entity.Product, error) {
	var updatedProducts []*entity.Product

	batch := &pgx.Batch{}

	for _, confirmation := range confirmations {
		query := `
            UPDATE products
            SET reserved_quantity = reserved_quantity - $2,
                version = version + 1,
                updated_at = CURRENT_TIMESTAMP
            WHERE id = $1
              AND reserved_quantity >= $2
            RETURNING id, name, price, quantity, version, reserved_quantity, created_at, updated_at
        `

		batch.Queue(
			query,
			confirmation.ProductID,
			confirmation.Quantity,
		)
	}

	results := r.db.SendBatch(ctx, batch)

	defer func() {
		if err := results.Close(); err != nil {
			return // ignore
		}
	}()

	for i := range batch.Len() {
		row := results.QueryRow()

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
					"confirmation failed for product %s: insufficient reserved quantity or version conflict",
					confirmations[i].ProductID,
				)
			}

			return nil, err
		}

		updatedProducts = append(updatedProducts, &product)
	}

	return updatedProducts, nil
}
