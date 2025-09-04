// Package entity defines the Product entity and its validation logic.
package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// Order represents an order in the marketplace.
type Order struct {
	ID             uuid.UUID
	IdempotencyKey uuid.UUID // generated from client
	CustomerID     uuid.UUID
	Status         constant.OrderStatus
	Currency       string
	TotalPrice     decimal.Decimal
	TotalTax       decimal.Decimal
	TotalDiscount  decimal.Decimal
	Items          []OrderItem
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// OrderPricing represents the pricing details for an order.
type OrderPricing struct {
	TotalPrice    decimal.Decimal
	TotalDiscount decimal.Decimal
	TotalTax      decimal.Decimal
}

// OrderItem represents an item in an order.
type OrderItem struct {
	ID            uuid.UUID
	OrderID       uuid.UUID
	ProductID     uuid.UUID
	Quantity      int64
	Currency      string
	UnitPrice     decimal.Decimal
	TotalTax      decimal.Decimal
	TotalDiscount decimal.Decimal
	TotalPrice    decimal.Decimal
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewOrderItem creates a new order item with validation and proper defaults.
func NewOrderItem(
	productID uuid.UUID,
	quantity int64,
	unitPrice decimal.Decimal,
	currency string,
) (*OrderItem, error) {
	if productID == uuid.Nil {
		return nil, errors.New("product_id must not be empty")
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	if unitPrice.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("unit_price must be greater than 0")
	}

	if currency == "" {
		currency = "IDR" // Default currency
	}

	now := time.Now()
	totalPrice := unitPrice.Mul(decimal.NewFromInt(quantity))

	return &OrderItem{
		ID:            uuid.New(),
		ProductID:     productID,
		Quantity:      quantity,
		Currency:      currency,
		UnitPrice:     unitPrice,
		TotalTax:      decimal.Zero, // Default to zero, can be updated later
		TotalDiscount: decimal.Zero, // Default to zero, can be updated later
		TotalPrice:    totalPrice,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// validate performs business rule validation.
func (o *Order) validate() error {
	if err := o.validateOrderFields(); err != nil {
		return err
	}

	if err := o.validateItems(); err != nil {
		return err
	}

	return o.validateTotals()
}

// validateOrderFields validates basic order fields.
func (o *Order) validateOrderFields() error {
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

	return nil
}

// validateItems validates each order item and checks for duplicates.
func (o *Order) validateItems() error {
	productSeen := make(map[uuid.UUID]bool)

	for i := range o.Items {
		item := &o.Items[i]

		if err := o.validateItem(item, i, o.Currency); err != nil {
			return err
		}

		// prevent duplicate products
		if productSeen[item.ProductID] {
			return fmt.Errorf("item[%d]: duplicate product_id %s", i, item.ProductID)
		}

		productSeen[item.ProductID] = true
	}

	return nil
}

// validateItem validates a single order item.
func (o *Order) validateItem(item *OrderItem, index int, currency string) error {
	if item.ProductID == uuid.Nil {
		return fmt.Errorf("item[%d]: product_id must not be empty", index)
	}

	if item.Currency != currency {
		return fmt.Errorf("item[%d]: currency must be %s", index, currency)
	}

	if item.Quantity <= 0 {
		return fmt.Errorf("item[%d]: quantity must be greater than 0", index)
	}

	if item.UnitPrice.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("item[%d]: price must be greater than 0", index)
	}

	if item.CreatedAt.After(item.UpdatedAt) {
		return fmt.Errorf("item[%d]: created_at must be before updated_at", index)
	}

	return nil
}

// validateTotals validates order totals against item calculations.
func (o *Order) validateTotals() error {
	totalCalculated := decimal.Zero
	discountCalculated := decimal.Zero
	taxCalculated := decimal.Zero

	for i := range o.Items {
		item := &o.Items[i]
		totalCalculated = totalCalculated.Add(
			item.UnitPrice.Mul(decimal.NewFromInt(item.Quantity)),
		)
		discountCalculated = discountCalculated.Add(item.TotalDiscount)
		taxCalculated = taxCalculated.Add(item.TotalTax)
	}

	if !o.TotalPrice.Equal(totalCalculated) {
		return fmt.Errorf(
			"total_price mismatch: expected %s, got %s",
			totalCalculated,
			o.TotalPrice,
		)
	}

	if !o.TotalDiscount.Equal(discountCalculated) {
		return fmt.Errorf(
			"total_discount mismatch: expected %s, got %s",
			discountCalculated,
			o.TotalDiscount,
		)
	}

	if !o.TotalTax.Equal(taxCalculated) {
		return fmt.Errorf(
			"total_tax mismatch: expected %s, got %s",
			taxCalculated,
			o.TotalTax,
		)
	}

	return nil
}

// NewOrder creates a new order with validation.
func NewOrder(customerID, idempotencyKey uuid.UUID, items []OrderItem) (*Order, error) {
	totalPrice := decimal.Zero
	totalDiscount := decimal.Zero
	totalTax := decimal.Zero

	for i := range items {
		totalPrice = totalPrice.Add(items[i].UnitPrice.Mul(decimal.NewFromInt(items[i].Quantity)))
		totalDiscount = totalDiscount.Add(items[i].TotalDiscount)
		totalTax = totalTax.Add(items[i].TotalTax)
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
		Currency:       "IDR", // Default currency
		TotalPrice:     totalPrice.Round(2),
		TotalTax:       totalTax.Round(2),
		TotalDiscount:  totalDiscount.Round(2),
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
	for i := range o.Items {
		totalPrice = totalPrice.Add(
			o.Items[i].UnitPrice.Mul(decimal.NewFromInt(o.Items[i].Quantity)),
		)
	}

	o.TotalPrice = totalPrice.Round(2)
	o.UpdatedAt = time.Now()

	return o.validate()
}

// RemoveItem removes an item from the order and recalculates total price.
func (o *Order) RemoveItem(itemID uuid.UUID) error {
	for i := range o.Items {
		if o.Items[i].ID == itemID {
			o.Items = append(o.Items[:i], o.Items[i+1:]...)

			break
		}
	}

	// Recalculate total price
	totalPrice := decimal.Zero
	for i := range o.Items {
		totalPrice = totalPrice.Add(
			o.Items[i].UnitPrice.Mul(decimal.NewFromInt(o.Items[i].Quantity)),
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
