// Package repository defines the interface for product data operations.
package repository

import (
	"context"

	"github.com/google/uuid"

	entity "github.com/raphaeldiscky/go-micro-template/product-service/internal/entity"
)

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
