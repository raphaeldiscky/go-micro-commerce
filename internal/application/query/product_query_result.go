package query

import "github.com/raphaeldiscky/go-ddd/internal/application/common"

type ProductQueryResult struct {
	Result *common.ProductResult
}

type ProductQueryListResult struct {
	Result []*common.ProductResult
}
