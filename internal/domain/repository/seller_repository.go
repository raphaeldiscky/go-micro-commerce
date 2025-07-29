// Package repository defines the interfaces for seller-related database operations.
package repository

import (
	"github.com/google/uuid"

	entity "github.com/raphaeldiscky/go-ddd-template/internal/domain/entity"
)

// SellerRepository defines the interface for seller-related database operations.
type SellerRepository interface {
	Create(seller *entity.ValidatedSeller) (*entity.Seller, error)
	FindByID(id uuid.UUID) (*entity.Seller, error)
	FindAll() ([]*entity.Seller, error)
	Update(seller *entity.ValidatedSeller) (*entity.Seller, error)
	Delete(id uuid.UUID) error
}
