package response

import "time"

// SellerResponse represents the response structure for a seller.
type SellerResponse struct {
	Id        string
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ListSellersResponse represents the response structure for a list of sellers.
type ListSellersResponse struct {
	Sellers []*SellerResponse
}
