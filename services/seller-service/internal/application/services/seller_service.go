package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/application/dto"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/domain/entities"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/domain/events"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/domain/repositories"
)

// SellerServiceInterface defines the interface for seller business operations
type SellerServiceInterface interface {
	CreateSeller(ctx context.Context, req dto.CreateSellerRequest) (*dto.SellerResponse, error)
	GetSeller(ctx context.Context, id uuid.UUID) (*dto.SellerResponse, error)
	GetSellerByEmail(ctx context.Context, email string) (*dto.SellerResponse, error)
	GetSellers(ctx context.Context, req dto.GetSellersRequest) (*dto.SellerListResponse, error)
	UpdateSeller(ctx context.Context, req dto.UpdateSellerRequest) (*dto.SellerResponse, error)
	UpdateSellerStatus(ctx context.Context, req dto.SellerStatusRequest) (*dto.SellerResponse, error)
	DeleteSeller(ctx context.Context, id uuid.UUID) error
}

// SellerService implements the SellerServiceInterface
type SellerService struct {
	sellerRepo     repositories.SellerRepository
	eventPublisher events.EventPublisher
}

// NewSellerService creates a new instance of SellerService
func NewSellerService(sellerRepo repositories.SellerRepository, eventPublisher events.EventPublisher) SellerServiceInterface {
	return &SellerService{
		sellerRepo:     sellerRepo,
		eventPublisher: eventPublisher,
	}
}

// CreateSeller creates a new seller
func (s *SellerService) CreateSeller(ctx context.Context, req dto.CreateSellerRequest) (*dto.SellerResponse, error) {
	// Check if seller with email already exists
	exists, err := s.sellerRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check seller email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("seller with email %s already exists", req.Email)
	}

	// Create domain entity
	seller, err := entities.NewSeller(req.Name, req.Email, req.Phone, req.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create seller entity: %w", err)
	}

	// Save to repository
	savedSeller, err := s.sellerRepo.Create(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("failed to save seller: %w", err)
	}

	// Publish domain event
	if s.eventPublisher != nil {
		event := events.NewSellerCreatedEvent(savedSeller.Id, savedSeller.Name, savedSeller.Email, savedSeller.Phone, savedSeller.Address)
		if err := s.eventPublisher.Publish(event); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish SellerCreated event: %v\n", err)
		}
	}

	return s.mapToResponse(savedSeller), nil
}

// GetSeller retrieves a seller by ID
func (s *SellerService) GetSeller(ctx context.Context, id uuid.UUID) (*dto.SellerResponse, error) {
	seller, err := s.sellerRepo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get seller: %w", err)
	}

	if seller == nil {
		return nil, fmt.Errorf("seller not found")
	}

	return s.mapToResponse(seller), nil
}

// GetSellerByEmail retrieves a seller by email
func (s *SellerService) GetSellerByEmail(ctx context.Context, email string) (*dto.SellerResponse, error) {
	seller, err := s.sellerRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get seller by email: %w", err)
	}

	if seller == nil {
		return nil, fmt.Errorf("seller not found")
	}

	return s.mapToResponse(seller), nil
}

// GetSellers retrieves sellers with pagination and filtering
func (s *SellerService) GetSellers(ctx context.Context, req dto.GetSellersRequest) (*dto.SellerListResponse, error) {
	var sellers []*entities.Seller
	var total int64
	var err error

	if req.ActiveOnly {
		sellers, err = s.sellerRepo.FindActive(ctx, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get active sellers: %w", err)
		}
		total, err = s.sellerRepo.CountActive(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count active sellers: %w", err)
		}
	} else {
		sellers, err = s.sellerRepo.FindAll(ctx, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get sellers: %w", err)
		}
		total, err = s.sellerRepo.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count sellers: %w", err)
		}
	}

	sellerResponses := make([]dto.SellerResponse, len(sellers))
	for i, seller := range sellers {
		sellerResponses[i] = *s.mapToResponse(seller)
	}

	return &dto.SellerListResponse{
		Sellers: sellerResponses,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}, nil
}

