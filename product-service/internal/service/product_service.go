// Package service provides business logic for product operations.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgDto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/event"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/httperror"
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
	DeductProducts(
		ctx context.Context,
		req dto.DeductProductsRequest,
	) ([]dto.ProductResponse, error)
	RestoreProducts(
		ctx context.Context,
		req dto.RestoreProductsRequest,
	) ([]dto.ProductResponse, error)
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

		// Invalidate list cache when new product is created
		cacheRepo := tx.CacheRepository()

		err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
		if err != nil {
			return err
		}

		res = dto.MapToProductResponse(savedProduct)

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
		return dto.MapToProductResponse(cachedProduct), nil
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

	return dto.MapToProductResponse(product), nil
}

// GetProducts retrieves products with pagination and caching.
func (s *ProductService) GetProducts(
	ctx context.Context,
	req dto.GetProductsRequest,
) ([]dto.ProductResponse, *pkgDto.PageMetaData, error) {
	cacheRepo := s.dataStore.CacheRepository()
	productRepo := s.dataStore.ProductRepository()

	// Try cache first if available
	cachedProducts, err := cacheRepo.GetProducts(ctx, req.Page, req.Limit)
	if err == nil && cachedProducts != nil {
		res := make([]dto.ProductResponse, len(cachedProducts))
		for i, product := range cachedProducts {
			res[i] = *dto.MapToProductResponse(product)
		}

		// Still need to get total count for metadata (could be cached separately)
		total, err := productRepo.Count(ctx)
		if err != nil {
			return nil, nil, httperror.NewInternalServerError("failed to count products")
		}

		metadata := pageutils.NewMetadata(total, req.Page, req.Limit)

		return res, metadata, nil
	}

	// Cache miss or unavailable, get from database
	offset := pageutils.GetOffset(req.Page, req.Limit)

	products, err := productRepo.FindAll(ctx, req.Limit, offset)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to get products")
	}

	res := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		res[i] = *dto.MapToProductResponse(product)
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
	// Convert DTO to repository format
	reservations := make([]entity.ProductReservation, len(req.Items))
	for i, item := range req.Items {
		reservations[i] = entity.ProductReservation{
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			ExpectedVersion: item.ExpectedVersion,
		}
	}

	var reservedProducts []*entity.Product

	var err error

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		productRepo := ds.ProductRepository()

		// Reserve stock with optimistic locking
		reservedProducts, err = productRepo.ReserveProducts(ctx, reservations)
		if err != nil {
			return err
		}

		// Invalidate cache for reserved products
		cacheRepo := ds.CacheRepository()
		for _, product := range reservedProducts {
			err = cacheRepo.DeleteProduct(ctx, product.ID)
			if err != nil {
				return err
			}
		}

		// Invalidate list cache patterns
		err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
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
	for i, product := range reservedProducts {
		res[i] = *dto.MapToProductResponse(product)
	}

	return res, nil
}

// ReleaseProducts releases reserved stock for products.
func (s *ProductService) ReleaseProducts(
	ctx context.Context,
	req dto.ReleaseProductsRequest,
) error {
	return s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		productRepo := ds.ProductRepository()

		// Get the products to release
		productIDs := make([]uuid.UUID, len(req.Items))
		for i, item := range req.Items {
			productIDs[i] = item.ProductID
		}

		products, err := productRepo.FindByIDsForUpdate(ctx, productIDs)
		if err != nil {
			return fmt.Errorf("failed to find products for release: %w", err)
		}

		// Release reserved quantities
		for i, product := range products {
			releaseQuantity := req.Items[i].Quantity
			if product.ReservedQuantity < releaseQuantity {
				return fmt.Errorf("insufficient reserved quantity for product %s", product.ID)
			}

			product.ReservedQuantity -= releaseQuantity
			product.Version++
		}

		// Update products with released quantities
		for _, product := range products {
			_, err = productRepo.UpdateWithOptimisticLock(ctx, product, product.Version-1)
			if err != nil {
				return fmt.Errorf("failed to release products: %w", err)
			}
		}

		// Invalidate cache
		cacheRepo := ds.CacheRepository()
		for _, product := range products {
			err = cacheRepo.DeleteProduct(ctx, product.ID)
			if err != nil {
				return err
			}
		}

		err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
		if err != nil {
			return err
		}

		return nil
	})
}

// DeductProducts confirms the stock deduction for reserved products.
func (s *ProductService) DeductProducts(
	ctx context.Context,
	req dto.DeductProductsRequest,
) ([]dto.ProductResponse, error) {
	var updatedProducts []*entity.Product

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		productRepo := ds.ProductRepository()

		// Get the products to confirm
		productIDs := make([]uuid.UUID, len(req.Items))
		for i, item := range req.Items {
			productIDs[i] = item.ProductID
		}

		products, err := productRepo.FindByIDsForUpdate(ctx, productIDs)
		if err != nil {
			return fmt.Errorf("failed to find products for confirmation: %w", err)
		}

		// Confirm stock deduction by reducing both quantity and reserved quantity
		for i, product := range products {
			deductionQuantity := req.Items[i].Quantity
			if product.ReservedQuantity < deductionQuantity {
				return fmt.Errorf("insufficient reserved quantity for product %s", product.ID)
			}

			if product.Quantity < deductionQuantity {
				return fmt.Errorf("insufficient total quantity for product %s", product.ID)
			}

			product.Quantity -= deductionQuantity
			product.ReservedQuantity -= deductionQuantity
			product.Version++
		}

		// Update products with confirmed deductions
		for _, product := range products {
			updated, err := productRepo.UpdateWithOptimisticLock(ctx, product, product.Version-1)
			if err != nil {
				return fmt.Errorf("failed to confirm stock deduction: %w", err)
			}

			updatedProducts = append(updatedProducts, updated)
		}

		// Invalidate cache
		cacheRepo := ds.CacheRepository()
		for _, product := range updatedProducts {
			err = cacheRepo.DeleteProduct(ctx, product.ID)
			if err != nil {
				return err
			}
		}

		err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
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
		res[i] = *dto.MapToProductResponse(product)
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

		// Invalidate cache
		cacheRepo := ds.CacheRepository()
		for _, product := range restoredProducts {
			err = cacheRepo.DeleteProduct(ctx, product.ID)
			if err != nil {
				return err
			}
		}

		err = cacheRepo.DeleteProductsPattern(ctx, redisutils.NewCacheListProductsPatternKey())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert to DTO responses
	res := make([]dto.ProductResponse, len(restoredProducts))
	for i, product := range restoredProducts {
		res[i] = *dto.MapToProductResponse(product)
	}

	return res, nil
}
