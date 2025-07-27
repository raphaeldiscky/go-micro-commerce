// Package command defines the UpdateSellerCommand and its result.
package command

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-ddd-template/internal/application/common"
)

// UpdateSellerCommand represents the command to update an existing seller.
type UpdateSellerCommand struct {
	// TODO: Implement idempotency key

	Id   uuid.UUID
	Name string
}

// UpdateSellerCommandResult represents the result of an UpdateSeller command.
type UpdateSellerCommandResult struct {
	Result *common.SellerResult
}