// UpdateSeller updates an existing seller
func (s *SellerService) UpdateSeller(ctx context.Context, req dto.UpdateSellerRequest) (*dto.SellerResponse, error) {
	// Check if seller exists
	existingSeller, err := s.sellerRepo.FindById(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to find seller: %w", err)
	}

	if existingSeller == nil {
		return nil, fmt.Errorf("seller not found")
	}

	// Check if email is being changed and if new email already exists
	if existingSeller.Email != req.Email {
		exists, err := s.sellerRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("seller with email %s already exists", req.Email)
		}
	}

	// Update fields
	if err := existingSeller.UpdateName(req.Name); err != nil {
		return nil, fmt.Errorf("failed to update seller name: %w", err)
	}

	if err := existingSeller.UpdateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("failed to update seller email: %w", err)
	}

	if err := existingSeller.UpdatePhone(req.Phone); err != nil {
		return nil, fmt.Errorf("failed to update seller phone: %w", err)
	}

	if err := existingSeller.UpdateAddress(req.Address); err != nil {
		return nil, fmt.Errorf("failed to update seller address: %w", err)
	}

	// Save updated seller
	updatedSeller, err := s.sellerRepo.Update(ctx, existingSeller)
	if err != nil {
		return nil, fmt.Errorf("failed to update seller: %w", err)
	}

	// Publish domain event
	if s.eventPublisher != nil {
		event := events.NewSellerUpdatedEvent(updatedSeller.Id, updatedSeller.Name, updatedSeller.Email, updatedSeller.Phone, updatedSeller.Address)
		if err := s.eventPublisher.Publish(event); err != nil {
			fmt.Printf("Failed to publish SellerUpdated event: %v\n", err)
		}
	}

	return s.mapToResponse(updatedSeller), nil
}

// UpdateSellerStatus updates a seller's active status
func (s *SellerService) UpdateSellerStatus(ctx context.Context, req dto.SellerStatusRequest) (*dto.SellerResponse, error) {
	// Check if seller exists
	existingSeller, err := s.sellerRepo.FindById(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to find seller: %w", err)
	}

	if existingSeller == nil {
		return nil, fmt.Errorf("seller not found")
	}

	// Update status
	if req.IsActive {
		existingSeller.Activate()
	} else {
		existingSeller.Deactivate()
	}

	// Save updated seller
	updatedSeller, err := s.sellerRepo.Update(ctx, existingSeller)
	if err != nil {
		return nil, fmt.Errorf("failed to update seller status: %w", err)
	}

	// Publish domain event
	if s.eventPublisher != nil {
		var event events.DomainEvent
		if req.IsActive {
			event = events.NewSellerActivatedEvent(updatedSeller.Id)
		} else {
			event = events.NewSellerDeactivatedEvent(updatedSeller.Id)
		}

		if err := s.eventPublisher.Publish(event); err != nil {
			fmt.Printf("Failed to publish seller status event: %v\n", err)
		}
	}

	return s.mapToResponse(updatedSeller), nil
}

// DeleteSeller deletes a seller by ID
func (s *SellerService) DeleteSeller(ctx context.Context, id uuid.UUID) error {
	// Check if seller exists
	exists, err := s.sellerRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check seller existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("seller not found")
	}

	// Delete seller
	if err := s.sellerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete seller: %w", err)
	}

	// Publish domain event
	if s.eventPublisher != nil {
		event := events.NewSellerDeletedEvent(id)
		if err := s.eventPublisher.Publish(event); err != nil {
			fmt.Printf("Failed to publish SellerDeleted event: %v\n", err)
		}
	}

	return nil
}

// mapToResponse converts domain entity to DTO response
func (s *SellerService) mapToResponse(seller *entities.Seller) *dto.SellerResponse {
	return &dto.SellerResponse{
		Id:        seller.Id,
		Name:      seller.Name,
		Email:     seller.Email,
		Phone:     seller.Phone,
		Address:   seller.Address,
		IsActive:  seller.IsActive,
		CreatedAt: seller.CreatedAt,
		UpdatedAt: seller.UpdatedAt,
	}
}
