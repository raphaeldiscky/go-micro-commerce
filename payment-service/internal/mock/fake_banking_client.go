// Package mock provides mock implementations of external service clients.
package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/random"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

// FakeBankingClient provides a mock implementation of BankingClientInterface for testing.
type FakeBankingClient struct {
	shouldFail bool
	delay      time.Duration
}

// NewFakeBankingClient creates a new instance of FakeBankingClient.
func NewFakeBankingClient() *FakeBankingClient {
	return &FakeBankingClient{
		shouldFail: false,
		delay:      time.Millisecond * 200, // Simulate network delay
	}
}

// SetShouldFail configures the client to simulate failures.
func (c *FakeBankingClient) SetShouldFail(shouldFail bool) {
	c.shouldFail = shouldFail
}

// SetDelay configures the simulated network delay.
func (c *FakeBankingClient) SetDelay(delay time.Duration) {
	c.delay = delay
}

// TransferFunds transfers money between bank accounts.
func (c *FakeBankingClient) TransferFunds(
	_ context.Context,
	req *dto.BankTransferRequest,
) (*dto.BankTransferResponse, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, fmt.Errorf("simulated banking API error")
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
	if req.Amount.GreaterThan(decimal.NewFromInt(10000000)) { // > 10M IDR
		status = constant.BankTransferStatusFailed
	}

	estimatedComplete := time.Now().Add(2 * time.Hour)
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
func (c *FakeBankingClient) GetTransferStatus(
	_ context.Context,
	transactionID uuid.UUID,
) (*dto.BankTransferResponse, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, fmt.Errorf("failed to get transfer status: banking API error")
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

	amount := decimal.NewFromFloat(float64(random.Int(1000000))) // Random amount up to 1M IDR
	fees := c.calculateTransferFees(amount)
	estimatedComplete := time.Now().Add(time.Duration(random.Int(24)) * time.Hour)

	return &dto.BankTransferResponse{
		TransactionID:     transactionID,
		BankReferenceID:   c.generateBankReferenceID(),
		Status:            status,
		Amount:            amount,
		Currency:          "IDR",
		ProcessedAt:       time.Now().Add(-time.Duration(random.Int(48)) * time.Hour),
		EstimatedComplete: &estimatedComplete,
		Fees:              &fees,
	}, nil
}

// VerifyAccount verifies if a bank account is valid and active.
func (c *FakeBankingClient) VerifyAccount(
	_ context.Context,
	account *dto.BankAccount,
) (*dto.AccountVerificationResponse, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, fmt.Errorf("failed to verify account: banking API error")
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
func (c *FakeBankingClient) GetAccountBalance(
	_ context.Context,
	account *dto.BankAccount,
) (decimal.Decimal, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return decimal.Zero, fmt.Errorf("failed to get account balance: banking API error")
	}

	// Simulate validation
	if !c.isValidAccountNumber(account.AccountNumber) {
		return decimal.Zero, fmt.Errorf("invalid account number")
	}

	// Generate mock balance based on account number
	balanceFloat := float64(random.Int(100000000)) // Up to 100M IDR

	return decimal.NewFromFloat(balanceFloat), nil
}

// CancelTransfer cancels a pending bank transfer.
func (c *FakeBankingClient) CancelTransfer(
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
func (c *FakeBankingClient) generateBankReferenceID() string {
	bankCodes := []string{"BCA", "BNI", "BRI", "MANDIRI", "CIMB"}
	bankCode := bankCodes[random.Int(int64(len(bankCodes)))]
	refNumber := random.NumericString(12)

	return fmt.Sprintf("%s%s", bankCode, refNumber)
}

// calculateTransferFees calculates transfer fees based on amount.
func (c *FakeBankingClient) calculateTransferFees(amount decimal.Decimal) decimal.Decimal {
	// Indonesian banking fees simulation
	baseFee := decimal.NewFromInt(6500)          // Base fee 6,500 IDR
	percentageFee := decimal.NewFromFloat(0.001) // 0.1% of amount

	calculatedFee := amount.Mul(percentageFee)
	if calculatedFee.LessThan(baseFee) {
		return baseFee
	}

	maxFee := decimal.NewFromInt(25000) // Max fee 25,000 IDR
	if calculatedFee.GreaterThan(maxFee) {
		return maxFee
	}

	return calculatedFee
}

// isValidAccountNumber validates Indonesian bank account numbers.
func (c *FakeBankingClient) isValidAccountNumber(accountNumber string) bool {
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
func (c *FakeBankingClient) isValidRoutingNumber(routingNumber string) bool {
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
func (c *FakeBankingClient) generateAccountName() string {
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
func (c *FakeBankingClient) getBankNameFromRouting(routingNumber string) string {
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
	if len(routingNumber) >= 3 {
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
