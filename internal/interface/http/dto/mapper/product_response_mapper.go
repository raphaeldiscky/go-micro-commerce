// Package mapper provides functions to convert domain models to API response DTOs.
package mapper

import (
	"github.com/raphaeldiscky/go-ddd-template/internal/app/common"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/http/dto/response"
)

// ToProductResponse converts a ProductResult to a ProductResponse DTO.
func ToProductResponse(product *common.ProductResult) *response.ProductResponse {
	return &response.ProductResponse{
		ID:        product.ID.String(),
		Name:      product.Name,
		Price:     product.Price,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}

// ToProductListResponse converts a slice of ProductResult to a ListProductsResponse DTO.
func ToProductListResponse(products []*common.ProductResult) *response.ListProductsResponse {
	var responseList []*response.ProductResponse
	for _, product := range products {
		responseList = append(responseList, ToProductResponse(product))
	}

	return &response.ListProductsResponse{Products: responseList}
}
