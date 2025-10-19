// Package dto defines data transfer objects for the auth service.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateAddressRequest represents a request to create a new address.
type CreateAddressRequest struct {
	ReceiverName string   `json:"receiver_name" validate:"required,min=1,max=255"`
	AddressLine1 string   `json:"address_line1" validate:"required,min=1,max=255"`
	AddressLine2 *string  `json:"address_line2" validate:"omitempty,max=255"`
	City         string   `json:"city"          validate:"required,min=1,max=100"`
	State        *string  `json:"state"         validate:"omitempty,max=100"`
	PostalCode   string   `json:"postal_code"   validate:"required,min=1,max=20"`
	CountryCode  string   `json:"country_code"  validate:"required,len=2,uppercase"`
	Latitude     *float64 `json:"latitude"      validate:"omitempty,min=-90,max=90"`
	Longitude    *float64 `json:"longitude"     validate:"omitempty,min=-180,max=180"`
	IsDefault    bool     `json:"is_default"`
	Note         *string  `json:"note"          validate:"omitempty"`
}

// UpdateAddressRequest represents a request to update an existing address.
type UpdateAddressRequest struct {
	ReceiverName *string  `json:"receiver_name" validate:"omitempty,min=1,max=255"`
	AddressLine1 *string  `json:"address_line1" validate:"omitempty,min=1,max=255"`
	AddressLine2 *string  `json:"address_line2" validate:"omitempty,max=255"`
	City         *string  `json:"city"          validate:"omitempty,min=1,max=100"`
	State        *string  `json:"state"         validate:"omitempty,max=100"`
	PostalCode   *string  `json:"postal_code"   validate:"omitempty,min=1,max=20"`
	CountryCode  *string  `json:"country_code"  validate:"omitempty,len=2,uppercase"`
	Latitude     *float64 `json:"latitude"      validate:"omitempty,min=-90,max=90"`
	Longitude    *float64 `json:"longitude"     validate:"omitempty,min=-180,max=180"`
	Note         *string  `json:"note"          validate:"omitempty"`
}

// AddressResponse represents an address in API responses.
type AddressResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	ReceiverName string    `json:"receiver_name"`
	AddressLine1 string    `json:"address_line1"`
	AddressLine2 *string   `json:"address_line2,omitempty"`
	City         string    `json:"city"`
	State        *string   `json:"state,omitempty"`
	PostalCode   string    `json:"postal_code"`
	CountryCode  string    `json:"country_code"`
	Latitude     *float64  `json:"latitude,omitempty"`
	Longitude    *float64  `json:"longitude,omitempty"`
	IsDefault    bool      `json:"is_default"`
	Note         *string   `json:"note,omitempty"`
	FullAddress  string    `json:"full_address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SetDefaultAddressRequest represents a request to set an address as default.
// The address ID is extracted from the URL path parameter.
type SetDefaultAddressRequest struct {
	// AddressID is extracted from path parameter, not from request body
}
