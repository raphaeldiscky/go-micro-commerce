// Package client provides external service clients for the payment service.
package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

// BankingClient defines the interface for banking service integration.
type BankingClient interface {
	// TransferFunds transfers money between bank accounts
	TransferFunds(
		ctx context.Context,
		req *dto.BankTransferRequest,
	) (*dto.BankTransferResponse, error)

	// GetTransferStatus retrieves the status of a bank transfer
	GetTransferStatus(
		ctx context.Context,
		transactionID uuid.UUID,
	) (*dto.BankTransferResponse, error)

	// VerifyAccount verifies if a bank account is valid and active
	VerifyAccount(
		ctx context.Context,
		account *dto.BankAccount,
	) (*dto.AccountVerificationResponse, error)

	// GetAccountBalance retrieves the balance of a bank account
	GetAccountBalance(ctx context.Context, account *dto.BankAccount) (decimal.Decimal, error)

	// CancelTransfer cancels a pending bank transfer
	CancelTransfer(ctx context.Context, transactionID uuid.UUID) error
}
