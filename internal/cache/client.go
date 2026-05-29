package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	gocache "github.com/eko/gocache/lib/v4/cache"
	lib_store "github.com/eko/gocache/lib/v4/store"
	go_cache "github.com/eko/gocache/store/go_cache/v4"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	models "github.com/pareshvernekar/homecooked/internal/models"
)

// =============================================================================
// GENERIC ENTITY INTERFACE (Go 1.18+ Generics)
// =============================================================================
type EntityRepository[T any] interface {
	ListByTenant(tenantID string) ([]T, error)
}

// =============================================================================
// ENTITY WITH CACHE KEY SUPPORT (Interface Constraint for Generic Cache Client)
// =============================================================================
type CacheableEntity interface {
	GetCacheKey() string
	ID() string
}

// =============================================================================
// MULTI-ENTITY CACHE INTERFACE (Go 1.18+ Generics)
// =============================================================================
type TypedClient[T any] interface {
	Get(ctx context.Context, key string) (T, error)
	Set(ctx context.Context, key string, value T, options ...SetOption) error
	Has(key string) bool
	PostInitialize(ctx context.Context, tenantID string, repos map[string]any) error
}

// SetOption is a function that modifies cache set options.
type SetOption func(*Option) error

// Option holds configuration for cache operations.
type Option struct {
	TTL time.Duration
}

// WithTTL sets the TTL duration for a cache operation.
func WithTTL(d time.Duration) SetOption {
	return func(o *Option) error {
		o.TTL = d
		return nil
	}
}

// ConvertSetOption converts SetOption to lib_store.Option
func ConvertSetOption(opt SetOption) lib_store.Option {

	var optImpl Option
	err := opt(&optImpl)
	if err != nil {
		return nil // or handle error as needed
	}
	if optImpl.TTL != 0 {
		return lib_store.WithExpiration(optImpl.TTL)
	}
	return nil
}

// =============================================================================
// GENERIC CACHE CLIENT FOR MULTI-ENTITY SUPPORT (Go 1.18+)
// Uses gocache with generic type parameter T for values. Keys are always strings.
// =============================================================================
type GenericCacheClient[T any] struct {
	store    CacheStore     // Type-safe cache store - underlying implementation is gocache.Cache[any]
	config   CacheConfig    // Per-entity TTL overrides and global settings
	logger   *logger.Logger // Structured logging
	maxItems int            // LRU eviction threshold
	repos    map[string]any // Map of entity_type -> repository
	mu       sync.RWMutex   // Protects concurrent access to repos map
}

// CacheStore wraps gocache.Cache[any] to provide a store interface for the generic cache client.
type CacheStore struct {
	store *gocache.Cache[any]
}

// Get retrieves a value from the underlying gocache store.
func (cs *CacheStore) Get(ctx context.Context, key any) (interface{}, error) {
	result, err := cs.store.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("cache get failed: %w", err)
	}
	return result, nil
}

// Set stores a value in the underlying gocache store.
func (cs *CacheStore) Set(ctx context.Context, key any, val interface{}, options ...SetOption) error {
	var setOpts []lib_store.Option
	for _, opt := range options {
		storeOpt := ConvertSetOption(opt)
		setOpts = append(setOpts, storeOpt)
	}

	err := cs.store.Set(ctx, key, val, setOpts...)
	if err != nil {
		return fmt.Errorf("cache set failed: %w", err)
	}
	return nil
}

// PostInitialize initializes the cache with data from a tenant's repositories.
func (c *GenericCacheClient[T]) PostInitialize(ctx context.Context, tenantID string, repos map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(repos) == 0 {
		return fmt.Errorf("no repositories registered for GenericCacheClient")
	}

	c.repos = repos

	for entityTypeName, repo := range c.repos {
		items, err := c.fetchItemsForEntity(entityTypeName, repo, tenantID)
		if err != nil {
			return fmt.Errorf("failed to fetch %s for tenant %s: %w", entityTypeName, tenantID, err)
		}

		if len(items) == 0 {
			continue // No entities to cache for this type
		}

		for _, item := range items {
			key := c.extractCacheKey(item)

			ttl, exists := c.config.TTLOverrides[entityTypeName]
			if !exists {
				ttl = c.config.DefaultTTL
			}

			var setErr error
			switch typedItem := any(item).(type) {
			case T:
				setErr = c.store.Set(ctx, key, typedItem, WithTTL(ttl))
			default:
				return fmt.Errorf("item is not of expected type T")
			}

			if setErr != nil {
				return fmt.Errorf("failed to cache %s: %w", entityTypeName, setErr)
			}
		}
	}

	return nil
}

