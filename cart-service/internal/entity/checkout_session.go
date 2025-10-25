// Package entity defines the CheckoutSession entity and its validation logic.
package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
)

// CheckoutSession represents a checkout session in the marketplace.
// CheckoutSession is a lightweight snapshot of selected cart items.
// No cart_id reference - uses snapshot pattern for immutability.
type CheckoutSession struct {
	ID             uuid.UUID
	IdempotencyKey uuid.UUID
	CustomerID     uuid.UUID
	CartID         uuid.UUID
	AddressID      *uuid.UUID
	CarrierID      *string
	Status         constant.CheckoutSessionStatus
	PaymentGateway *string
	Currency       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Items          []CheckoutSessionItem
}

// CheckoutSessionItem represents an item in a checkout session.
// Uses UUIDv7 for id, which provides natural chronological ordering.
type CheckoutSessionItem struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	ProductName string
	Quantity    int64
	UnitPrice   decimal.Decimal
}

// NewCheckoutSession creates a new checkout session with validation.
// Uses snapshot pattern - copies items from cart, no cart reference.
func NewCheckoutSession(
	idempotencyKey uuid.UUID,
	customerID uuid.UUID,
	cartID uuid.UUID,
	currency string,
	items []CheckoutSessionItem,
) (*CheckoutSession, error) {
	checkoutSessionID := uuid.New()
	now := time.Now()

	// Initialize items with UUIDv7 for chronological ordering
	for i := range items {
		items[i].ID = uuid.New() // Using UUIDv7 provides natural ordering
	}

	session := &CheckoutSession{
		ID:             checkoutSessionID,
		IdempotencyKey: idempotencyKey,
		CustomerID:     customerID,
		CartID:         cartID,
		AddressID:      nil,
		CarrierID:      nil,
		Status:         constant.CheckoutSessionStatusPending,
		PaymentGateway: nil,
		Currency:       currency,
		CreatedAt:      now,
		UpdatedAt:      now,
		Items:          items,
	}

	if err := session.validate(); err != nil {
		return nil, err
	}

	return session, nil
}

// NewCheckoutSessionItem creates a new checkout session item with validation.
func NewCheckoutSessionItem(
	productID uuid.UUID,
	productName string,
	quantity int64,
	unitPrice decimal.Decimal,
) (*CheckoutSessionItem, error) {
	if productID == uuid.Nil {
		return nil, errors.New("product_id must not be empty")
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	if unitPrice.LessThan(decimal.Zero) {
		return nil, errors.New("unit_price must not be negative")
	}

	return &CheckoutSessionItem{
		ID:          uuid.New(),
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
	}, nil
}

// validate performs business rule validation.
func (cs *CheckoutSession) validate() error {
	if cs.IdempotencyKey == uuid.Nil {
		return errors.New("idempotency_key must not be empty")
	}

	if cs.CustomerID == uuid.Nil {
		return errors.New("customer_id must not be empty")
	}

	if len(cs.Items) == 0 {
		return errors.New("checkout session must have at least one item")
	}

	if cs.CreatedAt.After(cs.UpdatedAt) {
		return errors.New("created_at must be before updated_at")
	}

	// Validate items
	productSeen := make(map[uuid.UUID]bool)

	for i := range cs.Items {
		item := &cs.Items[i]

		if item.ProductID == uuid.Nil {
			return fmt.Errorf("item[%d]: product_id must not be empty", i)
		}

		if item.Quantity <= 0 {
			return fmt.Errorf("item[%d]: quantity must be greater than 0", i)
		}

		if item.UnitPrice.LessThan(decimal.Zero) {
			return fmt.Errorf("item[%d]: unit_price must not be negative", i)
		}

		// prevent duplicate products
		if productSeen[item.ProductID] {
			return fmt.Errorf("item[%d]: duplicate product_id %s", i, item.ProductID)
		}

		productSeen[item.ProductID] = true
	}

	return nil
}

// UpdateStatus updates the checkout session status with validation.
func (cs *CheckoutSession) UpdateStatus(status constant.CheckoutSessionStatus) error {
	cs.Status = status
	cs.UpdatedAt = time.Now()

	return cs.validate()
}

// CanPlaceOrder checks if the checkout session can place an order.
func (cs *CheckoutSession) CanPlaceOrder() bool {
	return cs.Status == constant.CheckoutSessionStatusPending
}

// CanBeCanceled checks if the checkout session can be canceled.
func (cs *CheckoutSession) CanBeCanceled() bool {
	return cs.Status == constant.CheckoutSessionStatusPending
}
