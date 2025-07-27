package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/domain/entities"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	// Create saves a new product
	Create(ctx context.Context, product *entities.Product) (*entities.Product, error)

	// FindById retrieves a product by its ID
	FindById(ctx context.Context, id uuid.UUID) (*entities.Product, error)

	// FindAll retrieves all products with optional pagination
	FindAll(ctx context.Context, limit, offset int) ([]*entities.Product, error)

	// FindBySellerId retrieves all products for a specific seller
	FindBySellerId(ctx context.Context, sellerId uuid.UUID, limit, offset int) ([]*entities.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *entities.Product) (*entities.Product, error)

	// Delete removes a product by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if a product exists by ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// Count returns the total number of products
	Count(ctx context.Context) (int64, error)

	// CountBySellerId returns the total number of products for a seller
	CountBySellerId(ctx context.Context, sellerId uuid.UUID) (int64, error)
}
