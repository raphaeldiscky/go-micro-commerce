// Package mapper provides functions for mapping between DTOs and entities.
package mapper

import (
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/entity"
)

// ToEntity converts the DTO to a ProductDocument entity.
func ToEntity(r *dto.ProductIndexRequest) *entity.ProductDocument {
	return &entity.ProductDocument{
		ID:               r.ProductID,
		Name:             r.Name,
		Price:            r.Price,
		Quantity:         r.Quantity,
		ReservedQuantity: 0, // Not provided in event payload
		Version:          0, // Not provided in event payload
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}
