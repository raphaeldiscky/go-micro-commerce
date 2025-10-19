// Package service provides business logic for the auth service.
package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/repository"
)

// AddressService defines the methods for the address service.
type AddressService interface {
	CreateAddress(
		ctx context.Context,
		userID uuid.UUID,
		req *dto.CreateAddressRequest,
	) (*dto.AddressResponse, error)
	GetAddress(
		ctx context.Context,
		userID, addressID uuid.UUID,
	) (*dto.AddressResponse, error)
	ListUserAddresses(
		ctx context.Context,
		userID uuid.UUID,
		limit int64,
		cursor string,
	) ([]*dto.AddressResponse, *pkgdto.CursorPagination, error)
	GetDefaultAddress(
		ctx context.Context,
		userID uuid.UUID,
	) (*dto.AddressResponse, error)
	UpdateAddress(
		ctx context.Context,
		userID, addressID uuid.UUID,
		req *dto.UpdateAddressRequest,
	) (*dto.AddressResponse, error)
	DeleteAddress(
		ctx context.Context,
		userID, addressID uuid.UUID,
	) error
	SetDefaultAddress(
		ctx context.Context,
		userID, addressID uuid.UUID,
	) (*dto.AddressResponse, error)
}

// addressService implements AddressService.
type addressService struct {
	dataStore repository.DataStore
	logger    logger.Logger
}

// NewAddressService creates a new addressService.
func NewAddressService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
) AddressService {
	return &addressService{
		dataStore: dataStore,
		logger:    appLogger,
	}
}

// CreateAddress creates a new address for the user.
func (s *addressService) CreateAddress(
	ctx context.Context,
	userID uuid.UUID,
	req *dto.CreateAddressRequest,
) (*dto.AddressResponse, error) {
	addressRepo := s.dataStore.AddressRepository()

	// Check if user has reached the maximum number of addresses
	count, err := addressRepo.CountByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to count user addresses", "user_id", userID, "error", err)

		return nil, httperror.NewInternalServerError("failed to count user addresses")
	}

	if count >= constant.MaxAddressesPerUser {
		return nil, httperror.NewBadRequestError(constant.MaxAddressesReachedErrorMessage)
	}

	// If this is the first address or explicitly requested, set as default
	isDefault := count == 0 || req.IsDefault

	// Create address entity
	address := mapper.MapCreateRequestToEntity(req, userID, isDefault)

	// Validate coordinates if provided
	if err = address.ValidateCoordinates(); err != nil {
		return nil, httperror.NewBadRequestError(err.Error())
	}

	// Validate country code
	if err = address.ValidateCountryCode(); err != nil {
		return nil, httperror.NewBadRequestError(err.Error())
	}

	// If setting as default, unset all existing defaults first
	if isDefault && count > 0 {
		if err = addressRepo.UnsetAllDefaults(ctx, userID); err != nil {
			s.logger.Error("failed to unset default addresses", "user_id", userID, "error", err)

			return nil, httperror.NewInternalServerError("failed to update default address")
		}
	}

	// Create the address
	if err = addressRepo.Create(ctx, address); err != nil {
		s.logger.Error("failed to create address", "user_id", userID, "error", err)

		return nil, httperror.NewInternalServerError("failed to create address")
	}

	s.logger.Info("address created successfully", "user_id", userID, "address_id", address.ID)

	return mapper.MapToAddressResponse(address), nil
}

// GetAddress retrieves a single address by ID with ownership verification.
func (s *addressService) GetAddress(
	ctx context.Context,
	userID, addressID uuid.UUID,
) (*dto.AddressResponse, error) {
	addressRepo := s.dataStore.AddressRepository()

	address, err := addressRepo.GetByID(ctx, addressID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperror.NewNotFoundError(constant.AddressNotFoundErrorMessage)
		}

		s.logger.Error("failed to get address", "address_id", addressID, "error", err)

		return nil, httperror.NewInternalServerError("failed to get address")
	}

	// Verify ownership
	if address.UserID != userID {
		return nil, httperror.NewForbiddenError(constant.AddressAccessDeniedErrorMessage)
	}

	return mapper.MapToAddressResponse(address), nil
}

