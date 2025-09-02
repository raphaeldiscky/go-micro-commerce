// Package mapper provides functions for mapping entity.Product to dto.ProductResponse.
package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/entity"
)

// MapToProductResponse converts domain entity to DTO response.
func MapToProductResponse(product *entity.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:                product.ID,
		Name:              product.Name,
		Price:             product.Price,
		Quantity:          product.Quantity,
		Version:           product.Version,
		ReservedQuantity:  product.ReservedQuantity,
		AvailableQuantity: product.GetAvailableStock(),
		CreatedAt:         product.CreatedAt,
		UpdatedAt:         product.UpdatedAt,
	}
}
