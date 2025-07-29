// Package mapper provides functions to convert domain entities to API response DTOs.
package mapper

import (
	"github.com/raphaeldiscky/go-ddd-template/internal/application/common"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/api/rest/dto/response"
)

// ToSellerResponse converts a SellerResult to a SellerResponse.
func ToSellerResponse(product *common.SellerResult) *response.SellerResponse {
	return &response.SellerResponse{
		ID:        product.ID.String(),
		Name:      product.Name,
		Email:     product.Email,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}

// ToSellerListResponse converts a list of SellerResult to a ListSellersResponse.
func ToSellerListResponse(products []*common.SellerResult) *response.ListSellersResponse {
	var responseList []*response.SellerResponse

	for _, product := range products {
		responseList = append(responseList, ToSellerResponse(product))
	}

	return &response.ListSellersResponse{Sellers: responseList}
}
