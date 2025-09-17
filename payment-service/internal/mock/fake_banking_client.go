// Package mock provides mock implementations of external service clients.
package mock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/random"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

const (
	fakeBankingDelay              = time.Millisecond * 200 // Simulate network delay
	fakeBankingTransferFailAmount = 10000000               // Fail transfers > 10M IDR
	fakeBankingEstimateComplete   = time.Hour * 2
	fakeBankingSuccessAmount      = 1000000 // Success transfers > 1M IDR
	fakeBankingProcessedAt        = 2 * time.Hour
	fakeBankingBalance            = 100000000
	fakeBankingBaseFee            = 6500
	fakeBankingPercentageFee      = 0.001
	fakeBankingMaxFee             = 25000
	fakeBankingRefLength          = 12
	fakeBankingMinRoutingLength   = 3
)

// fakeBankingClient provides a mock implementation of BankingClient for testing.
type fakeBankingClient struct {
	shouldFail bool
	delay      time.Duration
}

// NewFakeBankingClient creates a new instance of fakeBankingClient.
func NewFakeBankingClient() client.BankingClient {
	return &fakeBankingClient{
		shouldFail: false,
		delay:      fakeBankingDelay,
	}
}

// SetShouldFail configures the client to simulate failures.
func (c *fakeBankingClient) SetShouldFail(shouldFail bool) {
	c.shouldFail = shouldFail
}

// SetDelay configures the simulated network delay.
func (c *fakeBankingClient) SetDelay(delay time.Duration) {
	c.delay = delay
}

// TransferFunds transfers money between bank accounts.
func (c *fakeBankingClient) TransferFunds(
	_ context.Context,
	req *dto.BankTransferRequest,
) (*dto.BankTransferResponse, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("simulated banking API error")
	}

	// Simulate validation
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return &dto.BankTransferResponse{
			TransactionID:   req.TransactionID,
			BankReferenceID: "",
			Status:          constant.BankTransferStatusFailed,
			Amount:          req.Amount,
			Currency:        req.Currency,
			ProcessedAt:     time.Now(),
		}, nil
	}

	bankReferenceID := c.generateBankReferenceID()
	status := constant.BankTransferStatusCompleted

	// Simulate failures for large amounts
	if req.Amount.GreaterThan(decimal.NewFromInt(fakeBankingTransferFailAmount)) { // > 10M IDR
		status = constant.BankTransferStatusFailed
	}

	estimatedComplete := time.Now().Add(fakeBankingEstimateComplete)
	if status == constant.BankTransferStatusCompleted {
		estimatedComplete = time.Now()
	}

	fees := c.calculateTransferFees(req.Amount)

	return &dto.BankTransferResponse{
		TransactionID:     req.TransactionID,
		BankReferenceID:   bankReferenceID,
		Status:            status,
		Amount:            req.Amount,
		Currency:          req.Currency,
		ProcessedAt:       time.Now(),
		EstimatedComplete: &estimatedComplete,
		Fees:              &fees,
	}, nil
}

// GetTransferStatus retrieves the status of a bank transfer.
func (c *fakeBankingClient) GetTransferStatus(
	_ context.Context,
	transactionID uuid.UUID,
) (*dto.BankTransferResponse, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("failed to get transfer status: banking API error")
	}

	// Simulate different statuses based on transaction ID
	statuses := []constant.BankTransferStatus{
		constant.BankTransferStatusPending,
		constant.BankTransferStatusProcessing,
		constant.BankTransferStatusCompleted,
		constant.BankTransferStatusFailed,
	}

	// Use transaction ID to determine status consistently
	statusIndex := len(transactionID.String()) % len(statuses)
	status := statuses[statusIndex]

	amount := decimal.NewFromFloat(
		float64(random.Int(fakeBankingSuccessAmount)),
	) // Random amount up to 1M IDR
	fees := c.calculateTransferFees(amount)
	estimatedComplete := time.Now().Add(fakeBankingEstimateComplete)

	return &dto.BankTransferResponse{
		TransactionID:     transactionID,
		BankReferenceID:   c.generateBankReferenceID(),
		Status:            status,
		Amount:            amount,
		Currency:          "IDR",
		ProcessedAt:       time.Now().Add(fakeBankingProcessedAt),
		EstimatedComplete: &estimatedComplete,
		Fees:              &fees,
	}, nil
}

// VerifyAccount verifies if a bank account is valid and active.
func (c *fakeBankingClient) VerifyAccount(
	_ context.Context,
	account *dto.BankAccount,
) (*dto.AccountVerificationResponse, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("failed to verify account: banking API error")
	}

	// Simulate validation logic
	isValid := c.isValidAccountNumber(account.AccountNumber) &&
		c.isValidRoutingNumber(account.RoutingNumber)

	result := &dto.AccountVerificationResponse{
		AccountNumber: account.AccountNumber,
		RoutingNumber: account.RoutingNumber,
		IsValid:       isValid,
	}

	if isValid {
		result.AccountName = c.generateAccountName()
		result.BankName = c.getBankNameFromRouting(account.RoutingNumber)
		result.AccountType = account.AccountType
	}

	return result, nil
}

