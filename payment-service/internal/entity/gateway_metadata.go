// Package entity defines payment gateway metadata types.
package entity

import (
	"encoding/json"
	"fmt"
)

// StripeMetadata contains Stripe-specific payment metadata stored in JSONB.
type StripeMetadata struct {
	PaymentMethodID *string `json:"payment_method_id,omitempty"` // pm_xxx - Stripe PaymentMethod ID
	CustomerID      *string `json:"customer_id,omitempty"`       // cus_xxx - Stripe Customer ID
	ClientSecret    *string `json:"client_secret,omitempty"`     // For frontend Stripe.js
	SetupIntentID   *string `json:"setup_intent_id,omitempty"`   // seti_xxx - SetupIntent ID
	PaymentIntentID *string `json:"payment_intent_id,omitempty"` // pi_xxx - PaymentIntent ID (can differ from gateway_transaction_id)
	ChargeID        *string `json:"charge_id,omitempty"`         // ch_xxx - Charge ID
}

// XenditMetadata contains Xendit-specific payment metadata (future use).
type XenditMetadata struct {
	InvoiceID      *string `json:"invoice_id,omitempty"`
	ExternalID     *string `json:"external_id,omitempty"`
	PaymentMethod  *string `json:"payment_method,omitempty"`
	PaymentChannel *string `json:"payment_channel,omitempty"`
	VABankCode     *string `json:"va_bank_code,omitempty"` // For virtual accounts
	AccountNumber  *string `json:"account_number,omitempty"`
	EwalletType    *string `json:"ewallet_type,omitempty"` // OVO, DANA, etc.
}

// MidtransMetadata contains Midtrans-specific payment metadata (future use).
type MidtransMetadata struct {
	OrderID           *string `json:"order_id,omitempty"`
	TransactionID     *string `json:"transaction_id,omitempty"`
	TransactionStatus *string `json:"transaction_status,omitempty"`
	PaymentType       *string `json:"payment_type,omitempty"`
	VANumber          *string `json:"va_number,omitempty"`   // Virtual Account number
	BankType          *string `json:"bank_type,omitempty"`   // Bank type (BCA, BNI, etc.)
	BillerCode        *string `json:"biller_code,omitempty"` // For e-wallet
	BillKey           *string `json:"bill_key,omitempty"`    // For e-wallet
}

// GatewayMetadata is a marker interface for type-safe gateway metadata.
type GatewayMetadata interface {
	ToMap() (map[string]any, error)
}

// NewStripeMetadataFromMap creates StripeMetadata from a map.
func NewStripeMetadataFromMap(data map[string]any) (*StripeMetadata, error) {
	if data == nil {
		return &StripeMetadata{}, nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map: %w", err)
	}

	var metadata StripeMetadata
	if err = json.Unmarshal(jsonData, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to StripeMetadata: %w", err)
	}

	return &metadata, nil
}

// ToMap converts StripeMetadata to map[string]any for database storage.
func (m *StripeMetadata) ToMap() (map[string]any, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal StripeMetadata: %w", err)
	}

	var result map[string]any
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal StripeMetadata to map: %w", err)
	}

	return result, nil
}

// NewMidtransMetadataFromMap creates MidtransMetadata from a map.
func NewMidtransMetadataFromMap(data map[string]any) (*MidtransMetadata, error) {
	if data == nil {
		return &MidtransMetadata{}, nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map: %w", err)
	}

	var metadata MidtransMetadata
	if err = json.Unmarshal(jsonData, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to MidtransMetadata: %w", err)
	}

	return &metadata, nil
}

// ToMap converts MidtransMetadata to map[string]any for database storage.
func (m *MidtransMetadata) ToMap() (map[string]any, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MidtransMetadata: %w", err)
	}

	var result map[string]any
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MidtransMetadata to map: %w", err)
	}

	return result, nil
}

// NewXenditMetadataFromMap creates XenditMetadata from a map.
func NewXenditMetadataFromMap(data map[string]any) (*XenditMetadata, error) {
	if data == nil {
		return &XenditMetadata{}, nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map: %w", err)
	}

	var metadata XenditMetadata
	if err = json.Unmarshal(jsonData, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to XenditMetadata: %w", err)
	}

	return &metadata, nil
}

// ToMap converts XenditMetadata to map[string]any for database storage.
func (m *XenditMetadata) ToMap() (map[string]any, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XenditMetadata: %w", err)
	}

	var result map[string]any
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XenditMetadata to map: %w", err)
	}

	return result, nil
}
