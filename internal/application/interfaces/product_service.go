package interfaces

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/command"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/query"
)

type ProductService interface {
	CreateProduct(productCommand *command.CreateProductCommand) (*command.CreateProductCommandResult, error)
	FindAllProducts() (*query.ProductQueryListResult, error)
	FindProductById(id uuid.UUID) (*query.ProductQueryResult, error)
}
