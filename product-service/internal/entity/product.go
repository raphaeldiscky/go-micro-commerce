// Package entity defines the Product entity and its validation logic.
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Product represents a product in the marketplace.
type Product struct {
	ID                uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Name              string
	Price             decimal.Decimal
	Quantity          int
	Version           int64 // for optimistic locking
	AllocatedQuantity int   // quantity reserved for orders
}

// validate performs business rule validation.
func (p *Product) validate() error {
	if p.Name == "" {
		return errors.New("name must not be empty")
	}

	if p.Price.LessThanOrEqual(decimal.Zero) {
		return errors.New("price must be greater than 0")
	}

	if p.Quantity < 0 {
		return errors.New("quantity must be greater than or equal to 0")
	}

	if p.AllocatedQuantity < 0 {
		return errors.New("allocated quantity must be greater than or equal to 0")
	}

	if p.Quantity < p.AllocatedQuantity {
		return errors.New(
			"available stock cannot be negative (quantity must be >= allocated quantity)",
		)
	}

	if p.CreatedAt.After(p.UpdatedAt) {
		return errors.New("created_at must be before updated_at")
	}

	return nil
}

// NewProduct creates a new product with validation.
func NewProduct(name string, price decimal.Decimal, quantity int) (*Product, error) {
	product := &Product{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Name:              name,
		Price:             price.Round(2), // Ensure precision of 2 decimal places
		Quantity:          quantity,
		Version:           1,
		AllocatedQuantity: 0,
	}

	if err := product.validate(); err != nil {
		return nil, err
	}

	return product, nil
}

// UpdateName updates the product name with validation.
func (p *Product) UpdateName(name string) error {
	p.Name = name
	p.UpdatedAt = time.Now()

	return p.validate()
}

// UpdatePrice updates the product price with validation.
func (p *Product) UpdatePrice(price decimal.Decimal) error {
	p.Price = price.Round(2) // Ensure precision of 2 decimal places
	p.UpdatedAt = time.Now()

	return p.validate()
}

// UpdateQuantity updates the product quantity with validation.
func (p *Product) UpdateQuantity(quantity int) error {
	p.Quantity = quantity
	p.UpdatedAt = time.Now()
	p.Version++ // increment version for optimistic locking

	return p.validate()
}

// ReserveStock reserves stock for an order.
func (p *Product) ReserveStock(quantity int) error {
	if quantity <= 0 {
		return errors.New("reservation quantity must be greater than 0")
	}

	availableStock := p.Quantity - p.AllocatedQuantity
	if availableStock < quantity {
		return errors.New("insufficient available stock for reservation")
	}

	p.AllocatedQuantity += quantity
	p.UpdatedAt = time.Now()
	p.Version++ // increment version for optimistic locking

	return p.validate()
}

// ReleaseStock releases reserved stock (for order cancellation/rollback).
func (p *Product) ReleaseStock(quantity int) error {
	if quantity <= 0 {
		return errors.New("release quantity must be greater than 0")
	}

	if p.AllocatedQuantity < quantity {
		return errors.New("cannot release more stock than allocated")
	}

	p.AllocatedQuantity -= quantity
	p.UpdatedAt = time.Now()
	p.Version++ // increment version for optimistic locking

	return p.validate()
}

// CommitStock commits reserved stock (converts allocated to sold).
func (p *Product) CommitStock(quantity int) error {
	if quantity <= 0 {
		return errors.New("commit quantity must be greater than 0")
	}

	if p.AllocatedQuantity < quantity {
		return errors.New("cannot commit more stock than allocated")
	}

	p.Quantity -= quantity
	p.AllocatedQuantity -= quantity
	p.UpdatedAt = time.Now()
	p.Version++ // increment version for optimistic locking

	return p.validate()
}

// GetAvailableStock returns the available stock (quantity - allocated).
func (p *Product) GetAvailableStock() int {
	return p.Quantity - p.AllocatedQuantity
}