// ListUserAddresses retrieves addresses for a user with cursor-based pagination.
func (s *addressService) ListUserAddresses(
	ctx context.Context,
	userID uuid.UUID,
	limit int64,
	cursor string,
) ([]*dto.AddressResponse, *pkgdto.CursorPagination, error) {
	addressRepo := s.dataStore.AddressRepository()

	var (
		cursorID        string
		cursorTimestamp int64
		cursorIsDefault string
	)

	// Decode cursor if provided
	if cursor != "" {
		cursorData, err := pageutils.DecodeCursor(cursor)
		if err != nil {
			return nil, nil, httperror.NewBadRequestError("invalid cursor")
		}

		cursorID = cursorData.ID
		cursorTimestamp = cursorData.Timestamp
		cursorIsDefault = cursorData.Value
	}

	// Fetch limit + 1 items to check if there are more
	fetchLimit := limit + 1

	addresses, err := addressRepo.GetByUserIDWithCursor(
		ctx,
		userID,
		fetchLimit,
		cursorID,
		cursorTimestamp,
		cursorIsDefault,
	)
	if err != nil {
		s.logger.Error("failed to list user addresses", "user_id", userID, "error", err)

		return nil, nil, httperror.NewInternalServerError("failed to list addresses")
	}

	// Check if there are more results
	hasNext := len(addresses) > int(limit)
	if hasNext {
		addresses = addresses[:limit]
	}

	// Map to response DTOs
	responses := mapper.MapToAddressResponseList(addresses)

	// Generate next cursor from the last item
	var nextCursor string

	if hasNext && len(addresses) > 0 {
		lastAddress := addresses[len(addresses)-1]

		// Convert is_default to string for cursor value
		isDefaultStr := "false"
		if lastAddress.IsDefault {
			isDefaultStr = "true"
		}

		nextCursor, err = pageutils.GenerateNextCursor(
			lastAddress.ID.String(),
			lastAddress.CreatedAt.UnixMilli(),
			isDefaultStr,
		)
		if err != nil {
			s.logger.Error("failed to generate cursor", "error", err)

			return nil, nil, httperror.NewInternalServerError("failed to generate cursor")
		}
	}

	// Build pagination response
	pagination := pageutils.NewCursorPagination(nextCursor, "", hasNext, false, limit)

	return responses, pagination, nil
}

// GetDefaultAddress retrieves the default address for a user.
func (s *addressService) GetDefaultAddress(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.AddressResponse, error) {
	addressRepo := s.dataStore.AddressRepository()

	address, err := addressRepo.GetDefaultByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperror.NewNotFoundError(constant.AddressNotFoundErrorMessage)
		}

		s.logger.Error("failed to get default address", "user_id", userID, "error", err)

		return nil, httperror.NewInternalServerError("failed to get default address")
	}

	return mapper.MapToAddressResponse(address), nil
}

// UpdateAddress updates an address with ownership verification.
func (s *addressService) UpdateAddress(
	ctx context.Context,
	userID, addressID uuid.UUID,
	req *dto.UpdateAddressRequest,
) (*dto.AddressResponse, error) {
	addressRepo := s.dataStore.AddressRepository()

	// Get existing address
	existing, err := addressRepo.GetByID(ctx, addressID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperror.NewNotFoundError(constant.AddressNotFoundErrorMessage)
		}

		s.logger.Error("failed to get address for update", "address_id", addressID, "error", err)

		return nil, httperror.NewInternalServerError("failed to get address")
	}

	// Verify ownership
	if existing.UserID != userID {
		return nil, httperror.NewForbiddenError(constant.AddressAccessDeniedErrorMessage)
	}

	// Merge update request into existing entity
	updatedAddress := mapper.MapUpdateRequestToEntity(req, existing)

	// Validate coordinates if provided
	if err = updatedAddress.ValidateCoordinates(); err != nil {
		return nil, httperror.NewBadRequestError(err.Error())
	}

	// Validate country code
	if err = updatedAddress.ValidateCountryCode(); err != nil {
		return nil, httperror.NewBadRequestError(err.Error())
	}

	// Update the address
	result, err := addressRepo.Update(ctx, updatedAddress)
	if err != nil {
		s.logger.Error("failed to update address", "address_id", addressID, "error", err)

		return nil, httperror.NewInternalServerError("failed to update address")
	}

	s.logger.Info("address updated successfully", "user_id", userID, "address_id", addressID)

	return mapper.MapToAddressResponse(result), nil
}