// GetAccountBalance retrieves the balance of a bank account.
func (c *fakeBankingClient) GetAccountBalance(
	_ context.Context,
	account *dto.BankAccount,
) (decimal.Decimal, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return decimal.Zero, errors.New("failed to get account balance: banking API error")
	}

	// Simulate validation
	if !c.isValidAccountNumber(account.AccountNumber) {
		return decimal.Zero, errors.New("invalid account number")
	}

	// Generate mock balance based on account number
	balanceFloat := float64(random.Int(fakeBankingBalance)) // Up to 100M IDR

	return decimal.NewFromFloat(balanceFloat), nil
}

// CancelTransfer cancels a pending bank transfer.
func (c *fakeBankingClient) CancelTransfer(
	_ context.Context,
	transactionID uuid.UUID,
) error {
	time.Sleep(c.delay)

	if c.shouldFail {
		return fmt.Errorf(
			"failed to cancel transfer: TransactionID: %s",
			transactionID,
		)
	}

	// Simulate successful cancellation
	return nil
}

// generateBankReferenceID creates a mock bank reference ID.
func (c *fakeBankingClient) generateBankReferenceID() string {
	bankCodes := []string{"BCA", "BNI", "BRI", "MANDIRI", "CIMB"}
	bankCode := bankCodes[random.Int(int64(len(bankCodes)))]
	refNumber := random.NumericString(fakeBankingRefLength)

	return fmt.Sprintf("%s%s", bankCode, refNumber)
}

// calculateTransferFees calculates transfer fees based on amount.
func (c *fakeBankingClient) calculateTransferFees(amount decimal.Decimal) decimal.Decimal {
	// Indonesian banking fees simulation
	baseFee := decimal.NewFromInt(fakeBankingBaseFee)               // Base fee 6,500 IDR
	percentageFee := decimal.NewFromFloat(fakeBankingPercentageFee) // 0.1% of amount

	calculatedFee := amount.Mul(percentageFee)
	if calculatedFee.LessThan(baseFee) {
		return baseFee
	}

	maxFee := decimal.NewFromInt(fakeBankingMaxFee) // Max fee 25,000 IDR
	if calculatedFee.GreaterThan(maxFee) {
		return maxFee
	}

	return calculatedFee
}

// isValidAccountNumber validates Indonesian bank account numbers.
func (c *fakeBankingClient) isValidAccountNumber(accountNumber string) bool {
	// Basic validation: 10-16 digits
	if len(accountNumber) < 10 || len(accountNumber) > 16 {
		return false
	}

	// Check if all characters are digits
	for _, char := range accountNumber {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// isValidRoutingNumber validates Indonesian bank routing numbers.
func (c *fakeBankingClient) isValidRoutingNumber(routingNumber string) bool {
	// Indonesian bank codes are typically 3-4 digits
	if len(routingNumber) < 3 || len(routingNumber) > 4 {
		return false
	}

	// Check if all characters are digits
	for _, char := range routingNumber {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// generateAccountName generates a mock account holder name.
func (c *fakeBankingClient) generateAccountName() string {
	firstNames := []string{
		"Ahmad", "Budi", "Citra", "Dewi", "Eko", "Fitri",
		"Gina", "Handi", "Indira", "Joko", "Kartika", "Lina",
	}
	lastNames := []string{
		"Santoso", "Wijaya", "Sari", "Pratama", "Kusuma", "Putri",
		"Nugroho", "Wati", "Hartono", "Anggraini", "Setiawan", "Rahayu",
	}

	firstName := firstNames[random.Int(int64(len(firstNames)))]
	lastName := lastNames[random.Int(int64(len(lastNames)))]

	return fmt.Sprintf("%s %s", firstName, lastName)
}

// getBankNameFromRouting returns bank name based on routing number.
func (c *fakeBankingClient) getBankNameFromRouting(routingNumber string) string {
	bankMappings := map[string]string{
		"014": "Bank Central Asia (BCA)",
		"009": "Bank Negara Indonesia (BNI)",
		"002": "Bank Rakyat Indonesia (BRI)",
		"008": "Bank Mandiri",
		"022": "CIMB Niaga",
		"213": "Bank Tabungan Negara (BTN)",
		"013": "Bank Permata",
		"200": "Bank Tabungan Pensiunan Nasional (BTPN)",
	}

	// Extract first 3 digits for mapping
	if len(routingNumber) >= fakeBankingMinRoutingLength {
		bankCode := routingNumber[:3]
		if bankName, exists := bankMappings[bankCode]; exists {
			return bankName
		}
	}

	// Default fallback
	bankNames := []string{
		"Bank Central Asia (BCA)", "Bank Negara Indonesia (BNI)",
		"Bank Rakyat Indonesia (BRI)", "Bank Mandiri",
	}

	return bankNames[random.Int(int64(len(bankNames)))]
}
