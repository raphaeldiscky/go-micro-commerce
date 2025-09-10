// Package dto provides data transfer objects for the search service.
package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/entity"
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
		return fmt.Errorf("product_id is required")
	}

	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if r.Price.IsNegative() {
		return fmt.Errorf("price cannot be negative")
	}

	if r.Quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}

	return nil
}

// ToEntity converts the DTO to a ProductDocument entity.
func (r *ProductIndexRequest) ToEntity() *entity.ProductDocument {
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

// BulkIndexRequest represents a bulk indexing request.
type BulkIndexRequest struct {
	Products []ProductIndexRequest `json:"products,omitempty"`
	// Orders - removed for now, only handling products
}

// Validate validates the bulk index request.
func (r *BulkIndexRequest) Validate() error {
	if len(r.Products) == 0 {
		return fmt.Errorf("at least one product must be provided")
	}

	for i, product := range r.Products {
		if err := product.Validate(); err != nil {
			return fmt.Errorf("products[%d]: %w", i, err)
		}
	}

	return nil
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

// Validate validates the search query request.
func (r *SearchQueryRequest) Validate() error {
	if r.Size < 0 || r.Size > 100 {
		return fmt.Errorf("size must be between 0 and 100")
	}

	if r.From < 0 {
		return fmt.Errorf("from must be non-negative")
	}

	for i, sort := range r.Sort {
		if sort.Field == "" {
			return fmt.Errorf("sort[%d]: field is required", i)
		}

		if sort.Order != "asc" && sort.Order != "desc" {
			return fmt.Errorf("sort[%d]: order must be 'asc' or 'desc'", i)
		}
	}

	if r.Highlight != nil {
		if len(r.Highlight.Fields) == 0 {
			return fmt.Errorf("highlight: fields are required")
		}
	}

	return nil
}

// ToEntity converts the DTO to a SearchQuery entity.
func (r *SearchQueryRequest) ToEntity() *entity.SearchQuery {
	sortFields := make([]entity.SortField, len(r.Sort))
	for i, sort := range r.Sort {
		sortFields[i] = entity.SortField{
			Field: sort.Field,
			Order: sort.Order,
		}
	}

	var highlightConfig *entity.HighlightConfig
	if r.Highlight != nil {
		highlightConfig = &entity.HighlightConfig{
			Fields:   r.Highlight.Fields,
			PreTags:  r.Highlight.PreTags,
			PostTags: r.Highlight.PostTags,
		}
	}

	return &entity.SearchQuery{
		Query:        r.Query,
		Filters:      r.Filters,
		Sort:         sortFields,
		From:         r.From,
		Size:         r.Size,
		Aggregations: r.Aggregations,
		Highlight:    highlightConfig,
	}
}
