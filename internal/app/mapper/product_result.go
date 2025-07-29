// Package mapper contains functions to convert domain entities to application DTOs.
package mapper

import (
	"github.com/raphaeldiscky/go-ddd-template/internal/app/common"
	entities "github.com/raphaeldiscky/go-ddd-template/internal/domain/entity"
)

// NewProductResultFromValidatedEntity converts a ValidatedProduct entity to a ProductResult DTO.
func NewProductResultFromValidatedEntity(product *entities.ValidatedProduct) *common.ProductResult {
	return NewProductResultFromEntity(&product.Product)
}

// NewProductResultFromEntity converts a Product entity to a ProductResult DTO.
func NewProductResultFromEntity(product *entities.Product) *common.ProductResult {
	if product == nil {
		return nil
	}

	return &common.ProductResult{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		Seller:    NewSellerResultFromEntity(&product.Seller),
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}
