package repositories

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd/internal/domain/entities"
)

type SellerRepository interface {
	Create(seller *entities.ValidatedSeller) (*entities.Seller, error)
	FindById(id uuid.UUID) (*entities.Seller, error)
	FindAll() ([]*entities.Seller, error)
	Update(seller *entities.ValidatedSeller) (*entities.Seller, error)
	Delete(id uuid.UUID) error
}
