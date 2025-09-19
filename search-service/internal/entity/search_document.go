package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// SearchDocument represents the base interface for all searchable documents.
type SearchDocument interface {
	GetID() string
	GetType() string
	GetIndexName() string
}

// ProductDocument represents a product in Elasticsearch.
// This matches the actual products table schema.
type ProductDocument struct {
	ID               uuid.UUID       `json:"id"`
	Name             string          `json:"name"`
	Price            decimal.Decimal `json:"price"`
	Quantity         int64           `json:"quantity"`
	ReservedQuantity int64           `json:"reserved_quantity"`
	Version          int64           `json:"version"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	Suggest          SuggestField    `json:"suggest"`
}

// GetID returns the document ID.
func (p *ProductDocument) GetID() string {
	return p.ID.String()
}

// GetType returns the document type.
func (p *ProductDocument) GetType() string {
	return "product"
}

// GetIndexName returns the index name for products.
func (p *ProductDocument) GetIndexName() string {
	return "products"
}

// SuggestField represents the suggest field for autocomplete functionality.
type SuggestField struct {
	Input []string `json:"input"`
}

// SearchResult represents a generic search result.
type SearchResult struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"`
	Score     float64             `json:"score"`
	Source    map[string]any      `json:"source"`
	Highlight map[string][]string `json:"highlight,omitempty"`
}

// SearchResponse represents a search response with pagination.
type SearchResponse struct {
	Results      []SearchResult `json:"results"`
	Total        int64          `json:"total"`
	Page         int            `json:"page"`
	PerPage      int            `json:"per_page"`
	TotalPages   int            `json:"total_pages"`
	Took         int            `json:"took_ms"`
	Aggregations map[string]any `json:"aggregations,omitempty"`
}

// SearchQuery represents a search query with filters and pagination.
type SearchQuery struct {
	Query        string           `json:"query"`
	Filters      map[string]any   `json:"filters"`
	Sort         []SortField      `json:"sort"`
	From         int              `json:"from"`
	Size         int              `json:"size"`
	Aggregations map[string]any   `json:"aggregations,omitempty"`
	Highlight    *HighlightConfig `json:"highlight,omitempty"`
}

// SortField represents a sort configuration.
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

// HighlightConfig represents highlight configuration.
type HighlightConfig struct {
	Fields   []string `json:"fields"`
	PreTags  []string `json:"pre_tags,omitempty"`
	PostTags []string `json:"post_tags,omitempty"`
}

// SuggestionResult represents an autocomplete suggestion.
type SuggestionResult struct {
	Text  string  `json:"text"`
	Score float64 `json:"score"`
}
