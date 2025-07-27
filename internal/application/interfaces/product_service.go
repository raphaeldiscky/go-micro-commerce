// Package interfaces defines the interfaces for application services.
package interfaces

//go:generate mockgen -source=product_service.go -destination=../../mocks/mock_product_service.go -package=mocks

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-ddd-template/internal/application/command"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/query"
)

// ProductService defines the interface for product-related operations.
type ProductService interface {
	CreateProduct(
		productCommand *command.CreateProductCommand,
	) (*command.CreateProductCommandResult, error)
	FindAllProducts() (*query.ProductQueryListResult, error)
	FindProductById(id uuid.UUID) (*query.ProductQueryResult, error)
}
