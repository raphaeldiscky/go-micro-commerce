// Package mapper contains functions to convert domain entity to application DTOs.
package mapper

import (
	"github.com/raphaeldiscky/go-ddd-template/internal/app/common"
	entity "github.com/raphaeldiscky/go-ddd-template/internal/domain/entity"
)

// NewSellerResultFromValidatedEntity converts a ValidatedSeller entity to a SellerResult DTO.
func NewSellerResultFromValidatedEntity(seller *entity.ValidatedSeller) *common.SellerResult {
	return NewSellerResultFromEntity(&seller.Seller)
}

// NewSellerResultFromEntity converts a Seller entity to a SellerResult DTO.
func NewSellerResultFromEntity(seller *entity.Seller) *common.SellerResult {
	if seller == nil {
		return nil
	}

	return &common.SellerResult{
		ID:        seller.ID,
		Name:      seller.Name,
		Email:     seller.Email,
		CreatedAt: seller.CreatedAt,
		UpdatedAt: seller.UpdatedAt,
	}
}
