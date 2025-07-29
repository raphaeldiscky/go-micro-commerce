package repositories

import (
	"github.com/google/uuid"

	entities "github.com/raphaeldiscky/go-ddd-template/internal/domain/entity"
)

// ProductRepository defines the interface for product repository operations.
type ProductRepository interface {
	Create(product *entities.ValidatedProduct) (*entities.Product, error)
	FindByID(id uuid.UUID) (*entities.Product, error)
	FindAll() ([]*entities.Product, error)
	Update(product *entities.ValidatedProduct) (*entities.Product, error)
	Delete(id uuid.UUID) error
}
