// Package repositories defines the interfaces for seller-related database operations.
package repositories

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-ddd-template/internal/domain/entities"
)

// SellerRepository defines the interface for seller-related database operations.
type SellerRepository interface {
	Create(seller *entities.ValidatedSeller) (*entities.Seller, error)
	FindByID(id uuid.UUID) (*entities.Seller, error)
	FindAll() ([]*entities.Seller, error)
	Update(seller *entities.ValidatedSeller) (*entities.Seller, error)
	Delete(id uuid.UUID) error
}
