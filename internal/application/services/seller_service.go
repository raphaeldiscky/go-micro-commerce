package services

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd/internal/application/command"
	"github.com/raphaeldiscky/go-ddd/internal/application/interfaces"
	"github.com/raphaeldiscky/go-ddd/internal/application/mapper"
	"github.com/raphaeldiscky/go-ddd/internal/application/query"
	"github.com/raphaeldiscky/go-ddd/internal/domain/entities"
	"github.com/raphaeldiscky/go-ddd/internal/domain/events"
	"github.com/raphaeldiscky/go-ddd/internal/domain/repositories"
)

type SellerService struct {
	repo           repositories.SellerRepository
	eventPublisher events.EventPublisher
}

// NewSellerService - Constructor for the service
func NewSellerService(repo repositories.SellerRepository, eventPublisher events.EventPublisher) interfaces.SellerService {
	return &SellerService{
		repo:           repo,
		eventPublisher: eventPublisher,
	}
}

// CreateSeller saves a new seller
func (s *SellerService) CreateSeller(sellerCommand *command.CreateSellerCommand) (*command.CreateSellerCommandResult, error) {
	var newSeller = entities.NewSeller(sellerCommand.Name)

	validatedSeller, err := entities.NewValidatedSeller(newSeller)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.Create(validatedSeller)
	if err != nil {
		return nil, err
	}

	// Publish SellerCreated event
	if s.eventPublisher != nil {
		sellerCreatedEvent := events.NewSellerCreatedEvent(
			validatedSeller.Id,
			validatedSeller.Name,
		)

		ctx := context.Background()
		if publishErr := s.eventPublisher.Publish(ctx, sellerCreatedEvent); publishErr != nil {
			log.Printf("Failed to publish SellerCreated event: %v", publishErr)
			// Note: In production, you might want to handle this differently
		}
	}

	result := command.CreateSellerCommandResult{
		Result: mapper.NewSellerResultFromValidatedEntity(validatedSeller),
	}

	return &result, nil
}

// FindAllSellers fetches all sellers
func (s *SellerService) FindAllSellers() (*query.SellerQueryListResult, error) {
	storedSellers, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	var queryResult query.SellerQueryListResult
	for _, seller := range storedSellers {
		queryResult.Result = append(queryResult.Result, mapper.NewSellerResultFromEntity(seller))
	}

	return &queryResult, nil
}

// FindSellerById fetches a specific seller by Id
func (s *SellerService) FindSellerById(id uuid.UUID) (*query.SellerQueryResult, error) {
	storedSeller, err := s.repo.FindById(id)
	if err != nil {
		return nil, err
	}

	var queryResult query.SellerQueryResult
	queryResult.Result = mapper.NewSellerResultFromEntity(storedSeller)

	return &queryResult, nil
}

// UpdateSeller updates a seller
func (s *SellerService) UpdateSeller(updateCommand *command.UpdateSellerCommand) (*command.UpdateSellerCommandResult, error) {
	seller, err := s.repo.FindById(updateCommand.Id)
	if err != nil {
		return nil, err
	}

	if seller == nil {
		return nil, errors.New("seller not found")
	}

	if err := seller.UpdateName(updateCommand.Name); err != nil {
		return nil, err
	}

	validatedUpdatedSeller, err := entities.NewValidatedSeller(seller)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.Update(validatedUpdatedSeller)
	if err != nil {
		return nil, err
	}

	result := command.UpdateSellerCommandResult{
		Result: mapper.NewSellerResultFromEntity(seller),
	}

	return &result, nil
}

func (s *SellerService) DeleteSeller(id uuid.UUID) error {
	return s.repo.Delete(id)
}
