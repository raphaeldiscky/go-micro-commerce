package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Seller struct {
	Id        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Email     string
}

func NewSeller(name string, email string) *Seller {
	return &Seller{
		Id:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Email:     email,
	}
}

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

func (s *Seller) UpdateName(name string) error {
	s.Name = name
	s.UpdatedAt = time.Now()

	return s.validate()
}

func (s *Seller) UpdateEmail(email string) error {
	s.Email = email
	s.UpdatedAt = time.Now()

	return s.validate()
}
