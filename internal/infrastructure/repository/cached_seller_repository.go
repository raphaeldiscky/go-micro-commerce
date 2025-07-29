// Package repository provides a cached implementation of the SellerRepository.
package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-ddd-template/internal/domain/entities"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/repositories"
	"github.com/raphaeldiscky/go-ddd-template/internal/infrastructure/cache"
)

// CachedSellerRepository decorates a SellerRepository with caching capabilities.
type CachedSellerRepository struct {
	repository repositories.SellerRepository
	cache      *cache.RedisCache
	cacheTTL   time.Duration
}

// NewCachedSellerRepository creates a new cached seller repository.
func NewCachedSellerRepository(
	repository repositories.SellerRepository,
	cch *cache.RedisCache,
	cacheTTL time.Duration,
) repositories.SellerRepository {
	return &CachedSellerRepository{
		repository: repository,
		cache:      cch,
		cacheTTL:   cacheTTL,
	}
}

// Create creates a new seller and invalidates related cache entries.
func (r *CachedSellerRepository) Create(
	seller *entities.ValidatedSeller,
) (*entities.Seller, error) {
	result, err := r.repository.Create(seller)
	if err != nil {
		return nil, err
	}

	// Cache the created seller
	ctx := context.Background()
	cacheKey := r.buildSellerCacheKey(result.Id)

	if cacheErr := r.cache.SetWithTTL(ctx, cacheKey, result, r.cacheTTL); cacheErr != nil {
		// Log cache error but don't fail the operation
		log.Printf("Failed to cache seller: %v", cacheErr)
	}

	// Invalidate sellers list cache
	if invalidateErr := r.cache.Delete(ctx, "sellers:all"); invalidateErr != nil {
		log.Printf("Failed to invalidate sellers list cache: %v", invalidateErr)
	}

	return result, nil
}

// FindByID retrieves a seller by ID, using cache when available.
func (r *CachedSellerRepository) FindByID(id uuid.UUID) (*entities.Seller, error) {
	ctx := context.Background()
	cacheKey := r.buildSellerCacheKey(id)

	// Try to get from cache first
	var cachedSeller entities.Seller
	if err := r.cache.Get(ctx, cacheKey, &cachedSeller); err == nil {
		return &cachedSeller, nil
	}

	// If not in cache, get from repository
	seller, err := r.repository.FindByID(id)
	if err != nil {
		return nil, err
	}

	if seller != nil {
		// Cache the result
		if cacheErr := r.cache.SetWithTTL(ctx, cacheKey, seller, r.cacheTTL); cacheErr != nil {
			log.Printf("Failed to cache seller: %v", cacheErr)
		}
	}

	return seller, nil
}

// FindAll retrieves all sellers, using cache when available.
func (r *CachedSellerRepository) FindAll() ([]*entities.Seller, error) {
	ctx := context.Background()
	cacheKey := "sellers:all"

	// Try to get from cache first
	var cachedSellers []*entities.Seller
	if err := r.cache.Get(ctx, cacheKey, &cachedSellers); err == nil {
		return cachedSellers, nil
	}

	// If not in cache, get from repository
	sellers, err := r.repository.FindAll()
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cacheErr := r.cache.SetWithTTL(ctx, cacheKey, sellers, r.cacheTTL); cacheErr != nil {
		log.Printf("Failed to cache sellers list: %v", cacheErr)
	}

	return sellers, nil
}

// Update updates a seller and invalidates related cache entries.
func (r *CachedSellerRepository) Update(
	seller *entities.ValidatedSeller,
) (*entities.Seller, error) {
	result, err := r.repository.Update(seller)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Update cache with new data
	cacheKey := r.buildSellerCacheKey(result.Id)
	if cacheErr := r.cache.SetWithTTL(ctx, cacheKey, result, r.cacheTTL); cacheErr != nil {
		log.Printf("Failed to update cached seller: %v", cacheErr)
	}

	// Invalidate sellers list cache
	if invalidateErr := r.cache.Delete(ctx, "sellers:all"); invalidateErr != nil {
		log.Printf("Failed to invalidate sellers list cache: %v", invalidateErr)
	}

	return result, nil
}

// Delete deletes a seller and removes it from cache.
func (r *CachedSellerRepository) Delete(id uuid.UUID) error {
	err := r.repository.Delete(id)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Remove from cache
	cacheKey := r.buildSellerCacheKey(id)
	if cacheErr := r.cache.Delete(ctx, cacheKey); cacheErr != nil {
		log.Printf("Failed to remove seller from cache: %v", cacheErr)
	}

	// Invalidate sellers list cache
	if invalidateErr := r.cache.Delete(ctx, "sellers:all"); invalidateErr != nil {
		log.Printf("Failed to invalidate sellers list cache: %v", invalidateErr)
	}

	return nil
}

// buildSellerCacheKey builds a cache key for a seller.
func (r *CachedSellerRepository) buildSellerCacheKey(sellerID uuid.UUID) string {
	return fmt.Sprintf("seller:%s", sellerID.String())
}