// DeleteAddress deletes an address with business rule checks.
func (s *addressService) DeleteAddress(
	ctx context.Context,
	userID, addressID uuid.UUID,
) error {
	addressRepo := s.dataStore.AddressRepository()

	// Get existing address
	existing, err := addressRepo.GetByID(ctx, addressID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return httperror.NewNotFoundError(constant.AddressNotFoundErrorMessage)
		}

		s.logger.Error("failed to get address for deletion", "address_id", addressID, "error", err)

		return httperror.NewInternalServerError("failed to get address")
	}

	// Verify ownership
	if existing.UserID != userID {
		return httperror.NewForbiddenError(constant.AddressAccessDeniedErrorMessage)
	}

	// Check if this is the only address
	count, err := addressRepo.CountByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to count user addresses", "user_id", userID, "error", err)

		return httperror.NewInternalServerError("failed to count addresses")
	}

	if count == 1 {
		return httperror.NewBadRequestError(constant.CannotDeleteLastAddressErrorMessage)
	}

	// Check if trying to delete default address
	if existing.IsDefault {
		return httperror.NewBadRequestError(constant.CannotDeleteDefaultAddressErrorMessage)
	}

	// Delete the address
	if err = addressRepo.Delete(ctx, addressID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return httperror.NewNotFoundError(constant.AddressNotFoundErrorMessage)
		}

		s.logger.Error("failed to delete address", "address_id", addressID, "error", err)

		return httperror.NewInternalServerError("failed to delete address")
	}

	s.logger.Info("address deleted successfully", "user_id", userID, "address_id", addressID)

	return nil
}

// SetDefaultAddress atomically sets an address as default with ownership verification.
func (s *addressService) SetDefaultAddress(
	ctx context.Context,
	userID, addressID uuid.UUID,
) (*dto.AddressResponse, error) {
	addressRepo := s.dataStore.AddressRepository()

	// Get existing address to verify ownership
	existing, err := addressRepo.GetByID(ctx, addressID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperror.NewNotFoundError(constant.AddressNotFoundErrorMessage)
		}

		s.logger.Error("failed to get address", "address_id", addressID, "error", err)

		return nil, httperror.NewInternalServerError("failed to get address")
	}

	// Verify ownership
	if existing.UserID != userID {
		return nil, httperror.NewForbiddenError(constant.AddressAccessDeniedErrorMessage)
	}

	// If already default, just return it
	if existing.IsDefault {
		return mapper.MapToAddressResponse(existing), nil
	}

	// Set as default (atomically unsets other defaults)
	if err = addressRepo.SetDefault(ctx, userID, addressID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperror.NewNotFoundError(constant.AddressNotFoundErrorMessage)
		}

		s.logger.Error("failed to set default address", "address_id", addressID, "error", err)

		return nil, httperror.NewInternalServerError("failed to set default address")
	}

	// Get updated address
	updatedAddress, err := addressRepo.GetByID(ctx, addressID)
	if err != nil {
		s.logger.Error("failed to get updated address", "address_id", addressID, "error", err)

		return nil, httperror.NewInternalServerError("failed to get updated address")
	}

	s.logger.Info("default address set successfully",
		"user_id", userID,
		"address_id", addressID)

	return mapper.MapToAddressResponse(updatedAddress), nil
}
