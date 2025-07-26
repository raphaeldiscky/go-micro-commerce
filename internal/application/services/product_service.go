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

type ProductService struct {
	productRepository repositories.ProductRepository
	sellerRepository  repositories.SellerRepository
	eventPublisher    events.EventPublisher
}

func NewProductService(
	productRepository repositories.ProductRepository,
	sellerRepository repositories.SellerRepository,
	eventPublisher events.EventPublisher,
) interfaces.ProductService {
	return &ProductService{
		productRepository: productRepository,
		sellerRepository:  sellerRepository,
		eventPublisher:    eventPublisher,
	}
}

func (s *ProductService) CreateProduct(productCommand *command.CreateProductCommand) (*command.CreateProductCommandResult, error) {
	storedSeller, err := s.sellerRepository.FindById(productCommand.SellerId)
	if err != nil {
		return nil, err
	}

	if storedSeller == nil {
		return nil, errors.New("seller not found")
	}

	validatedSeller, err := entities.NewValidatedSeller(storedSeller)
	if err != nil {
		return nil, err
	}

	var newProduct = entities.NewProduct(
		productCommand.Name,
		productCommand.Price,
		*validatedSeller,
	)

	validatedProduct, err := entities.NewValidatedProduct(newProduct)
	if err != nil {
		return nil, err
	}

	_, err = s.productRepository.Create(validatedProduct)
	if err != nil {
		return nil, err
	}

	// Publish ProductCreated event
	if s.eventPublisher != nil {
		productCreatedEvent := events.NewProductCreatedEvent(
			validatedProduct.Id,
			validatedProduct.Name,
			validatedProduct.Price,
			validatedProduct.Seller.Id,
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

func (s *ProductService) FindAllProducts() (*query.ProductQueryListResult, error) {
	storedProducts, err := s.productRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var queryListResult query.ProductQueryListResult
	for _, product := range storedProducts {
		queryListResult.Result = append(queryListResult.Result, mapper.NewProductResultFromEntity(product))
	}

	return &queryListResult, nil
}

func (s *ProductService) FindProductById(id uuid.UUID) (*query.ProductQueryResult, error) {
	storedProduct, err := s.productRepository.FindById(id)
	if err != nil {
		return nil, err
	}

	var queryResult query.ProductQueryResult
	queryResult.Result = mapper.NewProductResultFromEntity(storedProduct)

	return &queryResult, nil
}
