// Package entity defines the Product entity and its validation logic.
package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/constant"
)

// Order represents an order in the marketplace.
type Order struct {
	ID             uuid.UUID
	IdempotencyKey uuid.UUID // generated from client
	CustomerID     uuid.UUID
	Status         constant.OrderStatus
	TotalPrice     decimal.Decimal
	Items          []OrderItem
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// OrderItem represents an item in an order.
type OrderItem struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Quantity  int
	Price     decimal.Decimal
	CreatedAt time.Time
	UpdatedAt time.Time
}

// validate performs business rule validation.
func (o *Order) validate() error {
	if o.CustomerID == uuid.Nil {
		return errors.New("customer_id must not be empty")
	}

	if o.IdempotencyKey == uuid.Nil {
		return errors.New("idempotency_key must not be empty")
	}

	if o.TotalPrice.LessThan(decimal.Zero) {
		return errors.New("total_price must not be negative")
	}

	if len(o.Items) == 0 {
		return errors.New("order must have at least one item")
	}

	if o.CreatedAt.After(o.UpdatedAt) {
		return errors.New("created_at must be before updated_at")
	}

	// Validate each order item
	productSeen := make(map[uuid.UUID]bool)
	totalCalculated := decimal.Zero

	for i, item := range o.Items {
		if item.ProductID == uuid.Nil {
			return fmt.Errorf("item[%d]: product_id must not be empty", i)
		}

		if item.Quantity <= 0 {
			return fmt.Errorf("item[%d]: quantity must be greater than 0", i)
		}

		if item.Price.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("item[%d]: price must be greater than 0", i)
		}

		// prevent duplicate products
		if productSeen[item.ProductID] {
			return fmt.Errorf("item[%d]: duplicate product_id %s", i, item.ProductID)
		}

		productSeen[item.ProductID] = true

		// accumulate total for cross-check
		totalCalculated = totalCalculated.Add(
			item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))),
		)

		if item.CreatedAt.After(item.UpdatedAt) {
			return fmt.Errorf("item[%d]: created_at must be before updated_at", i)
		}
	}

	// cross-check with order total
	if !o.TotalPrice.Equal(totalCalculated) {
		return fmt.Errorf(
			"total_price mismatch: expected %s, got %s",
			totalCalculated,
			o.TotalPrice,
		)
	}

	return nil
}

// NewOrder creates a new order with validation.
func NewOrder(customerID, idempotencyKey uuid.UUID, items []OrderItem) (*Order, error) {
	totalPrice := decimal.Zero
	for _, item := range items {
		totalPrice = totalPrice.Add(item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
	}

	orderID := uuid.New()

	for i := range items {
		items[i].OrderID = orderID
		items[i].CreatedAt = time.Now()
		items[i].UpdatedAt = time.Now()
	}

	order := &Order{
		ID:             orderID,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CustomerID:     customerID,
		Status:         constant.OrderStatusPending,
		TotalPrice:     totalPrice.Round(2),
		Items:          items,
	}

	if err := order.validate(); err != nil {
		return nil, err
	}

	return order, nil
}

// UpdateStatus updates the order status with validation.
func (o *Order) UpdateStatus(status constant.OrderStatus) error {
	o.Status = status
	o.UpdatedAt = time.Now()

	return o.validate()
}

// AddItem adds an item to the order and recalculates total price.
func (o *Order) AddItem(item *OrderItem) error {
	if item == nil {
		return errors.New("item must not be nil")
	}

	o.Items = append(o.Items, *item)

	// Recalculate total price
	totalPrice := decimal.Zero
	for _, orderItem := range o.Items {
		totalPrice = totalPrice.Add(
			orderItem.Price.Mul(decimal.NewFromInt(int64(orderItem.Quantity))),
		)
	}

	o.TotalPrice = totalPrice.Round(2)
	o.UpdatedAt = time.Now()

	return o.validate()
}

// RemoveItem removes an item from the order and recalculates total price.
func (o *Order) RemoveItem(itemID uuid.UUID) error {
	for i, item := range o.Items {
		if item.ID == itemID {
			o.Items = append(o.Items[:i], o.Items[i+1:]...)

			break
		}
	}

	// Recalculate total price
	totalPrice := decimal.Zero
	for _, orderItem := range o.Items {
		totalPrice = totalPrice.Add(
			orderItem.Price.Mul(decimal.NewFromInt(int64(orderItem.Quantity))),
		)
	}

	o.TotalPrice = totalPrice.Round(2)
	o.UpdatedAt = time.Now()

	return o.validate()
}

// CanBeCancelled checks if order can be canceled.
func (o *Order) CanBeCancelled() bool {
	return o.Status != constant.OrderStatusDelivered && o.Status != constant.OrderStatusCanceled
}

// CanBePaid checks if order can be paid.
func (o *Order) CanBePaid() bool {
	return o.Status == constant.OrderStatusPending
}

// UpdateItems updates order items and recalculates total.
func (o *Order) UpdateItems(items []OrderItem) error {
	o.Items = items
	// Recalculate total price logic here
	return o.validate()
}
