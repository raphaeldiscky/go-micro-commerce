// Package entity defines the domain entities for the auth service.
package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
)

// Address represents a user's address in the system.
type Address struct {
	ID           uuid.UUID `json:"id"            db:"id"`
	UserID       uuid.UUID `json:"user_id"       db:"user_id"`
	ReceiverName string    `json:"receiver_name" db:"receiver_name"`
	AddressLine1 string    `json:"address_line1" db:"address_line1"`
	AddressLine2 *string   `json:"address_line2" db:"address_line2"`
	City         string    `json:"city"          db:"city"`
	State        *string   `json:"state"         db:"state"`
	PostalCode   string    `json:"postal_code"   db:"postal_code"`
	CountryCode  string    `json:"country_code"  db:"country_code"`
	Latitude     *float64  `json:"latitude"      db:"latitude"`
	Longitude    *float64  `json:"longitude"     db:"longitude"`
	IsDefault    bool      `json:"is_default"    db:"is_default"`
	Note         *string   `json:"note"          db:"note"`
	CreatedAt    time.Time `json:"created_at"    db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"    db:"updated_at"`
}

// GetFullAddress returns the complete formatted address string.
func (a *Address) GetFullAddress() string {
	var parts []string

	parts = append(parts, a.AddressLine1)

	if a.AddressLine2 != nil && *a.AddressLine2 != "" {
		parts = append(parts, *a.AddressLine2)
	}

	parts = append(parts, a.City)

	if a.State != nil && *a.State != "" {
		parts = append(parts, *a.State)
	}

	parts = append(parts, a.PostalCode)
	parts = append(parts, a.CountryCode)

	return strings.Join(parts, ", ")
}

// IsComplete checks if the address has all required fields.
func (a *Address) IsComplete() bool {
	return a.ReceiverName != "" &&
		a.AddressLine1 != "" &&
		a.City != "" &&
		a.PostalCode != "" &&
		a.CountryCode != "" &&
		a.ValidateCountryCode() == nil
}

// ValidateCoordinates validates that latitude and longitude are within valid ranges.
func (a *Address) ValidateCoordinates() error {
	if a.Latitude != nil {
		if *a.Latitude < constant.MinLatitude || *a.Latitude > constant.MaxLatitude {
			return fmt.Errorf(
				"%s: latitude %.7f",
				constant.InvalidCoordinatesErrorMessage,
				*a.Latitude,
			)
		}
	}

	if a.Longitude != nil {
		if *a.Longitude < constant.MinLongitude || *a.Longitude > constant.MaxLongitude {
			return fmt.Errorf(
				"%s: longitude %.7f",
				constant.InvalidCoordinatesErrorMessage,
				*a.Longitude,
			)
		}
	}

	return nil
}

// ValidateCountryCode validates that the country code is a 2-character ISO 3166-1 alpha-2 code.
func (a *Address) ValidateCountryCode() error {
	if len(a.CountryCode) != constant.CountryCodeLength {
		return fmt.Errorf(
			"%s: got %d characters",
			constant.InvalidCountryCodeErrorMessage,
			len(a.CountryCode),
		)
	}

	return nil
}

// HasCoordinates checks if the address has geographic coordinates.
func (a *Address) HasCoordinates() bool {
	return a.Latitude != nil && a.Longitude != nil
}
