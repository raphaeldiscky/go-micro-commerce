package services

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-ddd-template/internal/app/command"
	"github.com/raphaeldiscky/go-ddd-template/internal/app/interfaces"
	"github.com/raphaeldiscky/go-ddd-template/internal/app/mapper"
	"github.com/raphaeldiscky/go-ddd-template/internal/app/query"
	entity "github.com/raphaeldiscky/go-ddd-template/internal/domain/entity"
	event "github.com/raphaeldiscky/go-ddd-template/internal/domain/event"
	repository "github.com/raphaeldiscky/go-ddd-template/internal/domain/repository"
)

// ProductService is the service for managing products.
type ProductService struct {
	productRepository repository.ProductRepository
	sellerRepository  repository.SellerRepository
	eventPublisher    event.EventPublisher
}

// NewProductService - Constructor for the service.
func NewProductService(
	productRepository repository.ProductRepository,
	sellerRepository repository.SellerRepository,
	eventPublisher event.EventPublisher,
) interfaces.ProductService {
	return &ProductService{
		productRepository: productRepository,
		sellerRepository:  sellerRepository,
		eventPublisher:    eventPublisher,
	}
}

// CreateProduct saves a new product.
func (s *ProductService) CreateProduct(
	productCommand *command.CreateProductCommand,
) (*command.CreateProductCommandResult, error) {
	storedSeller, err := s.sellerRepository.FindByID(productCommand.SellerID)
	if err != nil {
		return nil, err
	}

	if storedSeller == nil {
		return nil, errors.New("seller not found")
	}

	validatedSeller, err := entity.NewValidatedSeller(storedSeller)
	if err != nil {
		return nil, err
	}

	newProduct := entity.NewProduct(
		productCommand.Name,
		productCommand.Price,
		validatedSeller,
	)

	validatedProduct, err := entity.NewValidatedProduct(newProduct)
	if err != nil {
		return nil, err
	}

	_, err = s.productRepository.Create(validatedProduct)
	if err != nil {
		return nil, err
	}

	// Publish ProductCreated event
	if s.eventPublisher != nil {
		productCreatedEvent := event.NewProductCreatedEvent(
			validatedProduct.ID,
			validatedProduct.Name,
			validatedProduct.Price,
			validatedProduct.Seller.ID,
			validatedProduct.Seller.Name,
		)

		ctx := context.Background()
		if publishErr := s.eventPublisher.Publish(ctx, productCreatedEvent); publishErr != nil {
			log.Printf("Failed to publish ProductCreated event: %v", publishErr)
			// Note: In production, you might want to handle this differently
		}
	}

	result := command.CreateProductCommandResult{
		Result: mapper.NewProductResultFromValidatedEntity(validatedProduct),
	}

	return &result, nil
}

// FindAllProducts finds all products.
func (s *ProductService) FindAllProducts() (*query.ProductQueryListResult, error) {
	storedProducts, err := s.productRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var queryListResult query.ProductQueryListResult
	for _, product := range storedProducts {
		queryListResult.Result = append(
			queryListResult.Result,
			mapper.NewProductResultFromEntity(product),
		)
	}

	return &queryListResult, nil
}

// FindProductByID retrieves a product by its ID.
func (s *ProductService) FindProductByID(id uuid.UUID) (*query.ProductQueryResult, error) {
	storedProduct, err := s.productRepository.FindByID(id)
	if err != nil {
		return nil, err
	}

	var queryResult query.ProductQueryResult
	queryResult.Result = mapper.NewProductResultFromEntity(storedProduct)

	return &queryResult, nil
}
