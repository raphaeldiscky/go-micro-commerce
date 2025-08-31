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
	ID               uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Name             string
	Price            decimal.Decimal
	Quantity         int64
	Version          int64 // for optimistic locking
	ReservedQuantity int64 // quantity reserved for orders
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

	if p.ReservedQuantity < 0 {
		return errors.New("reserved quantity must be greater than or equal to 0")
	}

	if p.Quantity < p.ReservedQuantity {
		return errors.New(
			"available stock cannot be negative (quantity must be >= reserved quantity)",
		)
	}

	if p.CreatedAt.After(p.UpdatedAt) {
		return errors.New("created_at must be before updated_at")
	}

	return nil
}

// NewProduct creates a new product with validation.
func NewProduct(name string, price decimal.Decimal, quantity int64) (*Product, error) {
	product := &Product{
		ID:               uuid.New(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Name:             name,
		Price:            price.Round(2), // Ensure precision of 2 decimal places
		Quantity:         quantity,
		Version:          1,
		ReservedQuantity: 0,
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
func (p *Product) UpdateQuantity(quantity int64) error {
	p.Quantity = quantity
	p.UpdatedAt = time.Now()
	p.Version++ // increment version for optimistic locking

	return p.validate()
}

// ReserveStock reserves stock for an order.
func (p *Product) ReserveStock(quantity int64) error {
	if quantity <= 0 {
		return errors.New("reservation quantity must be greater than 0")
	}

	availableStock := p.Quantity - p.ReservedQuantity
	if availableStock < quantity {
		return errors.New("insufficient available stock for reservation")
	}

	p.ReservedQuantity += quantity
	p.UpdatedAt = time.Now()
	p.Version++ // increment version for optimistic locking

	return p.validate()
}

// ReleaseStock releases reserved stock (for order cancellation/rollback).
func (p *Product) ReleaseStock(quantity int64) error {
	if quantity <= 0 {
		return errors.New("release quantity must be greater than 0")
	}

	if p.ReservedQuantity < quantity {
		return errors.New("cannot release more stock than reserved")
	}

	p.ReservedQuantity -= quantity
	p.UpdatedAt = time.Now()
	p.Version++ // increment version for optimistic locking

	return p.validate()
}

// CommitStock commits reserved stock (converts reserved to sold).
func (p *Product) CommitStock(quantity int64) error {
	if quantity <= 0 {
		return errors.New("commit quantity must be greater than 0")
	}

	if p.ReservedQuantity < quantity {
		return errors.New("cannot commit more stock than reserved")
	}

	p.Quantity -= quantity
	p.ReservedQuantity -= quantity
	p.UpdatedAt = time.Now()
	p.Version++ // increment version for optimistic locking

	return p.validate()
}

// GetAvailableStock returns the available stock (quantity - reserved).
func (p *Product) GetAvailableStock() int64 {
	return p.Quantity - p.ReservedQuantity
}
