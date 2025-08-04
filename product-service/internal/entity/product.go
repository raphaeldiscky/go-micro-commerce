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
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Price     decimal.Decimal
	Quantity  int
}

// validate performs business rule validation.
func (p *Product) validate() error {
	if p.Name == "" {
		return errors.New("name must not be empty")
	}

	if p.Price.LessThanOrEqual(decimal.Zero) {
		return errors.New("price must be greater than 0")
	}

	if p.CreatedAt.After(p.UpdatedAt) {
		return errors.New("created_at must be before updated_at")
	}

	return nil
}

// NewProduct creates a new product with validation.
func NewProduct(name string, price decimal.Decimal) (*Product, error) {
	product := &Product{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Price:     price.Round(2), // Ensure precision of 2 decimal places
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
