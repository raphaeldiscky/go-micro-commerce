// Package service defines the interface for search operations.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/repository"
)

// SearchService defines the interface for search operations.
type SearchService interface {
	IndexProduct(ctx context.Context, product *entity.ProductDocument) error
	UpdateProduct(ctx context.Context, product *entity.ProductDocument) error
	DeleteProduct(ctx context.Context, productID string) error
	SearchProducts(
		ctx context.Context,
		query *entity.SearchQuery,
	) ([]entity.SearchResult, *pkgdto.OffsetPagination, error)
	GetProduct(ctx context.Context, productID string) (*entity.ProductDocument, error)
	BulkIndexProducts(ctx context.Context, products []entity.ProductDocument) error
	InitializeIndices(ctx context.Context) error
	RefreshIndices(ctx context.Context) error
	AutoComplete(
		ctx context.Context,
		query string,
		documentType constant.DocumentType,
	) ([]string, error)
	GetSuggestions(
		ctx context.Context,
		query string,
		documentType constant.DocumentType,
	) ([]entity.SuggestionResult, error)
	ProcessProductCreated(ctx context.Context, inboxEvent *entity.InboxEvent) error
	ProcessProductUpdated(ctx context.Context, inboxEvent *entity.InboxEvent) error
	ProcessProductDeleted(ctx context.Context, inboxEvent *entity.InboxEvent) error
}

// searchService implements the SearchService interface.
type searchService struct {
	searchRepo repository.SearchRepository
	logger     logger.Logger
}

// NewSearchService creates a new search service.
func NewSearchService(
	searchRepo repository.SearchRepository,
	appLogger logger.Logger,
) SearchService {
	return &searchService{
		searchRepo: searchRepo,
		logger:     appLogger,
	}
}

// IndexProduct indexes a product document.
func (s *searchService) IndexProduct(ctx context.Context, product *entity.ProductDocument) error {
	s.logger.Infof("Indexing product: %s", product.ID)

	// Set timestamps if not already set
	now := time.Now()
	if product.CreatedAt.IsZero() {
		product.CreatedAt = now
	}

	product.UpdatedAt = now

	// Build suggest field
	s.buildProductSuggest(product)

	if err := s.searchRepo.IndexProduct(ctx, product); err != nil {
		s.logger.Errorf("Failed to index product %s: %v", product.ID, err)

		return fmt.Errorf("failed to index product: %w", err)
	}

	return nil
}

