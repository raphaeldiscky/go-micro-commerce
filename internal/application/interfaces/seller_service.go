package interfaces

//go:generate mockgen -source=seller_service.go -destination=../../mocks/mock_seller_service.go -package=mocks

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/command"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/query"
)

type SellerService interface {
	CreateSeller(sellerCommand *command.CreateSellerCommand) (*command.CreateSellerCommandResult, error)
	FindAllSellers() (*query.SellerQueryListResult, error)
	FindSellerById(id uuid.UUID) (*query.SellerQueryResult, error)
	UpdateSeller(updateCommand *command.UpdateSellerCommand) (*command.UpdateSellerCommandResult, error)
	DeleteSeller(id uuid.UUID) error
}
