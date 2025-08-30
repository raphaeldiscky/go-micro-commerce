package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/utils/redisutils"
)

// CacheRepositoryInterface defines the interface for cache operations.
type CacheRepositoryInterface interface {
	// GetProduct retrieves a single product from cache by ID
	GetProduct(ctx context.Context, id uuid.UUID) (*entity.Product, error)

	// SetProduct stores a single product in cache
	SetProduct(ctx context.Context, product *entity.Product, expiration time.Duration) error

	// GetProducts retrieves paginated products from cache
	GetProducts(ctx context.Context, page, limit int64) ([]*entity.Product, error)

	// SetProducts stores paginated products in cache
	SetProducts(
		ctx context.Context,
		page, limit int64,
		products []*entity.Product,
		expiration time.Duration,
	) error

	// DeleteProduct removes a product from cache by ID
	DeleteProduct(ctx context.Context, id uuid.UUID) error

	// DeleteProductsPattern removes products matching a pattern (for invalidating pagination cache)
	DeleteProductsPattern(ctx context.Context, pattern string) error
}

// CacheRepositoryRedis implements CacheRepositoryInterface using Redis.
type CacheRepositoryRedis struct {
	client redis.UniversalClient
}

// NewCacheRepositoryRedis creates a new Redis cache repository, or null repository if client is nil.
func NewCacheRepositoryRedis(client redis.UniversalClient) CacheRepositoryInterface {
	if client == nil {
		return &NullCacheRepository{}
	}

	return &CacheRepositoryRedis{
		client: client,
	}
}

// NullCacheRepository implements CacheRepositoryInterface as a no-op for when Redis is unavailable.
type NullCacheRepository struct{}

// GetProduct retrieves a single product from Redis cache.
func (r *CacheRepositoryRedis) GetProduct(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Product, error) {
	key := redisutils.NewCacheProductKey(id)

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	var product entity.Product
	if err := json.Unmarshal([]byte(result), &product); err != nil {
		return nil, err
	}

	return &product, nil
}

// SetProduct stores a single product in Redis cache.
func (r *CacheRepositoryRedis) SetProduct(
	ctx context.Context,
	product *entity.Product,
	expiration time.Duration,
) error {
	key := redisutils.NewCacheProductKey(product.ID)

	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// GetProducts retrieves paginated products from Redis cache.
func (r *CacheRepositoryRedis) GetProducts(
	ctx context.Context,
	page, limit int64,
) ([]*entity.Product, error) {
	key := redisutils.NewCacheListProductsKey(page, limit)

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss
		}

		return nil, err
	}

	var products []*entity.Product
	if err := json.Unmarshal([]byte(result), &products); err != nil {
		return nil, err
	}

	return products, nil
}

// SetProducts stores paginated products in Redis cache.
func (r *CacheRepositoryRedis) SetProducts(
	ctx context.Context,
	page, limit int64,
	products []*entity.Product,
	expiration time.Duration,
) error {
	key := redisutils.NewCacheListProductsKey(page, limit)

	data, err := json.Marshal(products)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// DeleteProduct removes a product from cache by ID.
func (r *CacheRepositoryRedis) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	key := redisutils.NewCacheProductKey(id)

	return r.client.Del(ctx, key).Err()
}

// DeleteProductsPattern removes products matching a pattern.
func (r *CacheRepositoryRedis) DeleteProductsPattern(ctx context.Context, pattern string) error {
	var cursor uint64

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

// NullCacheRepository methods - all return no-op behavior

// GetProduct always returns cache miss.
func (n *NullCacheRepository) GetProduct(
	_ context.Context,
	_ uuid.UUID,
) (*entity.Product, error) {
	return nil, nil
}

// SetProduct does nothing.
func (n *NullCacheRepository) SetProduct(
	_ context.Context,
	_ *entity.Product,
	_ time.Duration,
) error {
	return nil
}

// GetProducts always returns cache miss.
func (n *NullCacheRepository) GetProducts(
	_ context.Context,
	_, _ int64,
) ([]*entity.Product, error) {
	return nil, nil
}

// SetProducts does nothing.
func (n *NullCacheRepository) SetProducts(
	_ context.Context,
	_, _ int64,
	_ []*entity.Product,
	_ time.Duration,
) error {
	return nil
}

// DeleteProduct does nothing.
func (n *NullCacheRepository) DeleteProduct(_ context.Context, _ uuid.UUID) error {
	return nil
}

// DeleteProductsPattern does nothing.
func (n *NullCacheRepository) DeleteProductsPattern(_ context.Context, _ string) error {
	return nil
}
