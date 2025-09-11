// Package dto provides data transfer objects for the search service.
package dto

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductIndexRequest represents a request to index/update a product.
// Matches ProductCreatedPayload and ProductUpdatedPayload from pkg/event.
type ProductIndexRequest struct {
	ProductID uuid.UUID       `json:"product_id" validate:"required"`
	Name      string          `json:"name"       validate:"required,min=1,max=255"`
	Price     decimal.Decimal `json:"price"      validate:"required"`
	Quantity  int64           `json:"quantity"   validate:"min=0"`
}

// Validate validates the product index request.
func (r *ProductIndexRequest) Validate() error {
	if r.ProductID == uuid.Nil {
		return errors.New("product_id is required")
	}

	if r.Name == "" {
		return errors.New("name is required")
	}

	if r.Price.IsNegative() {
		return errors.New("price cannot be negative")
	}

	if r.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	return nil
}

// BulkIndexRequest represents a bulk indexing request.
type BulkIndexRequest struct {
	Products []ProductIndexRequest `json:"products,omitempty"`
	// Orders - removed for now, only handling products
}

// SearchQueryRequest represents a search request with advanced options.
type SearchQueryRequest struct {
	Query        string                 `json:"query"`
	Filters      map[string]interface{} `json:"filters"`
	Sort         []SortFieldRequest     `json:"sort"`
	From         int                    `json:"from"`
	Size         int                    `json:"size"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
	Highlight    *HighlightRequest      `json:"highlight,omitempty"`
}

// SortFieldRequest represents a sort field in the request.
type SortFieldRequest struct {
	Field string `json:"field" validate:"required"`
	Order string `json:"order" validate:"required,oneof=asc desc"`
}

// HighlightRequest represents highlight configuration in the request.
type HighlightRequest struct {
	Fields   []string `json:"fields"              validate:"required"`
	PreTags  []string `json:"pre_tags,omitempty"`
	PostTags []string `json:"post_tags,omitempty"`
}
