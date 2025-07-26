package request

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/command"
)

type CreateProductRequest struct {
	Name     string  `json:"Name"`
	Price    float64 `json:"Price"`
	SellerId string  `json:"SellerId"`
}

func (req *CreateProductRequest) ToCreateProductCommand() (*command.CreateProductCommand, error) {
	sellerId, err := uuid.Parse(req.SellerId)
	if err != nil {
		return nil, err
	}

	return &command.CreateProductCommand{
		Name:     req.Name,
		Price:    req.Price,
		SellerId: sellerId,
	}, nil
}
