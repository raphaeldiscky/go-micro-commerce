// Package service provides business logic for product operations.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/pageutils"

	pkgDto "github.com/raphaeldiscky/go-micro-template/pkg/dto"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/repository"
)

// ProductServiceInterface defines the interface for product business operations.
type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error)
	GetProducts(
		ctx context.Context,
		req dto.GetProductsRequest,
	) ([]dto.ProductResponse, *pkgDto.PageMetaData, error)
	GetProductsByIDs(ctx context.Context, ids []uuid.UUID) ([]dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, req dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
}

// ProductService implements the ProductServiceInterface.
type ProductService struct {
	dataStore              repository.DataStore
	productCreatedProducer mq.KafkaProducerInterface
	productUpdatedProducer mq.KafkaProducerInterface
	productDeletedProducer mq.KafkaProducerInterface
}

// NewProductService creates a new instance of ProductService.
func NewProductService(
	dataStore repository.DataStore,
	productCreatedProducer mq.KafkaProducerInterface,
	productUpdatedProducer mq.KafkaProducerInterface,
	productDeletedProducer mq.KafkaProducerInterface,
) ProductServiceInterface {
	return &ProductService{
		dataStore:              dataStore,
		productCreatedProducer: productCreatedProducer,
		productUpdatedProducer: productUpdatedProducer,
		productDeletedProducer: productDeletedProducer,
	}
}

// CreateProduct creates a new product.
func (s *ProductService) CreateProduct(
	ctx context.Context,
	req dto.CreateProductRequest,
) (*dto.ProductResponse, error) {
	res := new(dto.ProductResponse)

	err := s.dataStore.Atomic(ctx, func(tx repository.DataStore) error {
		productRepo := tx.ProductRepository()
		// Create domain entity
		product, err := entity.NewProduct(req.Name, req.Price, req.Quantity)
		if err != nil {
			return httperror.NewInvalidRequestBodyError()
		}
		// Save to repository
		savedProduct, err := productRepo.Create(ctx, product)
		if err != nil {
			return httperror.NewInternalServerError("failed to create product")
		}

		evt := event.NewProductCreatedEvent(
			savedProduct.ID,
			savedProduct.Name,
			savedProduct.Price,
			savedProduct.Quantity,
		)

		if err := s.productCreatedProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send product created event")
		}

		res = dto.MapToProductResponse(savedProduct)

		return nil
	})
	if err != nil {
		return nil, err
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
		return nil, httperror.NewInternalServerError("failed to get product")
	}

	if product == nil {
		return nil, httperror.NewProductNotFoundError()
	}

	return dto.MapToProductResponse(product), nil
}

// GetProducts retrieves products with pagination and filtering.
func (s *ProductService) GetProducts(
	ctx context.Context,
	req dto.GetProductsRequest,
) ([]dto.ProductResponse, *pkgDto.PageMetaData, error) {
	var products []*entity.Product

	var total int64

	var err error

	productRepo := s.dataStore.ProductRepository()
	offset := pageutils.GetOffset(req.Page, req.Limit)

	products, err = productRepo.FindAll(ctx, req.Limit, offset)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to get products")
	}

	res := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		res[i] = *dto.MapToProductResponse(product)
	}

	total, err = productRepo.Count(ctx)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to count products")
	}

	metadata := pageutils.NewMetadata(total, req.Page, req.Limit)

	return res, metadata, nil
}

// GetProductsByIDs retrieves products by their IDs.
func (s *ProductService) GetProductsByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]dto.ProductResponse, error) {
	productRepo := s.dataStore.ProductRepository()

	products, err := productRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get products")
	}

	res := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		res[i] = *dto.MapToProductResponse(product)
	}

	return res, nil
}

// UpdateProduct updates an existing product.
func (s *ProductService) UpdateProduct(
	ctx context.Context,
	req dto.UpdateProductRequest,
) (*dto.ProductResponse, error) {
	res := new(dto.ProductResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		// Check if product exists
		productRepo := ds.ProductRepository()

		existingProduct, err := productRepo.FindByID(ctx, req.ID)
		if err != nil {
			return httperror.NewInternalServerError("failed to get product")
		}

		if existingProduct == nil {
			return httperror.NewProductNotFoundError()
		}

		if err := existingProduct.UpdateName(req.Name); err != nil {
			return httperror.NewBadRequestError("invalid product name")
		}

		if err := existingProduct.UpdatePrice(req.Price); err != nil {
			return httperror.NewBadRequestError("invalid product price")
		}

		if err := existingProduct.UpdateQuantity(req.Quantity); err != nil {
			return httperror.NewBadRequestError("invalid product quantity")
		}

		// Save updated product
		updatedProduct, err := productRepo.Update(ctx, existingProduct)
		if err != nil {
			return httperror.NewInternalServerError("failed to update product")
		}

		// Produce domain event
		evt := event.NewProductUpdatedEvent(
			updatedProduct.ID,
			updatedProduct.Name,
			updatedProduct.Price,
			updatedProduct.Quantity,
		)
		if err := s.productUpdatedProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send product updated event")
		}

		res = dto.MapToProductResponse(updatedProduct)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// DeleteProduct deletes a product by ID.
func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		productRepo := ds.ProductRepository()

		exists, err := productRepo.Exists(ctx, id)
		if err != nil {
			return httperror.NewInternalServerError("failed to check product existence")
		}

		if !exists {
			return httperror.NewProductNotFoundError()
		}

		// Delete product
		if err := productRepo.Delete(ctx, id); err != nil {
			return httperror.NewInternalServerError("failed to delete product")
		}

		// Produce domain event
		evt := event.NewProductDeletedEvent(id)
		if err := s.productDeletedProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send product deleted event")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
