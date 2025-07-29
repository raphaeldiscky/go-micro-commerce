// Package entities defines the Seller entity and its methods.
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Seller represents a seller entity.
type Seller struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Email     string
}

// NewSeller creates a new Seller with the provided name and email.
func NewSeller(name, email string) *Seller {
	return &Seller{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Email:     email,
	}
}

// NewValidatedSeller creates a new Seller and validates it.
func (s *Seller) validate() error {
	if s.Name == "" {
		return errors.New("name must not be empty")
	}

	if s.Email == "" {
		return errors.New("email must not be empty")
	}

	if s.CreatedAt.After(s.UpdatedAt) {
		return errors.New("created_at must be before updated_at")
	}

	return nil
}

// UpdateName updates the seller's name and validates it.
func (s *Seller) UpdateName(name string) error {
	s.Name = name
	s.UpdatedAt = time.Now()

	return s.validate()
}

// UpdateEmail updates the seller's email and validates it.
func (s *Seller) UpdateEmail(email string) error {
	s.Email = email
	s.UpdatedAt = time.Now()

	return s.validate()
}