// UpdateProduct updates a product document.
func (s *searchService) UpdateProduct(ctx context.Context, product *entity.ProductDocument) error {
	s.logger.Infof("Updating product: %s", product.ID)

	product.UpdatedAt = time.Now()
	s.buildProductSuggest(product)

	if err := s.searchRepo.UpdateProduct(ctx, product); err != nil {
		s.logger.Errorf("Failed to update product %s: %v", product.ID, err)

		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// DeleteProduct deletes a product document.
func (s *searchService) DeleteProduct(ctx context.Context, productID string) error {
	s.logger.Infof("Deleting product: %s", productID)

	if err := s.searchRepo.DeleteProduct(ctx, productID); err != nil {
		s.logger.Errorf("Failed to delete product %s: %v", productID, err)

		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// SearchProducts searches for products.
func (s *searchService) SearchProducts(
	ctx context.Context,
	query *entity.SearchQuery,
) ([]entity.SearchResult, *pkgdto.OffsetPagination, error) {
	s.logger.Infof("Searching products with query: %s", query.Query)

	result, err := s.searchRepo.SearchProducts(ctx, query)
	if err != nil {
		s.logger.Errorf("Failed to search products: %v", err)

		return nil, nil, fmt.Errorf("failed to search products: %w", err)
	}

	// Create pagination metadata
	pagination := pageutils.NewOffsetPagination(
		result.Total,
		int64(result.Page),
		int64(result.PerPage),
	)

	return result.Results, pagination, nil
}

// GetProduct retrieves a product by ID.
func (s *searchService) GetProduct(
	ctx context.Context,
	productID string,
) (*entity.ProductDocument, error) {
	s.logger.Infof("Getting product: %s", productID)

	product, err := s.searchRepo.GetProduct(ctx, productID)
	if err != nil {
		s.logger.Errorf("Failed to get product %s: %v", productID, err)

		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// BulkIndexProducts performs bulk indexing of products.
func (s *searchService) BulkIndexProducts(
	ctx context.Context,
	products []entity.ProductDocument,
) error {
	s.logger.Infof("Bulk indexing %d products", len(products))

	documents := make([]entity.SearchDocument, 0, len(products))

	for i := range products {
		now := time.Now()
		if products[i].CreatedAt.IsZero() {
			products[i].CreatedAt = now
		}

		products[i].UpdatedAt = now
		s.buildProductSuggest(&products[i])
		documents = append(documents, &products[i])
	}

	if err := s.searchRepo.BulkIndex(ctx, documents); err != nil {
		s.logger.Errorf("Failed to bulk index products: %v", err)

		return fmt.Errorf("failed to bulk index products: %w", err)
	}

	return nil
}

// InitializeIndices creates all necessary indices.
func (s *searchService) InitializeIndices(ctx context.Context) error {
	s.logger.Info("Initializing search indices")

	if err := s.searchRepo.CreateIndices(ctx); err != nil {
		s.logger.Errorf("Failed to initialize indices: %v", err)

		return fmt.Errorf("failed to initialize indices: %w", err)
	}

	s.logger.Info("Successfully initialized search indices")

	return nil
}

// RefreshIndices refreshes all indices.
func (s *searchService) RefreshIndices(ctx context.Context) error {
	s.logger.Info("Refreshing search indices")

	if err := s.searchRepo.RefreshIndices(ctx); err != nil {
		s.logger.Errorf("Failed to refresh indices: %v", err)

		return fmt.Errorf("failed to refresh indices: %w", err)
	}

	return nil
}

// AutoComplete provides autocomplete functionality.
func (s *searchService) AutoComplete(
	ctx context.Context,
	query string,
	documentType constant.DocumentType,
) ([]string, error) {
	s.logger.Infof("Auto-completing query '%s' for type '%s'", query, documentType)

	results, err := s.searchRepo.AutoComplete(ctx, query, documentType)
	if err != nil {
		s.logger.Errorf("Failed to auto-complete: %v", err)

		return nil, fmt.Errorf("failed to auto-complete: %w", err)
	}

	return results, nil
}

// GetSuggestions provides enhanced suggestions.
func (s *searchService) GetSuggestions(
	ctx context.Context,
	query string,
	documentType constant.DocumentType,
) ([]entity.SuggestionResult, error) {
	s.logger.Infof("Getting suggestions for query '%s' and type '%s'", query, documentType)

	results, err := s.searchRepo.GetSuggestions(ctx, query, documentType)
	if err != nil {
		s.logger.Errorf("Failed to get suggestions: %v", err)

		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	return results, nil
}

// ProcessProductCreated processes a product created event from the inbox.
func (s *searchService) ProcessProductCreated(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
) error {
	s.logger.Infof("Processing product created event from inbox: %s", inboxEvent.ID)

	// Parse the full event envelope
	var eventEnvelope struct {
		Metadata event.Metadata              `json:"metadata"`
		Payload  event.ProductCreatedPayload `json:"payload"`
	}
	if err := json.Unmarshal(inboxEvent.Payload, &eventEnvelope); err != nil {
		return fmt.Errorf("failed to unmarshal product created event envelope: %w", err)
	}

	payload := eventEnvelope.Payload

	// Convert event payload to search document
	productDoc := &entity.ProductDocument{
		ID:       payload.ProductID,
		Name:     payload.Name,
		Price:    payload.Price,
		Quantity: payload.Quantity,
		// Other fields will be zero values since not provided in event
	}

	if err := s.IndexProduct(ctx, productDoc); err != nil {
		return fmt.Errorf("failed to index product %s: %w", payload.ProductID, err)
	}

	s.logger.Infof(
		"Successfully processed product created event for product: %s",
		payload.ProductID,
	)

	return nil
}

// ProcessProductUpdated processes a product updated event from the inbox.
func (s *searchService) ProcessProductUpdated(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
) error {
	s.logger.Infof("Processing product updated event from inbox: %s", inboxEvent.ID)

	// Parse the full event envelope
	var eventEnvelope struct {
		Metadata event.Metadata              `json:"metadata"`
		Payload  event.ProductUpdatedPayload `json:"payload"`
	}
	if err := json.Unmarshal(inboxEvent.Payload, &eventEnvelope); err != nil {
		return fmt.Errorf("failed to unmarshal product updated event envelope: %w", err)
	}

	payload := eventEnvelope.Payload

	// Convert event payload to search document
	productDoc := &entity.ProductDocument{
		ID:       payload.ProductID,
		Name:     payload.Name,
		Price:    payload.Price,
		Quantity: payload.Quantity,
		// Other fields will be zero values since not provided in event
	}

	if err := s.UpdateProduct(ctx, productDoc); err != nil {
		return fmt.Errorf("failed to update product %s: %w", payload.ProductID, err)
	}

	s.logger.Infof(
		"Successfully processed product updated event for product: %s",
		payload.ProductID,
	)

	return nil
}

// ProcessProductDeleted processes a product deleted event from the inbox.
func (s *searchService) ProcessProductDeleted(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
) error {
	s.logger.Infof("Processing product deleted event from inbox: %s", inboxEvent.ID)

	var payload event.ProductDeletedPayload
	if err := json.Unmarshal(inboxEvent.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal product deleted payload: %w", err)
	}

	if err := s.DeleteProduct(ctx, payload.ProductID.String()); err != nil {
		return fmt.Errorf("failed to delete product %s: %w", payload.ProductID, err)
	}

	s.logger.Infof(
		"Successfully processed product deleted event for product: %s",
		payload.ProductID,
	)

	return nil
}

// buildProductSuggest builds the suggest field for products.
func (s *searchService) buildProductSuggest(product *entity.ProductDocument) {
	inputs := []string{product.Name}

	product.Suggest = entity.SuggestField{
		Input: inputs,
	}
}
