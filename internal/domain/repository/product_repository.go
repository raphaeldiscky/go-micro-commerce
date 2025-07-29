package repository

import (
	"github.com/google/uuid"

	entity "github.com/raphaeldiscky/go-ddd-template/internal/domain/entity"
)

// ProductRepository defines the interface for product repository operations.
type ProductRepository interface {
	Create(product *entity.ValidatedProduct) (*entity.Product, error)
	FindByID(id uuid.UUID) (*entity.Product, error)
	FindAll() ([]*entity.Product, error)
	Update(product *entity.ValidatedProduct) (*entity.Product, error)
	Delete(id uuid.UUID) error
}
