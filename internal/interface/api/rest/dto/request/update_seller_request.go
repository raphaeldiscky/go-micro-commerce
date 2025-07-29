package request

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-ddd-template/internal/application/command"
)

// UpdateSellerRequest represents the request to update a seller.
type UpdateSellerRequest struct {
	ID   uuid.UUID `json:"ID"`
	Name string    `json:"Name"`
}

// ToUpdateSellerCommand converts the UpdateSellerRequest to an UpdateSellerCommand.
func (req *UpdateSellerRequest) ToUpdateSellerCommand() (*command.UpdateSellerCommand, error) {
	return &command.UpdateSellerCommand{
		ID:   req.ID,
		Name: req.Name,
	}, nil
}
