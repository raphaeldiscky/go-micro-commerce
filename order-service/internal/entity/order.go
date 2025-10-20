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
	Reason         *string
	PaymentGateway constant.PaymentGateway
	PaymentMethod  constant.PaymentMethod
	Currency       string
	ShippingCost   decimal.Decimal // generated from fulfillment-service
	Subtotal       decimal.Decimal // SUM(unit_price * quantity) for all items
	TotalPrice     decimal.Decimal // SUM(unit_price * quantity) + SUM(total_tax) - SUM(total_discount) + shipping_cost
	TotalTax       decimal.Decimal // SUM(total_tax) for all items
	TotalDiscount  decimal.Decimal // SUM(total_discount) for all items
	Items          []OrderItem
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// OrderItem represents an item in an order.
type OrderItem struct {
	ID            uuid.UUID
	OrderID       uuid.UUID
	ProductID     uuid.UUID
	Quantity      int64
	UnitPrice     decimal.Decimal
	TaxRate       decimal.Decimal
	TotalTax      decimal.Decimal
	TotalDiscount decimal.Decimal
	TotalPrice    decimal.Decimal
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewOrder creates a new order with validation.
func NewOrder(
	customerID, idempotencyKey uuid.UUID,
	currency string,
	items []OrderItem,
) (*Order, error) {
	// 1. Calculate core values from items
	subtotal := decimal.Zero
	totalDiscount := decimal.Zero
	totalTax := decimal.Zero

	for i := range items {
		item := &items[i]
		subtotal = subtotal.Add(item.UnitPrice.Mul(decimal.NewFromInt(item.Quantity)))
		totalDiscount = totalDiscount.Add(item.TotalDiscount)
		totalTax = totalTax.Add(item.TotalTax)
	}

	// 2. Set defaults for other costs
	shippingCost := decimal.Zero // Will be updated later in the fulfillment saga

	// 3. Calculate the FINAL total price
	totalPrice := subtotal.
		Sub(totalDiscount).
		Add(totalTax).
		Add(shippingCost).
		Round(constant.DefaultPricingScale)

	orderID := uuid.New()

	// 4. Initialize items with OrderID
	for i := range items {
		items[i].OrderID = orderID
		items[i].CreatedAt = time.Now()
		items[i].UpdatedAt = time.Now()
	}

	// 5. Create the order
	order := &Order{
		ID:             orderID,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CustomerID:     customerID,
		Status:         constant.OrderStatusPending,
		Currency:       currency,
		ShippingCost:   shippingCost,
		Subtotal:       subtotal.Round(constant.DefaultPricingScale),
		TotalPrice:     totalPrice,
		TotalTax:       totalTax.Round(constant.DefaultPricingScale),
		TotalDiscount:  totalDiscount.Round(constant.DefaultPricingScale),
		Items:          items,
	}

	if err := order.validate(); err != nil {
		return nil, err
	}

	return order, nil
}

// NewOrderItem creates a new order item with validation and proper defaults.
func NewOrderItem(
	productID uuid.UUID,
	quantity int64,
	unitPrice decimal.Decimal,
) (*OrderItem, error) {
	if productID == uuid.Nil {
		return nil, errors.New("product_id must not be empty")
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	if unitPrice.LessThan(decimal.Zero) {
		return nil, errors.New("unit_price must not be negative")
	}

	now := time.Now()
	totalPrice := unitPrice.Mul(decimal.NewFromInt(quantity))

	return &OrderItem{
		ID:            uuid.New(),
		ProductID:     productID,
		Quantity:      quantity,
		UnitPrice:     unitPrice,
		TotalTax:      decimal.Zero, // Default to zero, can be updated later
		TotalDiscount: decimal.Zero, // Default to zero, can be updated later
		TaxRate:       decimal.Zero, // Default to zero, can be updated latera
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

		if err := o.validateItem(item, i); err != nil {
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
func (o *Order) validateItem(item *OrderItem, index int) error {
	if item.ProductID == uuid.Nil {
		return fmt.Errorf("item[%d]: product_id must not be empty", index)
	}

	if item.TaxRate.LessThan(decimal.Zero) {
		return fmt.Errorf("item[%d]: tax_rate must not be negative", index)
	}

	if item.Quantity <= 0 {
		return fmt.Errorf("item[%d]: quantity must be greater than 0", index)
	}

	// Allow zero unit price for saga workflows (prices will be set during saga execution)
	// For non-saga workflows, unit price must be greater than zero
	if item.UnitPrice.LessThan(decimal.Zero) {
		return fmt.Errorf("item[%d]: unit_price must not be negative", index)
	}

	if item.CreatedAt.After(item.UpdatedAt) {
		return fmt.Errorf("item[%d]: created_at must be before updated_at", index)
	}

	return nil
}

// validateTotals validates order totals against item calculations.
func (o *Order) validateTotals() error {
	// 1. Calculate the subtotal from items
	subtotalCalculated := decimal.Zero
	discountCalculated := decimal.Zero
	taxCalculated := decimal.Zero

	for i := range o.Items {
		item := &o.Items[i]
		subtotalCalculated = subtotalCalculated.Add(
			item.UnitPrice.Mul(decimal.NewFromInt(item.Quantity)),
		)
		discountCalculated = discountCalculated.Add(item.TotalDiscount)
		taxCalculated = taxCalculated.Add(item.TotalTax)
	}

	// 2. Validate the stored Subtotal matches the calculated items subtotal
	if !o.Subtotal.Equal(subtotalCalculated) {
		return fmt.Errorf(
			"subtotal mismatch: expected %s, got %s",
			subtotalCalculated, o.Subtotal,
		)
	}

	// 3. Validate the stored discounts and taxes match the sum of items
	if !o.TotalDiscount.Equal(discountCalculated) {
		return fmt.Errorf(
			"total_discount mismatch: expected %s, got %s",
			discountCalculated, o.TotalDiscount,
		)
	}

	if !o.TotalTax.Equal(taxCalculated) {
		return fmt.Errorf(
			"total_tax mismatch: expected %s, got %s",
			taxCalculated, o.TotalTax,
		)
	}

	// 4. Validate the FINAL TotalPrice including shipping
	totalPriceCalculated := o.Subtotal.
		Sub(o.TotalDiscount).
		Add(o.TotalTax).
		Add(o.ShippingCost)

	if !o.TotalPrice.Equal(totalPriceCalculated) {
		return fmt.Errorf(
			"total_price mismatch: expected %s, got %s, check if shipping cost is included",
			totalPriceCalculated, o.TotalPrice,
		)
	}

	// 5. Ensure all totals are non-negative
	if o.Subtotal.LessThan(decimal.Zero) {
		return errors.New("subtotal must not be negative")
	}

	if o.TotalTax.LessThan(decimal.Zero) {
		return errors.New("total_tax must not be negative")
	}

	if o.TotalDiscount.LessThan(decimal.Zero) {
		return errors.New("total_discount must not be negative")
	}

	if o.ShippingCost.LessThan(decimal.Zero) {
		return errors.New("shipping_cost must not be negative")
	}

	return nil
}

// UpdateStatus updates the order status with validation.
func (o *Order) UpdateStatus(status constant.OrderStatus) error {
	o.Status = status
	o.UpdatedAt = time.Now()

	return o.validate()
}

// AddItem adds an item to the order and recalculates totals.
func (o *Order) AddItem(item *OrderItem) error {
	if item == nil {
		return errors.New("item must not be nil")
	}

	o.Items = append(o.Items, *item)

	// Recalculate ALL totals, not just the item price
	o.recalculateTotals()
	o.UpdatedAt = time.Now()

	return o.validate()
}

// RemoveItem removes an item from the order and recalculates totals.
func (o *Order) RemoveItem(itemID uuid.UUID) error {
	for i := range o.Items {
		if o.Items[i].ID == itemID {
			o.Items = append(o.Items[:i], o.Items[i+1:]...)

			break
		}
	}

	o.recalculateTotals()
	o.UpdatedAt = time.Now()

	return o.validate()
}

// CanBeCancelled checks if order can be canceled.
func (o *Order) CanBeCancelled() bool {
	return o.Status != constant.OrderStatusDelivered && o.Status != constant.OrderStatusCanceled &&
		o.Status != constant.OrderStatusPaymentExpired
}

// CanBePaid checks if order can be paid.
func (o *Order) CanBePaid() bool {
	return o.Status == constant.OrderStatusPending
}

// IsPaymentConfirmed checks if payment is confirmed.
func (o *Order) IsPaymentConfirmed() bool {
	return o.Status == constant.OrderStatusPaid || o.Status == constant.OrderStatusDelivered ||
		o.Status == constant.OrderStatusShipped
}

// UpdateItems updates order items and recalculates total.
func (o *Order) UpdateItems(items []OrderItem) error {
	o.Items = items
	// Recalculate total price logic here
	return o.validate()
}

// recalculateTotals recalculates subtotal, discounts, taxes, and the final total price.
// This is a new helper method to keep the logic DRY.
func (o *Order) recalculateTotals() {
	subtotal := decimal.Zero
	totalDiscount := decimal.Zero
	totalTax := decimal.Zero

	for i := range o.Items {
		item := &o.Items[i]
		subtotal = subtotal.Add(item.UnitPrice.Mul(decimal.NewFromInt(item.Quantity)))
		totalDiscount = totalDiscount.Add(item.TotalDiscount)
		totalTax = totalTax.Add(item.TotalTax)
	}

	o.Subtotal = subtotal.Round(constant.DefaultPricingScale)
	o.TotalDiscount = totalDiscount.Round(constant.DefaultPricingScale)
	o.TotalTax = totalTax.Round(constant.DefaultPricingScale)

	// Recalculate the final total including shipping
	o.TotalPrice = o.Subtotal.
		Sub(o.TotalDiscount).
		Add(o.TotalTax).
		Add(o.ShippingCost).
		Round(constant.DefaultPricingScale)
}

// UpdateShippingCost updates the shipping cost and recalculates the total price.
// This is crucial for your saga workflow.
func (o *Order) UpdateShippingCost(newShippingCost decimal.Decimal) error {
	if newShippingCost.LessThan(decimal.Zero) {
		return errors.New("shipping cost must not be negative")
	}

	o.ShippingCost = newShippingCost
	o.recalculateTotals() // Recalculate the total with the new shipping cost
	o.UpdatedAt = time.Now()

	return o.validate()
}
