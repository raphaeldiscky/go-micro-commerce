package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/dto"
)

// MapUserToGraphQL maps a UserResponse to a User GraphQL type.
func MapUserToGraphQL(user *dto.UserResponse) *graph.User {
	return &graph.User{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		IsActive:      user.IsActive,
		EmailVerified: user.IsEmailVerified,
		Roles:         user.Roles,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

// MapAuthResponseToGraphQL maps an AuthResponse to an AuthPayload GraphQL type.
func MapAuthResponseToGraphQL(auth *dto.AuthResponse) *graph.AuthPayload {
	return &graph.AuthPayload{
		Token:        auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		User:         MapUserToGraphQL(auth.User),
	}
}

// MapAddressDTOToGraphQL maps an AddressResponse DTO to a GraphQL Address type.
func MapAddressDTOToGraphQL(addressDTO *dto.AddressResponse) *graph.Address {
	return &graph.Address{
		ID:           addressDTO.ID,
		UserID:       addressDTO.UserID,
		ReceiverName: addressDTO.ReceiverName,
		AddressLine1: addressDTO.AddressLine1,
		AddressLine2: addressDTO.AddressLine2,
		City:         addressDTO.City,
		State:        addressDTO.State,
		PostalCode:   addressDTO.PostalCode,
		CountryCode:  addressDTO.CountryCode,
		Latitude:     addressDTO.Latitude,
		Longitude:    addressDTO.Longitude,
		IsDefault:    addressDTO.IsDefault,
		Note:         addressDTO.Note,
		FullAddress:  addressDTO.FullAddress,
		CreatedAt:    addressDTO.CreatedAt,
		UpdatedAt:    addressDTO.UpdatedAt,
	}
}
