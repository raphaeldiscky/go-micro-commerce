package query

import "github.com/raphaeldiscky/go-ddd-template/internal/app/common"

// SellerQueryResult represents the result of a query for a seller.
type SellerQueryResult struct {
	Result *common.SellerResult
}

// SellerQueryListResult represents the result of a query for a list of sellers.
type SellerQueryListResult struct {
	Result []*common.SellerResult
}
