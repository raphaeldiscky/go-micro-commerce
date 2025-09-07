// Package service provides business logic for product operations.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgDto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/utils/redisutils"
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
	ReserveProducts(
		ctx context.Context,
		req dto.ReserveProductsRequest,
	) ([]dto.ProductResponse, error)
	ReleaseProducts(
		ctx context.Context,
		req dto.ReleaseProductsRequest,
	) error
	ConfirmProductsDeduction(
		ctx context.Context,
		req dto.ConfirmProductsDeductionRequest,
	) ([]dto.ProductResponse, error)
	RestoreProducts(
		ctx context.Context,
		req dto.RestoreProductsRequest,
	) ([]dto.ProductResponse, error)
}

// ProductService implements the ProductServiceInterface.
type ProductService struct {
	dataStore              repository.DataStore
	logger                 logger.Logger
	productCreatedProducer kafka.ProducerInterface
	productUpdatedProducer kafka.ProducerInterface
	productDeletedProducer kafka.ProducerInterface
}

// NewProductService creates a new instance of ProductService.
func NewProductService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	productCreatedProducer kafka.ProducerInterface,
	productUpdatedProducer kafka.ProducerInterface,
	productDeletedProducer kafka.ProducerInterface,
) ProductServiceInterface {
	return &ProductService{
		dataStore:              dataStore,
		logger:                 appLogger,
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

		evt := mq.NewProductCreatedEvent(
			savedProduct.ID,
			savedProduct.Name,
			savedProduct.Price,
			savedProduct.Quantity,
		)

		if err := s.productCreatedProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send product created event")
		}

		// Invalidate list cache when new product is created
		cacheRepo := tx.CacheRepository()

		err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
		if err != nil {
			return err
		}

		res = mapper.MapToProductResponse(savedProduct)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetProduct retrieves a product by ID with caching.
func (s *ProductService) GetProduct(
	ctx context.Context,
	id uuid.UUID,
) (*dto.ProductResponse, error) {
	cacheRepo := s.dataStore.CacheRepository()
	productRepo := s.dataStore.ProductRepository()

	// Try cache first if available
	cachedProduct, err := cacheRepo.GetProduct(ctx, id)
	if err == nil && cachedProduct != nil {
		return mapper.MapToProductResponse(cachedProduct), nil
	}

	// Cache miss or unavailable, get from database
	product, err := productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get product")
	}

	if product == nil {
		return nil, httperror.NewProductNotFoundError()
	}

	// Store in cache for future requests if cache is available
	err = cacheRepo.SetProduct(ctx, product, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	return mapper.MapToProductResponse(product), nil
}

// GetProducts retrieves products with pagination and caching.
func (s *ProductService) GetProducts(
	ctx context.Context,
	req dto.GetProductsRequest,
) ([]dto.ProductResponse, *pkgDto.PageMetaData, error) {
	cacheRepo := s.dataStore.CacheRepository()
	productRepo := s.dataStore.ProductRepository()

	s.logger.Debugf("====1====", req)
	// Try cache first if available
	cachedProducts, err := cacheRepo.GetProducts(ctx, req.Page, req.Limit)
	if err == nil && cachedProducts != nil {
		res := make([]dto.ProductResponse, len(cachedProducts))
		for i, product := range cachedProducts {
			res[i] = *mapper.MapToProductResponse(product)
		}

		// Still need to get total count for metadata (could be cached separately)
		total, err := productRepo.Count(ctx)
		if err != nil {
			return nil, nil, httperror.NewInternalServerError("failed to count products")
		}

		metadata := pageutils.NewMetadata(total, req.Page, req.Limit)

		return res, metadata, nil
	}

	s.logger.Debugf("====2====", cachedProducts)

	// Cache miss or unavailable, get from database
	offset := pageutils.GetOffset(req.Page, req.Limit)

	products, err := productRepo.FindAll(ctx, req.Limit, offset)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to get products")
	}

	s.logger.Debugf("====3====", products)

	res := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		res[i] = *mapper.MapToProductResponse(product)
	}

	total, err := productRepo.Count(ctx)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to count products")
	}

	// Store in cache for future requests if cache is available
	err = cacheRepo.SetProducts(ctx, req.Page, req.Limit, products, 15*time.Minute)
	if err != nil {
		return nil, nil, err
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
		res[i] = *mapper.MapToProductResponse(product)
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
		evt := mq.NewProductUpdatedEvent(
			updatedProduct.ID,
			updatedProduct.Name,
			updatedProduct.Price,
			updatedProduct.Quantity,
		)
		if err := s.productUpdatedProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send product updated event")
		}

		// Invalidate cache for updated product
		cacheRepo := ds.CacheRepository()

		err = cacheRepo.DeleteProduct(ctx, updatedProduct.ID)
		if err != nil {
			return err
		}

		err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
		if err != nil {
			return err
		}

		res = mapper.MapToProductResponse(updatedProduct)

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
		evt := mq.NewProductDeletedEvent(id)
		if err := s.productDeletedProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send product deleted event")
		}

		// Invalidate cache for deleted product
		cacheRepo := ds.CacheRepository()

		err = cacheRepo.DeleteProduct(ctx, id)
		if err != nil {
			return err
		}

		err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ReserveProducts reserves stock for products atomically with optimistic locking.
