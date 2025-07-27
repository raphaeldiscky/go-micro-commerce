package mapper

import (
	"github.com/raphaeldiscky/go-ddd-template/internal/application/common"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/entities"
)

func NewSellerResultFromValidatedEntity(seller *entities.ValidatedSeller) *common.SellerResult {
	return NewSellerResultFromEntity(&seller.Seller)
}

func NewSellerResultFromEntity(seller *entities.Seller) *common.SellerResult {
	if seller == nil {
		return nil
	}

	return &common.SellerResult{
		Id:        seller.Id,
		Name:      seller.Name,
		Email:     seller.Email,
		CreatedAt: seller.CreatedAt,
		UpdatedAt: seller.UpdatedAt,
	}
}