// fetchItemsForEntity safely fetches items for a specific entity type using runtime type assertion.
func (c *GenericCacheClient[T]) fetchItemsForEntity(entityTypeName string, repo any, tenantID string) ([]T, error) {
	switch r := repo.(type) {
	case EntityRepository[T]:
		return r.ListByTenant(tenantID)
	default:
		return nil, fmt.Errorf("repository for %s does not implement EntityRepository[T]", entityTypeName)
	}
}

// extractCacheKey safely extracts the cache key from a typed entity.
func (c *GenericCacheClient[T]) extractCacheKey(item T) string {
	var key string
	if ce, ok := any(item).(CacheableEntity); ok {
		key = ce.GetCacheKey()
	}
	return key
}

// Get retrieves a value of type T from the cache using entity_type:ID format.
func (c *GenericCacheClient[T]) Get(ctx context.Context, key string) (T, error) {
	var zero T
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return zero, fmt.Errorf("invalid cache key format: %s (expected entity_type:ID)", key)
	}

	entityType := parts[0]

	c.mu.RLock()
	cachedRepo, exists := c.repos[entityType]
	c.mu.RUnlock()
	if !exists {
		return zero, fmt.Errorf("repository for entity type '%s' not found", entityType)
	}

	hit, err := c.store.Get(ctx, key)
	if err != nil {
		return zero, fmt.Errorf("cache get failed: %w", err)
	}

	// Check if cache miss and load from repository if needed
	if hit == nil {
		items, err := c.fetchItemsForEntity(entityType, cachedRepo, "placeholder")
		if err != nil {
			return zero, fmt.Errorf("failed to fetch from repository: %w", err)
		}
		if len(items) == 0 {
			return zero, fmt.Errorf("no items found for entity ID %s", key)
		}

		// Store the item back with TTL
		ttl, exists := c.config.TTLOverrides[entityType]
		if !exists {
			ttl = c.config.DefaultTTL
		}
		var setErr error
		switch typedItem := any(items[0]).(type) {
		case T:
			setErr = c.store.Set(ctx, key, typedItem, WithTTL(ttl))
		default:
			return zero, fmt.Errorf("item is not of expected type T")
		}

		if setErr != nil {
			return zero, fmt.Errorf("failed to cache %s: %w", entityType, setErr)
		}
	}

	var result T
	result, ok := hit.(T)
	if !ok {
		return zero, fmt.Errorf("cached value is not of type %s", entityType)
	}

	return result, nil
}

// Set stores a value of type T in the cache with the given key and TTL.
func (c *GenericCacheClient[T]) Set(ctx context.Context, key string, value T, options ...SetOption) error {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid cache key format: %s", key)
	}

	err := c.store.Set(ctx, key, value, options...)
	if err != nil {
		return fmt.Errorf("failed to set cache entry: %w", err)
	}

	return nil
}

// Has checks if a value of type T exists in the cache using entity_type:ID format.
func (c *GenericCacheClient[T]) Has(key string) bool {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return false
	}

	c.mu.RLock()
	var err error
	_, err = c.store.Get(context.Background(), key)
	c.mu.RUnlock()
	if err != nil {
		return false
	}

	return true
}

// =============================================================================
// FACTORY FUNCTIONS (Go 1.18+ Generics)
// =============================================================================

// NewGenericFoodCategoryCache creates a new GenericCacheClient for FoodCategory entity type.
func NewGenericFoodCategoryCache(config CacheConfig, l *logger.Logger) TypedClient[models.FoodCategory] {
	store := CacheStore{
		store: gocache.New[any](go_cache.NewGoCache(nil)),
	}

	return &GenericCacheClient[models.FoodCategory]{
		store:    store,
		config:   config,
		logger:   l,
		maxItems: config.MaxItems,
		repos:    make(map[string]any),
	}
}

// NewGenericFoodItemCache creates a new GenericCacheClient for FoodItem entity type.
func NewGenericFoodItemCache(config CacheConfig, l *logger.Logger) TypedClient[models.FoodItem] {
	store := CacheStore{
		store: gocache.New[any](go_cache.NewGoCache(nil)),
	}

	return &GenericCacheClient[models.FoodItem]{
		store:    store,
		config:   config,
		logger:   l,
		maxItems: config.MaxItems,
		repos:    make(map[string]any),
	}
}
