// Package response contains the response structures for the API.
package response

import "time"

// ProductResponse represents the response structure for a product.
type ProductResponse struct {
	ID        string
	Name      string
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ListProductsResponse represents the response structure for a list of products.
type ListProductsResponse struct {
	Products []*ProductResponse `json:"Products"`
}
