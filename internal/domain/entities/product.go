// Package entities defines the core domain entities and their behaviors.
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Product represents a product that has been validated.
type Product struct {
	Id        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Price     float64
	Seller    Seller
}

// validate checks if the Product is valid.
func (p *Product) validate() error {
	if p.Name == "" {
		return errors.New("name must not be empty")
	}

	if p.Price <= 0 {
		return errors.New("price must be greater than 0")
	}

	if p.CreatedAt.After(p.UpdatedAt) {
		return errors.New("created_at must be before updated_at")
	}

	return nil
}

// NewProduct creates a new Product with the provided name, price, and seller.
func NewProduct(name string, price float64, seller ValidatedSeller) *Product {
	return &Product{
		Id:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Price:     price,
		Seller:    seller.Seller,
	}
}

// UpdateName updates the product's name and validates it.
func (p *Product) UpdateName(name string) error {
	p.Name = name
	p.UpdatedAt = time.Now()

	return p.validate()
}

// UpdatePrice updates the product's price and validates it.
func (p *Product) UpdatePrice(price float64) error {
	p.Price = price
	p.UpdatedAt = time.Now()

	return p.validate()
}
