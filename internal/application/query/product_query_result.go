// Package query provides the query results for product-related operations.
package query

import "github.com/raphaeldiscky/go-ddd-template/internal/application/common"

// ProductQueryResult represents the result of a query for a product.
type ProductQueryResult struct {
	Result *common.ProductResult
}

// ProductQueryListResult represents the result of a query for a list of products.
type ProductQueryListResult struct {
	Result []*common.ProductResult
}
