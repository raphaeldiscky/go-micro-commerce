package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Seller represents a seller in the marketplace
type Seller struct {
	Id        uuid.UUID
	Name      string
	Email     string
	Phone     string
	Address   string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewSeller creates a new seller with validation
func NewSeller(name, email, phone, address string) (*Seller, error) {
	if err := validateSellerData(name, email, phone, address); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Seller{
		Id:        uuid.New(),
		Name:      strings.TrimSpace(name),
		Email:     strings.TrimSpace(strings.ToLower(email)),
		Phone:     strings.TrimSpace(phone),
		Address:   strings.TrimSpace(address),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateName updates the seller's name
func (s *Seller) UpdateName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return errors.New("seller name must be at least 2 characters long")
	}
	if len(name) > 100 {
		return errors.New("seller name must not exceed 100 characters")
	}

	s.Name = name
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateEmail updates the seller's email
func (s *Seller) UpdateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if !isValidEmail(email) {
		return errors.New("invalid email format")
	}

	s.Email = email
	s.UpdatedAt = time.Now()
	return nil
}

// UpdatePhone updates the seller's phone
func (s *Seller) UpdatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if len(phone) < 10 {
		return errors.New("phone number must be at least 10 characters long")
	}
	if len(phone) > 20 {
		return errors.New("phone number must not exceed 20 characters")
	}

	s.Phone = phone
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateAddress updates the seller's address
func (s *Seller) UpdateAddress(address string) error {
	address = strings.TrimSpace(address)
	if len(address) < 10 {
		return errors.New("address must be at least 10 characters long")
	}
	if len(address) > 255 {
		return errors.New("address must not exceed 255 characters")
	}

	s.Address = address
	s.UpdatedAt = time.Now()
	return nil
}

// Activate activates the seller
func (s *Seller) Activate() {
	s.IsActive = true
	s.UpdatedAt = time.Now()
}

// Deactivate deactivates the seller
func (s *Seller) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}

// IsValidForRegistration checks if the seller has all required information
func (s *Seller) IsValidForRegistration() bool {
	return len(s.Name) >= 2 &&
		isValidEmail(s.Email) &&
		len(s.Phone) >= 10 &&
		len(s.Address) >= 10
}

// validateSellerData validates seller creation data
func validateSellerData(name, email, phone, address string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return errors.New("seller name must be at least 2 characters long")
	}
	if len(name) > 100 {
		return errors.New("seller name must not exceed 100 characters")
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if !isValidEmail(email) {
		return errors.New("invalid email format")
	}

	phone = strings.TrimSpace(phone)
	if len(phone) < 10 {
		return errors.New("phone number must be at least 10 characters long")
	}
	if len(phone) > 20 {
		return errors.New("phone number must not exceed 20 characters")
	}

	address = strings.TrimSpace(address)
	if len(address) < 10 {
		return errors.New("address must be at least 10 characters long")
	}
	if len(address) > 255 {
		return errors.New("address must not exceed 255 characters")
	}

	return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	if len(email) < 5 || len(email) > 254 {
		return false
	}

	// Basic email format check
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local := parts[0]
	domain := parts[1]

	if len(local) < 1 || len(local) > 64 {
		return false
	}

	if len(domain) < 3 || len(domain) > 253 {
		return false
	}

	// Domain must contain at least one dot
	return strings.Contains(domain, ".")
}
