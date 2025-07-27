package request

import "github.com/raphaeldiscky/go-ddd-template/internal/application/command"

type CreateSellerRequest struct {
	Name  string `json:"Name"`
	Email string `json:"Email"`
}

func (req *CreateSellerRequest) ToCreateSellerCommand() (*command.CreateSellerCommand, error) {
	return &command.CreateSellerCommand{
		Name:  req.Name,
		Email: req.Email,
	}, nil
}
