package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/app/dto"
	entity "github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/domain/entity"
	event "github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/domain/event"
	repository "github.com/raphaeldiscky/go-ddd-template/services/product-service/internal/domain/repository"
)

// ProductServiceInterface defines the interface for product business operations
type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error)
	GetProducts(ctx context.Context, req dto.GetProductsRequest) (*dto.ProductListResponse, error)
	UpdateProduct(ctx context.Context, req dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
}

// ProductService implements the ProductServiceInterface
type ProductService struct {
	productRepo    repository.ProductRepository
	eventPublisher event.EventPublisher
}

// NewProductService creates a new instance of ProductService
func NewProductService(productRepo repository.ProductRepository, eventPublisher event.EventPublisher) ProductServiceInterface {
	return &ProductService{
		productRepo:    productRepo,
		eventPublisher: eventPublisher,
	}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	// Create domain entity
	product, err := entity.NewProduct(req.Name, req.Price, req.SellerId)
	if err != nil {
		return nil, fmt.Errorf("failed to create product entity: %w", err)
	}

	// Save to repository
	savedProduct, err := s.productRepo.Create(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to save product: %w", err)
	}

	// Publish domain event
	if s.eventPublisher != nil {
		event := event.NewProductCreatedEvent(savedProduct.Id, savedProduct.Name, savedProduct.Price, savedProduct.SellerId)
		if err := s.eventPublisher.Publish(event); err != nil {
			// Log error but don't fail the operation
			// In production, you might want to implement event outbox pattern
			fmt.Printf("Failed to publish ProductCreated event: %v\n", err)
		}
	}

	return s.mapToResponse(savedProduct), nil
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error) {
	product, err := s.productRepo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	return s.mapToResponse(product), nil
}

// GetProducts retrieves products with pagination and filtering
func (s *ProductService) GetProducts(ctx context.Context, req dto.GetProductsRequest) (*dto.ProductListResponse, error) {
	var products []*entity.Product
	var total int64
	var err error

	if req.SellerId != nil {
		products, err = s.productRepo.FindBySellerId(ctx, *req.SellerId, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get products by seller: %w", err)
		}
		total, err = s.productRepo.CountBySellerId(ctx, *req.SellerId)
		if err != nil {
			return nil, fmt.Errorf("failed to count products by seller: %w", err)
		}
	} else {
		products, err = s.productRepo.FindAll(ctx, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get products: %w", err)
		}
		total, err = s.productRepo.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count products: %w", err)
		}
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

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(ctx context.Context, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	// Check if product exists
	existingProduct, err := s.productRepo.FindById(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	if existingProduct == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Update fields
	if err := existingProduct.UpdateName(req.Name); err != nil {
		return nil, fmt.Errorf("failed to update product name: %w", err)
	}

	if err := existingProduct.UpdatePrice(req.Price); err != nil {
		return nil, fmt.Errorf("failed to update product price: %w", err)
	}

	if err := existingProduct.UpdateSeller(req.SellerId); err != nil {
		return nil, fmt.Errorf("failed to update product seller: %w", err)
	}

	// Save updated product
	updatedProduct, err := s.productRepo.Update(ctx, existingProduct)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Publish domain event
	if s.eventPublisher != nil {
		event := event.NewProductUpdatedEvent(updatedProduct.Id, updatedProduct.Name, updatedProduct.Price, updatedProduct.SellerId)
		if err := s.eventPublisher.Publish(event); err != nil {
			fmt.Printf("Failed to publish ProductUpdated event: %v\n", err)
		}
	}

	return s.mapToResponse(updatedProduct), nil
}

// DeleteProduct deletes a product by ID
func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	// Check if product exists
	exists, err := s.productRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check product existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("product not found")
	}

	// Delete product
	if err := s.productRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	// Publish domain event
	if s.eventPublisher != nil {
		event := event.NewProductDeletedEvent(id)
		if err := s.eventPublisher.Publish(event); err != nil {
			fmt.Printf("Failed to publish ProductDeleted event: %v\n", err)
		}
	}

	return nil
}

// mapToResponse converts domain entity to DTO response
func (s *ProductService) mapToResponse(product *entity.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		Id:        product.Id,
		Name:      product.Name,
		Price:     product.Price,
		SellerId:  product.SellerId,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}
