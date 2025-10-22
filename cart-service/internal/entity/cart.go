// Package entity defines the Cart entity and its validation logic.
package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
)

// Cart represents a lightweight cart in the marketplace.
// Carts only persist data and track item selection.
// All pricing calculations are handled by CheckoutSession.
type Cart struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Status     constant.CartStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Items      []CartItem // Not stored in carts table, loaded from cart_items
}

// CartItem represents an item in a cart.
// CartItems only store product reference and quantity.
// Pricing is calculated in CheckoutSession.
type CartItem struct {
	ID                  uuid.UUID
	CartID              uuid.UUID
	ProductID           uuid.UUID
	Quantity            int64
	SelectedForCheckout bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewCart creates a new lightweight cart with validation.
func NewCart(
	customerID uuid.UUID,
	items []CartItem,
) (*Cart, error) {
	if customerID == uuid.Nil {
		return nil, errors.New("customer_id must not be empty")
	}

	cartID := uuid.New()
	now := time.Now()

	// Initialize items with CartID and timestamps
	for i := range items {
		items[i].CartID = cartID
		if items[i].CreatedAt.IsZero() {
			items[i].CreatedAt = now
		}

		if items[i].UpdatedAt.IsZero() {
			items[i].UpdatedAt = now
		}
	}

	cart := &Cart{
		ID:         cartID,
		CustomerID: customerID,
		Status:     constant.CartStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
		Items:      items,
	}

	if err := cart.validate(); err != nil {
		return nil, err
	}

	return cart, nil
}

// NewCartItem creates a new cart item with validation.
func NewCartItem(
	productID uuid.UUID,
	quantity int64,
) (*CartItem, error) {
	if productID == uuid.Nil {
		return nil, errors.New("product_id must not be empty")
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	now := time.Now()

	return &CartItem{
		ID:                  uuid.New(),
		ProductID:           productID,
		Quantity:            quantity,
		SelectedForCheckout: false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// validate performs business rule validation.
func (c *Cart) validate() error {
	if c.CustomerID == uuid.Nil {
		return errors.New("customer_id must not be empty")
	}

	if c.CreatedAt.After(c.UpdatedAt) {
		return errors.New("created_at must be before updated_at")
	}

	// Validate items if present
	if len(c.Items) > 0 {
		return c.validateItems()
	}

	return nil
}

// validateItems validates each cart item and checks for duplicates.
func (c *Cart) validateItems() error {
	productSeen := make(map[uuid.UUID]bool)

	for i := range c.Items {
		item := &c.Items[i]

		if err := c.validateItem(item, i); err != nil {
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

// validateItem validates a single cart item.
func (c *Cart) validateItem(item *CartItem, index int) error {
	if item.ProductID == uuid.Nil {
		return fmt.Errorf("item[%d]: product_id must not be empty", index)
	}

	if item.Quantity <= 0 {
		return fmt.Errorf("item[%d]: quantity must be greater than 0", index)
	}

	if item.CreatedAt.IsZero() {
		return fmt.Errorf("item[%d]: created_at must not be zero", index)
	}

	if item.UpdatedAt.IsZero() {
		return fmt.Errorf("item[%d]: updated_at must not be zero", index)
	}

	return nil
}

// AddItem adds an item to the cart.
func (c *Cart) AddItem(item *CartItem) error {
	if item == nil {
		return errors.New("item must not be nil")
	}

	c.Items = append(c.Items, *item)
	c.UpdatedAt = time.Now()

	return c.validate()
}

// RemoveItem removes an item from the cart.
func (c *Cart) RemoveItem(itemID uuid.UUID) error {
	for i := range c.Items {
		if c.Items[i].ID == itemID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)

			break
		}
	}

	c.UpdatedAt = time.Now()

	return c.validate()
}

// UpdateItemQuantity updates the quantity of an item in the cart.
func (c *Cart) UpdateItemQuantity(itemID uuid.UUID, quantity int64) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	for i := range c.Items {
		if c.Items[i].ID == itemID {
			c.Items[i].Quantity = quantity
			c.UpdatedAt = time.Now()

			return c.validate()
		}
	}

	return errors.New("item not found in cart")
}

// SelectItemForCheckout marks an item as selected for checkout.
func (c *Cart) SelectItemForCheckout(itemID uuid.UUID, selected bool) error {
	for i := range c.Items {
		if c.Items[i].ID == itemID {
			c.Items[i].SelectedForCheckout = selected
			c.UpdatedAt = time.Now()

			return nil
		}
	}

	return errors.New("item not found in cart")
}

// GetSelectedItems returns all items marked for checkout.
func (c *Cart) GetSelectedItems() []CartItem {
	var selectedItems []CartItem

	for i := range c.Items {
		if c.Items[i].SelectedForCheckout {
			selectedItems = append(selectedItems, c.Items[i])
		}
	}

	return selectedItems
}

// HasItems checks if the cart has any items.
func (c *Cart) HasItems() bool {
	return len(c.Items) > 0
}

// HasSelectedItems checks if the cart has any items selected for checkout.
func (c *Cart) HasSelectedItems() bool {
	for i := range c.Items {
		if c.Items[i].SelectedForCheckout {
			return true
		}
	}

	return false
}

// MarkAsCheckedOut transitions the cart to checked_out status.
func (c *Cart) MarkAsCheckedOut() error {
	if c.Status != constant.CartStatusActive {
		return fmt.Errorf("cannot checkout cart with status %s", c.Status)
	}

	c.Status = constant.CartStatusCheckedOut
	c.UpdatedAt = time.Now()

	return nil
}

// RevertToActive reverts the cart from checked_out back to active.
// Used when checkout is canceled or fails.
func (c *Cart) RevertToActive() error {
	if c.Status != constant.CartStatusCheckedOut {
		return fmt.Errorf("cannot revert cart with status %s to active", c.Status)
	}

	c.Status = constant.CartStatusActive
	c.UpdatedAt = time.Now()

	return nil
}

// MarkAsArchived archives the cart after successful order placement.
func (c *Cart) MarkAsArchived() error {
	if c.Status != constant.CartStatusCheckedOut {
		return fmt.Errorf("cannot archive cart with status %s", c.Status)
	}

	c.Status = constant.CartStatusArchived
	c.UpdatedAt = time.Now()

	return nil
}

// CanCheckout validates if the cart can proceed to checkout.
func (c *Cart) CanCheckout() error {
	if c.Status != constant.CartStatusActive {
		return fmt.Errorf("cart must be active to checkout, current status: %s", c.Status)
	}

	if !c.HasSelectedItems() {
		return errors.New("cart must have selected items to checkout")
	}

	return nil
}
