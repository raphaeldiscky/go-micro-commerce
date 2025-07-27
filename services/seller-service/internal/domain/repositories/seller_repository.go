package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/domain/entities"
)

// SellerRepository defines the interface for seller data operations
type SellerRepository interface {
	// Create creates a new seller
	Create(ctx context.Context, seller *entities.Seller) (*entities.Seller, error)

	// FindById finds a seller by ID
	FindById(ctx context.Context, id uuid.UUID) (*entities.Seller, error)

	// FindByEmail finds a seller by email
	FindByEmail(ctx context.Context, email string) (*entities.Seller, error)

	// FindAll finds all sellers with pagination
	FindAll(ctx context.Context, limit, offset int) ([]*entities.Seller, error)

	// FindActive finds all active sellers with pagination
	FindActive(ctx context.Context, limit, offset int) ([]*entities.Seller, error)

	// Update updates an existing seller
	Update(ctx context.Context, seller *entities.Seller) (*entities.Seller, error)

	// Delete deletes a seller by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if a seller exists by ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// ExistsByEmail checks if a seller exists by email
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Count returns the total number of sellers
	Count(ctx context.Context) (int64, error)

	// CountActive returns the total number of active sellers
	CountActive(ctx context.Context) (int64, error)
}
