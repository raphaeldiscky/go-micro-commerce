package request

import "github.com/raphaeldiscky/go-ddd-template/internal/application/command"

// CreateSellerRequest represents the request to create a new seller.
type CreateSellerRequest struct {
	Name  string `json:"Name"`
	Email string `json:"Email"`
}

// ToCreateSellerCommand converts the CreateSellerRequest to a CreateSellerCommand.
func (req *CreateSellerRequest) ToCreateSellerCommand() (*command.CreateSellerCommand, error) {
	return &command.CreateSellerCommand{
		Name:  req.Name,
		Email: req.Email,
	}, nil
}
