package command

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd/internal/application/common"
)

type UpdateSellerCommand struct {
	// TODO: Implement idempotency key

	Id   uuid.UUID
	Name string
}

type UpdateSellerCommandResult struct {
	Result *common.SellerResult
}
