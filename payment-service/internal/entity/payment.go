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
	CreatedAt            time.Time
	UpdatedAt            time.Time
	PaymentGateway       constant.PaymentGateway
	GatewayTransactionID *string        // Gateway transaction ID (e.g., Stripe PaymentIntent ID: pi_xxx)
	GatewayMetadata      map[string]any // JSONB field storing gateway-specific metadata
	CompletedAt          *time.Time
	FailedAt             *time.Time
	ExpiresAt            *time.Time // 24-hour payment window expiry
	Currency             string
	Status               constant.PaymentStatus
	Amount               decimal.Decimal
	ID                   uuid.UUID
	OrderID              uuid.UUID
}

// NewPayment creates a new payment with validation.
// Sets 24-hour payment window by default.
func NewPayment(
	orderID uuid.UUID,
	amount decimal.Decimal,
	currency string,
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
// The metadata parameter should contain gateway-specific data as a map.
func (p *Payment) SetGatewayReference(
	gateway constant.PaymentGateway,
	transactionID string,
	metadata map[string]any,
) error {
	p.PaymentGateway = gateway
	p.GatewayTransactionID = &transactionID
	p.GatewayMetadata = metadata
	p.UpdatedAt = time.Now()

	return p.validate()
}

// SetGatewayMetadataTyped sets typed gateway metadata.
// This provides type safety by accepting a GatewayMetadata interface.
func (p *Payment) SetGatewayMetadataTyped(metadata GatewayMetadata) error {
	metadataMap, err := metadata.ToMap()
	if err != nil {
		return err
	}

	p.GatewayMetadata = metadataMap
	p.UpdatedAt = time.Now()

	return p.validate()
}

// GetStripeMetadata retrieves Stripe-specific metadata with type safety.
// Returns nil if metadata doesn't exist or parsing fails.
func (p *Payment) GetStripeMetadata() (*StripeMetadata, error) {
	if p.GatewayMetadata == nil {
		return &StripeMetadata{}, nil
	}

	return NewStripeMetadataFromMap(p.GatewayMetadata)
}

// SetStripeMetadata sets Stripe-specific metadata with type safety.
func (p *Payment) SetStripeMetadata(metadata *StripeMetadata) error {
	return p.SetGatewayMetadataTyped(metadata)
}

// GetMidtransMetadata retrieves Midtrans-specific metadata with type safety.
func (p *Payment) GetMidtransMetadata() (*MidtransMetadata, error) {
	if p.GatewayMetadata == nil {
		return &MidtransMetadata{}, nil
	}

	return NewMidtransMetadataFromMap(p.GatewayMetadata)
}

// SetMidtransMetadata sets Midtrans-specific metadata with type safety.
func (p *Payment) SetMidtransMetadata(metadata *MidtransMetadata) error {
	return p.SetGatewayMetadataTyped(metadata)
}

// GetXenditMetadata retrieves Xendit-specific metadata with type safety.
func (p *Payment) GetXenditMetadata() (*XenditMetadata, error) {
	if p.GatewayMetadata == nil {
		return &XenditMetadata{}, nil
	}

	return NewXenditMetadataFromMap(p.GatewayMetadata)
}

// SetXenditMetadata sets Xendit-specific metadata with type safety.
func (p *Payment) SetXenditMetadata(metadata *XenditMetadata) error {
	return p.SetGatewayMetadataTyped(metadata)
}

// GetMetadataField retrieves a specific field from the gateway metadata.
// Returns nil if metadata doesn't exist or key not found.
// Use typed getters (GetStripeMetadata, etc.) when possible for type safety.
func (p *Payment) GetMetadataField(key string) any {
	if p.GatewayMetadata == nil {
		return nil
	}

	return p.GatewayMetadata[key]
}

// SetMetadataField sets a specific field in the gateway metadata.
// Creates metadata map if it doesn't exist.
// Prefer using typed setters (SetStripeMetadata, etc.) for type safety.
func (p *Payment) SetMetadataField(key string, value any) error {
	if p.GatewayMetadata == nil {
		p.GatewayMetadata = make(map[string]any)
	}

	p.GatewayMetadata[key] = value
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

	return nil
}
