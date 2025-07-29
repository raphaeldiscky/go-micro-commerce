// Package request provides the DTO for creating a product.
package request

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-ddd-template/internal/application/command"
)

// CreateProductRequest represents the request to create a product.
type CreateProductRequest struct {
	Name     string  `json:"Name"`
	Price    float64 `json:"Price"`
	SellerID string  `json:"SellerId"`
}

// ToCreateProductCommand converts the CreateProductRequest to a CreateProductCommand.
func (req *CreateProductRequest) ToCreateProductCommand() (*command.CreateProductCommand, error) {
	sellerID, err := uuid.Parse(req.SellerID)
	if err != nil {
		return nil, err
	}

	return &command.CreateProductCommand{
		Name:     req.Name,
		Price:    req.Price,
		SellerID: sellerID,
	}, nil
}
