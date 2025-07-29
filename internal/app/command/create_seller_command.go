// Package command defines the commands for the application layer.
package command

import "github.com/raphaeldiscky/go-ddd-template/internal/app/common"

// CreateSellerCommand represents the command to create a new seller.
type CreateSellerCommand struct {
	// TODO: Implement idempotency key

	Name  string
	Email string
}

// CreateSellerCommandResult represents the result of creating a seller.
type CreateSellerCommandResult struct {
	Result *common.SellerResult
}