func (s *ProductService) ReserveProducts(
	ctx context.Context,
	req dto.ReserveProductsRequest,
) ([]dto.ProductResponse, error) {
	var err error

	var reservedProducts []*entity.Product

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		productRepo := ds.ProductRepository()

		reservations := make([]entity.ProductReservation, len(req.Items))
		for i, item := range req.Items {
			reservations[i] = entity.ProductReservation{
				ProductID:       item.ProductID,
				Quantity:        item.Quantity,
				ExpectedVersion: item.ExpectedVersion,
			}
		}

		reservedProducts, err = productRepo.ReserveProducts(ctx, reservations)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert to DTO responses
	res := make([]dto.ProductResponse, len(reservedProducts))

	// Invalidate cache for reserved products
	cacheRepo := s.dataStore.CacheRepository()
	for _, product := range reservedProducts {
		err = cacheRepo.DeleteProduct(ctx, product.ID)
		if err != nil {
			return nil, err
		}
	}

	// Invalidate list cache patterns
	err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
	if err != nil {
		return nil, err
	}

	for i, product := range reservedProducts {
		res[i] = *mapper.MapToProductResponse(product)
	}

	return res, nil
}

// ReleaseProducts releases reserved stock for products.
func (s *ProductService) ReleaseProducts(
	ctx context.Context,
	req dto.ReleaseProductsRequest,
) error {
	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		productRepo := ds.ProductRepository()

		reservations := make([]entity.ProductReservation, len(req.Items))
		for i, item := range req.Items {
			reservations[i] = entity.ProductReservation{
				ProductID:       item.ProductID,
				Quantity:        item.Quantity,
				ExpectedVersion: item.ExpectedVersion,
			}
		}

		_, err := productRepo.ReleaseProducts(ctx, reservations)
		if err != nil {
			return fmt.Errorf("failed to release products: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheRepo := s.dataStore.CacheRepository()
	for _, item := range req.Items {
		err = cacheRepo.DeleteProduct(ctx, item.ProductID)
		if err != nil {
			return err
		}
	}

	err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
	if err != nil {
		return err
	}

	return nil
}

// ConfirmProductsDeduction confirms the stock deduction for reserved products and removes reserved quantity.
func (s *ProductService) ConfirmProductsDeduction(
	ctx context.Context,
	req dto.ConfirmProductsDeductionRequest,
) ([]dto.ProductResponse, error) {
	var updatedProducts []*entity.Product

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		productRepo := ds.ProductRepository()

		reservations := make([]entity.ProductReservation, len(req.Items))
		for i, item := range req.Items {
			reservations[i] = entity.ProductReservation{
				ProductID:       item.ProductID,
				Quantity:        item.Quantity,
				ExpectedVersion: item.ExpectedVersion,
			}
		}

		var err error

		updatedProducts, err = productRepo.ConfirmProductsDeduction(ctx, reservations)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert to DTO responses
	res := make([]dto.ProductResponse, len(updatedProducts))
	for i, product := range updatedProducts {
		res[i] = *mapper.MapToProductResponse(product)
	}

	// Invalidate cache
	cacheRepo := s.dataStore.CacheRepository()
	for _, product := range updatedProducts {
		err = cacheRepo.DeleteProduct(ctx, product.ID)
		if err != nil {
			return nil, err
		}
	}

	err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
	if err != nil {
		return nil, err
	}

	return res, nil
}

// RestoreProducts restores stock quantities for products (compensation).
func (s *ProductService) RestoreProducts(
	ctx context.Context,
	req dto.RestoreProductsRequest,
) ([]dto.ProductResponse, error) {
	var restoredProducts []*entity.Product

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		productRepo := ds.ProductRepository()

		// Get the products to restore
		productIDs := make([]uuid.UUID, len(req.Items))
		for i, item := range req.Items {
			productIDs[i] = item.ProductID
		}

		products, err := productRepo.FindByIDsForUpdate(ctx, productIDs)
		if err != nil {
			return fmt.Errorf("failed to find products for restoration: %w", err)
		}

		// Restore stock quantities
		for i, product := range products {
			restoreQuantity := req.Items[i].Quantity
			product.Quantity += restoreQuantity
			product.Version++
		}

		// Update products with restored quantities
		for _, product := range products {
			updated, err := productRepo.UpdateWithOptimisticLock(ctx, product, product.Version-1)
			if err != nil {
				return fmt.Errorf("failed to restore stocks: %w", err)
			}

			restoredProducts = append(restoredProducts, updated)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert to DTO responses
	res := make([]dto.ProductResponse, len(restoredProducts))
	for i, product := range restoredProducts {
		res[i] = *mapper.MapToProductResponse(product)
	}

	// Invalidate cache
	cacheRepo := s.dataStore.CacheRepository()
	for _, product := range restoredProducts {
		err = cacheRepo.DeleteProduct(ctx, product.ID)
		if err != nil {
			return nil, err
		}
	}

	err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
	if err != nil {
		return nil, err
	}

	return res, nil
}
