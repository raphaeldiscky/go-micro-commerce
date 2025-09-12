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
	FromAccount    BankAccount     `json:"from_account"`
	ToAccount      BankAccount     `json:"to_account"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Description    string          `json:"description,omitempty"`
	Reference      string          `json:"reference,omitempty"`
	IdempotencyKey string          `json:"idempotency_key"`
	TransactionID  uuid.UUID       `json:"transaction_id"`
}

// BankTransferResponse represents the result of a bank transfer.
type BankTransferResponse struct {
	ProcessedAt       time.Time                   `json:"processed_at"`
	EstimatedComplete *time.Time                  `json:"estimated_complete,omitempty"`
	Fees              *decimal.Decimal            `json:"fees,omitempty"`
	BankReferenceID   string                      `json:"bank_reference_id"`
	Status            constant.BankTransferStatus `json:"status"`
	Amount            decimal.Decimal             `json:"amount"`
	Currency          string                      `json:"currency"`
	TransactionID     uuid.UUID                   `json:"transaction_id"`
}

// AccountVerificationResponse represents the result of account verification.
type AccountVerificationResponse struct {
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number"`
	AccountName   string `json:"account_name,omitempty"`
	BankName      string `json:"bank_name,omitempty"`
	AccountType   string `json:"account_type,omitempty"`
	IsValid       bool   `json:"is_valid"`
}
