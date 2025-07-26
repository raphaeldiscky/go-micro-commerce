package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd/internal/domain/entities"
	"github.com/raphaeldiscky/go-ddd/internal/domain/repositories"
	"github.com/raphaeldiscky/go-ddd/internal/infrastructure/cache"
)

// CachedProductRepository decorates a ProductRepository with caching capabilities
type CachedProductRepository struct {
	repository repositories.ProductRepository
	cache      *cache.RedisCache
	cacheTTL   time.Duration
}

// NewCachedProductRepository creates a new cached product repository
func NewCachedProductRepository(
	repository repositories.ProductRepository,
	cache *cache.RedisCache,
	cacheTTL time.Duration,
) repositories.ProductRepository {
	return &CachedProductRepository{
		repository: repository,
		cache:      cache,
		cacheTTL:   cacheTTL,
	}
}

// Create creates a new product and invalidates related cache entries
func (r *CachedProductRepository) Create(product *entities.ValidatedProduct) (*entities.Product, error) {
	result, err := r.repository.Create(product)
	if err != nil {
		return nil, err
	}

	// Cache the created product
	ctx := context.Background()
	cacheKey := r.buildProductCacheKey(result.Id)
	if cacheErr := r.cache.SetWithTTL(ctx, cacheKey, result, r.cacheTTL); cacheErr != nil {
		// Log cache error but don't fail the operation
		fmt.Printf("Failed to cache product: %v\n", cacheErr)
	}

	// Invalidate products list cache
	if invalidateErr := r.cache.Delete(ctx, "products:all"); invalidateErr != nil {
		fmt.Printf("Failed to invalidate products list cache: %v\n", invalidateErr)
	}

	return result, nil
}

// FindById retrieves a product by ID, using cache when available
func (r *CachedProductRepository) FindById(id uuid.UUID) (*entities.Product, error) {
	ctx := context.Background()
	cacheKey := r.buildProductCacheKey(id)

	// Try to get from cache first
	var cachedProduct entities.Product
	if err := r.cache.Get(ctx, cacheKey, &cachedProduct); err == nil {
		return &cachedProduct, nil
	}

	// If not in cache, get from repository
	product, err := r.repository.FindById(id)
	if err != nil {
		return nil, err
	}

	if product != nil {
		// Cache the result
		if cacheErr := r.cache.SetWithTTL(ctx, cacheKey, product, r.cacheTTL); cacheErr != nil {
			fmt.Printf("Failed to cache product: %v\n", cacheErr)
		}
	}

	return product, nil
}

// FindAll retrieves all products, using cache when available
func (r *CachedProductRepository) FindAll() ([]*entities.Product, error) {
	ctx := context.Background()
	cacheKey := "products:all"

	// Try to get from cache first
	var cachedProducts []*entities.Product
	if err := r.cache.Get(ctx, cacheKey, &cachedProducts); err == nil {
		return cachedProducts, nil
	}

	// If not in cache, get from repository
	products, err := r.repository.FindAll()
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cacheErr := r.cache.SetWithTTL(ctx, cacheKey, products, r.cacheTTL); cacheErr != nil {
		fmt.Printf("Failed to cache products list: %v\n", cacheErr)
	}

	return products, nil
}

// Update updates a product and invalidates related cache entries
func (r *CachedProductRepository) Update(product *entities.ValidatedProduct) (*entities.Product, error) {
	result, err := r.repository.Update(product)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Update cache with new data
	cacheKey := r.buildProductCacheKey(result.Id)
	if cacheErr := r.cache.SetWithTTL(ctx, cacheKey, result, r.cacheTTL); cacheErr != nil {
		fmt.Printf("Failed to update cached product: %v\n", cacheErr)
	}

	// Invalidate products list cache
	if invalidateErr := r.cache.Delete(ctx, "products:all"); invalidateErr != nil {
		fmt.Printf("Failed to invalidate products list cache: %v\n", invalidateErr)
	}

	return result, nil
}

// Delete deletes a product and removes it from cache
func (r *CachedProductRepository) Delete(id uuid.UUID) error {
	err := r.repository.Delete(id)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Remove from cache
	cacheKey := r.buildProductCacheKey(id)
	if cacheErr := r.cache.Delete(ctx, cacheKey); cacheErr != nil {
		fmt.Printf("Failed to remove product from cache: %v\n", cacheErr)
	}

	// Invalidate products list cache
	if invalidateErr := r.cache.Delete(ctx, "products:all"); invalidateErr != nil {
		fmt.Printf("Failed to invalidate products list cache: %v\n", invalidateErr)
	}

	return nil
}

// buildProductCacheKey builds a cache key for a product
func (r *CachedProductRepository) buildProductCacheKey(productID uuid.UUID) string {
	return fmt.Sprintf("product:%s", productID.String())
}
