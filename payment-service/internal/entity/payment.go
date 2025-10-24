// Package entity defines the Payment entity and its validation logic.
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// Payment represents a payment transaction in the marketplace.
type Payment struct {
	CreatedAt          time.Time
	UpdatedAt          time.Time
	PaymentGateway     constant.PaymentGateway
	GatewayReferenceID *string
	GatewayResponse    map[string]any
	CompletedAt        *time.Time
	FailedAt           *time.Time
	ExpiresAt          *time.Time // 24-hour payment window expiry
	Currency           string
	Status             constant.PaymentStatus
	PaymentMethod      constant.PaymentMethod
	Amount             decimal.Decimal
	ID                 uuid.UUID
	OrderID            uuid.UUID
	PaymentMethodID    *string // Stripe PaymentMethod ID (pm_xxx) for off-session charging
	StripeCustomerID   *string // Stripe Customer ID (cus_xxx) for payment method attachment
}

// NewPayment creates a new payment with validation.
// Sets 24-hour payment window by default.
func NewPayment(
	orderID uuid.UUID,
	amount decimal.Decimal,
	currency string,
	paymentMethod constant.PaymentMethod,
	paymentGateway constant.PaymentGateway,
) (*Payment, error) {
	now := time.Now()
	expiresAt := now.Add(constant.PaymentExpiryDuration)
	payment := &Payment{
		ID:             uuid.New(),
		OrderID:        orderID,
		Amount:         amount.Round(constant.DefaultPricingScale),
		Currency:       currency,
		Status:         constant.PaymentStatusPending,
		PaymentMethod:  paymentMethod,
		PaymentGateway: paymentGateway,
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      &expiresAt,
	}

	if err := payment.validate(); err != nil {
		return nil, err
	}

	return payment, nil
}

// UpdateStatus updates the payment status with validation.
func (p *Payment) UpdateStatus(status constant.PaymentStatus) error {
	p.Status = status
	p.UpdatedAt = time.Now()

	// Set completion/failure timestamps
	switch status {
	case constant.PaymentStatusCompleted:
		now := time.Now()
		p.CompletedAt = &now
	case constant.PaymentStatusFailed:
		now := time.Now()
		p.FailedAt = &now
	case constant.PaymentStatusPending,
		constant.PaymentStatusTimeout,
		constant.PaymentStatusProcessing,
		constant.PaymentStatusRefunded:
		// No action needed
	default:
		return errors.New("invalid payment status")
	}

	return p.validate()
}

// SetGatewayReference sets the payment gateway reference information.
func (p *Payment) SetGatewayReference(
	gateway constant.PaymentGateway,
	referenceID string,
	response map[string]any,
) error {
	p.PaymentGateway = gateway
	p.GatewayReferenceID = &referenceID
	p.GatewayResponse = response
	p.UpdatedAt = time.Now()

	return p.validate()
}

// SetPaymentMethodInfo sets the Stripe payment method and customer IDs.
// Used for storing payment method during SetupIntent flow for later off-session charging.
func (p *Payment) SetPaymentMethodInfo(paymentMethodID, stripeCustomerID string) error {
	p.PaymentMethodID = &paymentMethodID
	p.StripeCustomerID = &stripeCustomerID
	p.UpdatedAt = time.Now()

	return p.validate()
}

// CanBeProcessed checks if payment can be processed.
func (p *Payment) CanBeProcessed() bool {
	return p.Status == constant.PaymentStatusPending
}

// CanBeRefunded checks if payment can be refunded.
func (p *Payment) CanBeRefunded() bool {
	return p.Status == constant.PaymentStatusCompleted
}

// IsCompleted checks if payment is completed.
func (p *Payment) IsCompleted() bool {
	return p.Status == constant.PaymentStatusCompleted
}

// IsFailed checks if payment has failed.
func (p *Payment) IsFailed() bool {
	return p.Status == constant.PaymentStatusFailed
}

// IsExpired checks if payment has exceeded the 24-hour window.
func (p *Payment) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false
	}

	return time.Now().After(*p.ExpiresAt)
}

// CanBeTimedOut checks if payment can be timed out.
// Only pending payments that have expired can be timed out.
func (p *Payment) CanBeTimedOut() bool {
	return p.Status == constant.PaymentStatusPending && p.IsExpired()
}

// validate performs business rule validation.
func (p *Payment) validate() error {
	if p.OrderID == uuid.Nil {
		return errors.New("order_id must not be empty")
	}

	if p.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be greater than zero")
	}

	if p.Currency == "" {
		return errors.New("currency must not be empty")
	}

	if len(p.Currency) != constant.CurrencyLength {
		return errors.New("currency must be a 3-character ISO code")
	}

	if p.CreatedAt.After(p.UpdatedAt) {
		return errors.New("created_at must be before or equal to updated_at")
	}

	// Status validation is handled by database constraints
	// PaymentMethod validation is handled by database constraints

	return nil
}
