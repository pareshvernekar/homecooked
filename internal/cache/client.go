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
	inmemoryCache "github.com/patrickmn/go-cache"
)

// Define SetOption as a generic function type or interface (as an example)
type SetOption func(*Option) error

type Option struct {
	TTL time.Duration
	// Add other fields here
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
}

// Define cache entry structure for TTL tracking
type CacheEntry struct {
	Key        string        // Cache key in format: entity_type:id
	Value      interface{}   // Stored value
	AccessedAt time.Time     // When last accessed (for eviction)
	TTL        time.Duration // Time-to-live duration
}

// cacheImpl implements the Client interface using eko/gocache library.
type cacheImpl struct {
	store  *gocache.Cache[any] // Memory store backend from go_cache flavor
	config CacheConfig         // Configuration for per-operation behavior
	mu     sync.RWMutex        // Thread safety lock
}

// NewCacheClient creates a new cache client with the given configuration.
// Correct eko/gocache initialization pattern:
//
//	gocacheClient := inmemoryCache.New(defaultTTL, maxTTL)        // Default & max TTL params
//	gocacheStore := goCacheStore.NewGoCache(gocacheClient)  // Memory store backend implementation
//	cacheManager := gocache.New[string, any](gocacheStore)  // Cache client instance
func NewCacheClient(config CacheConfig) (Client, error) {
	// Create eko/gocache client with correct API pattern from docs:
	if config.DefaultTTL <= 0 {
		return nil, fmt.Errorf("invalid default TTL: must be positive")
	}

	if config.MaxItems <= 0 {
		return nil, fmt.Errorf("invalid max items: must be positive")
	}
	gocacheClient := inmemoryCache.New(config.DefaultTTL, config.DefaultTTL*2) // Default & max TTL params
	gocacheStore := goCacheStore.NewGoCache(gocacheClient)                     // Memory store backend implementation
	cacheManager := gocache.New[any](gocacheStore)                             // Cache client instance

	if cacheManager == nil {
		return nil, fmt.Errorf("failed to create cache manager")
	}

	return &cacheImpl{store: cacheManager, config: config}, nil
}

// Get retrieves cached value for the given key or loads missing data from source.
// Uses O(1) complexity map lookup - if hit, instant return; if miss, triggers DB query with TTL applied.
func (c *cacheImpl) Get(ctx context.Context, key string) (any, error) {
	c.mu.RLock() // Read lock for safe concurrent access
	defer c.mu.RUnlock()

	// Parse cache key format: "entity_type:id"
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cache key format: %s (expected entity:type:ID)", key)
	}

	entityType := parts[0]
	entityId := parts[1]

	// Get from memory store using eko/gocache API
	var value any
	hit, err := c.store.Get(ctx, entityId)

	if err != nil {
		return nil, fmt.Errorf("cache get failed: %w", err)
	}

	if hit == nil {
		// Cache miss - determine TTL for this entity type
		ttl, exists := c.config.TTLOverrides[entityType]
		if !exists {
			ttl = c.config.DefaultTTL
		}

		// Load from database (placeholder) and store with custom TTL
		value, err := c.loadFromDatabase(ctx, entityType, entityId)
		if err != nil {
			return nil, fmt.Errorf("db load failed: %w", err)
		}

		// Store with per-entity TTL override applied
		ttlOption := store.WithExpiration(ttl)
		err = c.store.Set(ctx, entityId, value, ttlOption)
		if err != nil {

			return nil, fmt.Errorf("cache set after load failed: %w", err)
		}
	} else {
		value = hit
	}

	// Update access time for LRU eviction tracking
	keyStr := fmt.Sprintf("%s:%s", entityType, entityId)
	c.updateAccessTime(keyStr)
	return value, nil
}

// Set stores a value with the given TTL and custom options (overrides global defaults).
// Variadic options pattern allows per-operation configuration: WithTTL(30*time.Minute), etc.
func (c *cacheImpl) Set(ctx context.Context, key string, value any, options ...SetOption) error {
	c.mu.Lock() // Write lock - exclusive access for safe eviction handling
	defer c.mu.Unlock()

	// Parse cache key format: "entity_type:id"
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid cache key format: %s", key)
	}

	entityType := parts[0]
	entityId := parts[1]

	// Determine TTL - override from option or use global default
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

	// Store with custom TTL applied via eko/gocache WithTTL option
	ttlOption := store.WithExpiration(o.TTL)
	err := c.store.Set(ctx, entityId, value, ttlOption)

	if err != nil {
		return fmt.Errorf("failed to set cache entry: %w", err)
	}

	// Track access time for eviction tracking
	keyStr := fmt.Sprintf("%s:%s", entityType, entityId)
	c.updateAccessTime(keyStr)
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

// updateAccessTime records when a key was last accessed for eviction tracking.
func (c *cacheImpl) updateAccessTime(key string) {
	// In real implementation, would track last access time per key in partition
	// This placeholder ensures thread-safety pattern works correctly
	_ = key // Suppress unused variable warning in test builds
}

// Helper functions for actual implementation
func (c *cacheImpl) loadFromDatabase(ctx context.Context, entityType, entityId string) (*CacheEntry, error) {
	// Placeholder - in real implementation would query database
	return &CacheEntry{
		Key:        entityType + ":" + entityId,
		Value:      nil,
		AccessedAt: time.Now(),
		TTL:        c.config.DefaultTTL,
	}, nil
}
