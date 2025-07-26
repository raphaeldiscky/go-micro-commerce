package query

import "github.com/raphaeldiscky/go-ddd/internal/application/common"

type SellerQueryResult struct {
	Result *common.SellerResult
}

type SellerQueryListResult struct {
	Result []*common.SellerResult
}
