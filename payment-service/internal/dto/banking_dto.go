package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// BankAccount represents a bank account.
type BankAccount struct {
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number"`
	AccountType   string `json:"account_type"` // checking, savings
	BankName      string `json:"bank_name"`
	Currency      string `json:"currency"`
}

// BankTransferRequest represents a bank transfer request.
type BankTransferRequest struct {
	TransactionID  uuid.UUID       `json:"transaction_id"`
	FromAccount    BankAccount     `json:"from_account"`
	ToAccount      BankAccount     `json:"to_account"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Description    string          `json:"description,omitempty"`
	Reference      string          `json:"reference,omitempty"`
	IdempotencyKey string          `json:"idempotency_key"`
}

// BankTransferResponse represents the result of a bank transfer.
type BankTransferResponse struct {
	TransactionID     uuid.UUID                   `json:"transaction_id"`
	BankReferenceID   string                      `json:"bank_reference_id"`
	Status            constant.BankTransferStatus `json:"status"`
	Amount            decimal.Decimal             `json:"amount"`
	Currency          string                      `json:"currency"`
	ProcessedAt       time.Time                   `json:"processed_at"`
	EstimatedComplete *time.Time                  `json:"estimated_complete,omitempty"`
	Fees              *decimal.Decimal            `json:"fees,omitempty"`
}

// AccountVerificationResponse represents the result of account verification.
type AccountVerificationResponse struct {
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number"`
	IsValid       bool   `json:"is_valid"`
	AccountName   string `json:"account_name,omitempty"`
	BankName      string `json:"bank_name,omitempty"`
	AccountType   string `json:"account_type,omitempty"`
}
