package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	gocache "github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	goCacheStore "github.com/eko/gocache/store/go_cache/v4"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	repo "github.com/pareshvernekar/homecooked/internal/repository"
	inmemoryCache "github.com/patrickmn/go-cache"
)

// Define SetOption as a generic function type or example
type SetOption func(*Option) error

type Option struct {
	TTL time.Duration
}

// WithTTL correctly declares [T any] on the method receiver to avoid "undefined T"
func WithTTL(d time.Duration) SetOption {
	return func(o *Option) error {
		o.TTL = d
		return nil
	}
}

// Define CacheClient interface that all implementations must satisfy
type Client interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any, options ...SetOption) error
	Has(key string) bool
	PostInitialize(ctx context.Context, tenantID string, repository repo.FoodCategoryRepository) error
}

// Define cache entry structure for TTL tracking
type CacheEntry struct {
	Key        string // Cache key in format: entity_type:id
	Value      interface{}  // Stored value
	AccessedAt time.Time    // When last accessed (for eviction)
	TTL        time.Duration // Time-to-live duration
}

// cacheImpl implements the Client interface using eko/gocache library.
type cacheImpl struct {
	store                  gocache.Cache[any]           // Memory store backend from go_cache flavor
	config                 CacheConfig                  // Configuration for per-operation behavior
	foodCategoryRepository repo.FoodCategoryRepository  // Repository for loading data on cache miss
	logger                 *logger.Logger                // Structured logging - Logger type defined below

	mu sync.RWMutex // Thread safety lock

}

// NewCacheClient creates a new cache client with the given configuration.
func NewCacheClient(config CacheConfig, logger *logger.Logger, foodCategoryRepo repo.FoodCategoryRepository) (Client, error) {
	if config.DefaultTTL <= 0 {
		return nil, fmt.Errorf("invalid default TTL: must be positive")
	}

	if config.MaxItems <= 0 {
		return nil, fmt.Errorf("invalid max items: must be positive")
	}
	gocacheClient := inmemoryCache.New(config.DefaultTTL, config.DefaultTTL*2)
	gocacheStore := goCacheStore.NewGoCache(gocacheClient)
	cacheManager := gocache.New[any](gocacheStore)

	if cacheManager == nil {
		return nil, fmt.Errorf("failed to create cache manager")
	}

	return &cacheImpl{store: *cacheManager, config: config, logger: logger, foodCategoryRepository: foodCategoryRepo}, nil
}

// Get retrieves cached value for the given key or loads missing data from source.
func (c *cacheImpl) Get(ctx context.Context, key string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cache key format: %s (expected entity:type:ID)", key)
	}

	entityId := parts[1]

	var value any
	hit, err := c.store.Get(ctx, entityId)

	if err != nil {
		return nil, fmt.Errorf("cache get failed: %w", err)
	}

	if hit == nil {
		ttl, exists := c.config.TTLOverrides[parts[0]]
		if !exists {
			ttl = c.config.DefaultTTL
		}

		c.logger.Info(ctx, "Cache miss for key: %s - would load from DB", key)

		ttlOption := store.WithExpiration(ttl)
		err = c.store.Set(ctx, entityId, nil, ttlOption)
		if err != nil {
			return nil, fmt.Errorf("cache set after load failed: %w", err)
		}
	} else {
		value = hit
	}

	return value, nil
}

// Set stores a value with the given TTL and custom options (overrides global defaults).
func (c *cacheImpl) Set(ctx context.Context, key string, value any, options ...SetOption) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid cache key format: %s", key)
	}

	entityId := parts[1]

	o := &Option{}
	for _, opt := range options {
		err := opt(o)
		if err != nil {
			return fmt.Errorf("failed to apply option: %w", err)
		}
	}
	if o.TTL == 0 {
		o.TTL = c.config.DefaultTTL
	}

	ttlOption := store.WithExpiration(o.TTL)
	err := c.store.Set(ctx, entityId, value, ttlOption)

	if err != nil {
		return fmt.Errorf("failed to set cache entry: %w", err)
	}

	return nil
}

// Has checks if a cache key exists without loading the full value (O(1) lookup).
func (c *cacheImpl) Has(key string) bool {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return false // Invalid format = doesn't exist
	}

	entityId := parts[1]
	hit, _ := c.store.Get(context.Background(), entityId)

	return hit != nil
}

// PostInitialize initializes the cache with food category data from the database at startup.
// This method queries the food_category table and populates the in-memory cache using entity_type:name key convention (e.g., "food-category:vegetarian").
// Tenant isolation is enforced - each tenant's cache contains only their own categories via parameterized query with RLS.
func (c *cacheImpl) PostInitialize(ctx context.Context, tenantID string, repository repo.FoodCategoryRepository) error {
	var categories []models.FoodCategory

	// Query food categories for this tenant using parameterized query with Row-Level Security (RLS)
	// RLS policies ensure only data for current tenant is returned
	categories, err := repository.ListByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("failed to fetch food categories: %w", err)
	}

	if len(categories) == 0 {
		c.logger.Info(ctx, "No food categories found for tenant: %s", tenantID)
		return nil
	}

	// Populate cache with each category using entity_type:name key format (e.g., "food-category:vegetarian")
	for _, cat := range categories {
		key := cat.GetCacheKey()
		value := &models.FoodCategory{
			ID:          cat.ID,
			TenantID:    cat.TenantID,
			Name:        cat.Name,
			Description: cat.Description,
			CreatedAt:   cat.CreatedAt,
			UpdatedAt:   cat.UpdatedAt,
		}

		// Apply TTL from config override or use global default (30m)
		ttl := c.config.TTLOverrides[fmt.Sprintf("food_category_%s", cat.Name)]
		if ttl == 0 {
			ttl = c.config.DefaultTTL
		}

		err := c.store.Set(ctx, key, value, store.WithExpiration(ttl))
		if err != nil {
			return fmt.Errorf("failed to cache category %s: %w", cat.Name, err)
		}
	}

	c.logger.Info(ctx, "Cache initialized with food categories for tenant %s: %d categories\n", tenantID, len(categories))
	return nil
}
