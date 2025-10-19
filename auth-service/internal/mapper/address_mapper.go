// Package mapper provides functions for mapping between address entities and DTOs.
package mapper

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/entity"
)

// MapToAddressResponse converts entity.Address to dto.AddressResponse.
func MapToAddressResponse(address *entity.Address) *dto.AddressResponse {
	return &dto.AddressResponse{
		ID:           address.ID,
		UserID:       address.UserID,
		ReceiverName: address.ReceiverName,
		AddressLine1: address.AddressLine1,
		AddressLine2: address.AddressLine2,
		City:         address.City,
		State:        address.State,
		PostalCode:   address.PostalCode,
		CountryCode:  address.CountryCode,
		Latitude:     address.Latitude,
		Longitude:    address.Longitude,
		IsDefault:    address.IsDefault,
		Note:         address.Note,
		FullAddress:  address.GetFullAddress(),
		CreatedAt:    address.CreatedAt,
		UpdatedAt:    address.UpdatedAt,
	}
}

// MapToAddressResponseList converts a slice of entity.Address to a slice of dto.AddressResponse.
func MapToAddressResponseList(addresses []*entity.Address) []*dto.AddressResponse {
	responses := make([]*dto.AddressResponse, len(addresses))
	for i, address := range addresses {
		responses[i] = MapToAddressResponse(address)
	}

	return responses
}

// MapCreateRequestToEntity converts dto.CreateAddressRequest to entity.Address.
func MapCreateRequestToEntity(
	req *dto.CreateAddressRequest,
	userID uuid.UUID,
	isDefault bool,
) *entity.Address {
	now := time.Now()

	return &entity.Address{
		ID:           uuid.New(),
		UserID:       userID,
		ReceiverName: req.ReceiverName,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		CountryCode:  req.CountryCode,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		IsDefault:    isDefault,
		Note:         req.Note,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// MapUpdateRequestToEntity merges dto.UpdateAddressRequest into existing entity.Address.
func MapUpdateRequestToEntity(
	req *dto.UpdateAddressRequest,
	existing *entity.Address,
) *entity.Address {
	// Only update fields that are provided (non-nil)
	if req.ReceiverName != nil {
		existing.ReceiverName = *req.ReceiverName
	}

	if req.AddressLine1 != nil {
		existing.AddressLine1 = *req.AddressLine1
	}

	if req.AddressLine2 != nil {
		existing.AddressLine2 = req.AddressLine2
	}

	if req.City != nil {
		existing.City = *req.City
	}

	if req.State != nil {
		existing.State = req.State
	}

	if req.PostalCode != nil {
		existing.PostalCode = *req.PostalCode
	}

	if req.CountryCode != nil {
		existing.CountryCode = *req.CountryCode
	}

	if req.Latitude != nil {
		existing.Latitude = req.Latitude
	}

	if req.Longitude != nil {
		existing.Longitude = req.Longitude
	}

	if req.Note != nil {
		existing.Note = req.Note
	}

	existing.UpdatedAt = time.Now()

	return existing
}
