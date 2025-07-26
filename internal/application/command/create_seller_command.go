package command

import "github.com/raphaeldiscky/go-ddd-template/internal/application/common"

type CreateSellerCommand struct {
	// TODO: Implement idempotency key

	Name string
}

type CreateSellerCommandResult struct {
	Result *common.SellerResult
}
