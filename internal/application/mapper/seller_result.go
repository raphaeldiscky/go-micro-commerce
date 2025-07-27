// Package mapper contains functions to convert domain entities to application DTOs.
package mapper

import (
	"github.com/raphaeldiscky/go-ddd-template/internal/application/common"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/entities"
)

// NewSellerResultFromValidatedEntity converts a ValidatedSeller entity to a SellerResult DTO.
func NewSellerResultFromValidatedEntity(seller *entities.ValidatedSeller) *common.SellerResult {
	return NewSellerResultFromEntity(&seller.Seller)
}

// NewSellerResultFromEntity converts a Seller entity to a SellerResult DTO.
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
