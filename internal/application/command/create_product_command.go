// Package command defines the CreateProductCommand and its result.
package command

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-ddd-template/internal/application/common"
)

// CreateProductCommand represents the command to create a new product.
type CreateProductCommand struct {
	// TODO: Implement idempotency key

	Id       uuid.UUID
	Name     string
	Price    float64
	SellerId uuid.UUID
}

// CreateProductCommandResult represents the result of a CreateProduct command.
type CreateProductCommandResult struct {
	Result *common.ProductResult
}
