package dto

import "github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/entity"

// MapToUserResponse converts entity.User to dto.UserResponse.
func MapToUserResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Roles:           user.Roles,
		IsActive:        user.IsActive,
		IsEmailVerified: user.IsEmailVerified,
		EmailVerifiedAt: user.EmailVerifiedAt,
		LastLoginAt:     user.LastLoginAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}
