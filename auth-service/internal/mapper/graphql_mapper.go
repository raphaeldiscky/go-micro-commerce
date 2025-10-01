package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/dto"
)

// MapUserToGraphQL maps a UserResponse to a User GraphQL type.
func MapUserToGraphQL(user *dto.UserResponse) *graph.User {
	return &graph.User{
		ID:            user.ID.String(),
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		IsActive:      user.IsActive,
		EmailVerified: user.IsEmailVerified,
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
