// Package service provides business logic for product operations.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/dto"
	entity "github.com/raphaeldiscky/go-micro-template/product-service/internal/entity"
	event "github.com/raphaeldiscky/go-micro-template/product-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/log"
	repository "github.com/raphaeldiscky/go-micro-template/product-service/internal/repository"
)

// ProductServiceInterface defines the interface for product business operations.
type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error)
	GetProducts(ctx context.Context, req dto.GetProductsRequest) (*dto.ProductListResponse, error)
	UpdateProduct(ctx context.Context, req dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
}

// ProductService implements the ProductServiceInterface.
type ProductService struct {
	dataStore repository.DataStore
	producer  event.Producer
	topics    constant.ProductTopics
	logger    log.LoggerInterface
}

// NewProductService creates a new instance of ProductService.
func NewProductService(
	dataStore repository.DataStore,
	producer event.Producer,
	topics constant.ProductTopics,
	logger log.LoggerInterface,
) ProductServiceInterface {
	return &ProductService{
		dataStore: dataStore,
		producer:  producer,
		topics:    topics,
		logger:    logger,
	}
}

// CreateProduct creates a new product.
func (s *ProductService) CreateProduct(
	ctx context.Context,
	req dto.CreateProductRequest,
) (*dto.ProductResponse, error) {
	res := new(dto.ProductResponse)
	err := s.dataStore.WithTransaction(ctx, func(tx repository.DataStore) error {
		productRepo := tx.ProductRepository()
		// Create domain entity
		product, err := entity.NewProduct(req.Name, req.Price, req.Quantity)
		if err != nil {
			return fmt.Errorf("failed to create product entity: %w", err)
		}
		// Save to repository
		savedProduct, err := productRepo.Create(ctx, product)
		if err != nil {
			return fmt.Errorf("failed to save product: %w", err)
		}
		// Produce domain event
		if s.producer != nil {
			evt := event.NewProductCreatedEvent(
				savedProduct.ID,
				savedProduct.Name,
				savedProduct.Price,
				savedProduct.Quantity,
			)
			topic := s.topics.ProductLifecycle

			if err := s.producer.Produce(topic, evt); err != nil {
				// Log error but don't fail the operation
				// In production, you might want to implement event outbox pattern
				s.logger.Errorf("failed to produce ProductCreated event: %v", err)
			}
		}

		res = s.mapToResponse(savedProduct)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return res, nil
}

// GetProduct retrieves a product by ID.
func (s *ProductService) GetProduct(
	ctx context.Context,
	id uuid.UUID,
) (*dto.ProductResponse, error) {
	productRepo := s.dataStore.ProductRepository()
	product, err := productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	return s.mapToResponse(product), nil
}

// GetProducts retrieves products with pagination and filtering.
func (s *ProductService) GetProducts(
	ctx context.Context,
	req dto.GetProductsRequest,
) (*dto.ProductListResponse, error) {
	var products []*entity.Product

	var total int64

	var err error
	productRepo := s.dataStore.ProductRepository()
	products, err = productRepo.FindAll(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	total, err = productRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = *s.mapToResponse(product)
	}

	return &dto.ProductListResponse{
		Products: productResponses,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}, nil
}

// UpdateProduct updates an existing product.
func (s *ProductService) UpdateProduct(
	ctx context.Context,
	req dto.UpdateProductRequest,
) (*dto.ProductResponse, error) {
	res := new(dto.ProductResponse)
	err := s.dataStore.WithTransaction(ctx, func(tx repository.DataStore) error {
		// Check if product exists
		productRepo := s.dataStore.ProductRepository()
		existingProduct, err := productRepo.FindByID(ctx, req.ID)
		if err != nil {
			return fmt.Errorf("failed to find product: %w", err)
		}

		if existingProduct == nil {
			return fmt.Errorf("product not found")
		}

		// Update fields
		if err := existingProduct.UpdateName(req.Name); err != nil {
			return fmt.Errorf("failed to update product name: %w", err)
		}

		if err := existingProduct.UpdatePrice(req.Price); err != nil {
			return fmt.Errorf("failed to update product price: %w", err)
		}

		if err := existingProduct.UpdateQuantity(req.Quantity); err != nil {
			return fmt.Errorf("failed to update product quantity: %w", err)
		}

		// Save updated product
		updatedProduct, err := productRepo.Update(ctx, existingProduct)
		if err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		// Produce domain event
		if s.producer != nil {
			evt := event.NewProductUpdatedEvent(
				updatedProduct.ID,
				updatedProduct.Name,
				updatedProduct.Price,
				updatedProduct.Quantity,
			)
			topic := s.topics.ProductLifecycle

			if err := s.producer.Produce(topic, evt); err != nil {
				s.logger.Errorf("failed to produce ProductUpdated event: %v", err)
			}
		}
		res = s.mapToResponse(updatedProduct)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return res, nil
}

// DeleteProduct deletes a product by ID.
func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	err := s.dataStore.WithTransaction(ctx, func(tx repository.DataStore) error {
		productRepo := tx.ProductRepository()
		exists, err := productRepo.Exists(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to check product existence: %w", err)
		}

		if !exists {
			return fmt.Errorf("product not found")
		}

		// Delete product
		if err := productRepo.Delete(ctx, id); err != nil {
			return fmt.Errorf("failed to delete product: %w", err)
		}

		// Produce domain event
		if s.producer != nil {
			evt := event.NewProductDeletedEvent(id)
			topic := s.topics.ProductLifecycle

			if err := s.producer.Produce(topic, evt); err != nil {
				s.logger.Errorf("failed to produce ProductDeleted event: %v", err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// mapToResponse converts domain entity to DTO response.
func (s *ProductService) mapToResponse(product *entity.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		Quantity:  product.Quantity,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}
