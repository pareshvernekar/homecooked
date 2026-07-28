
# Cache Framework Implementation Plan

**Branch**: 20260516-174330-cache-framework  
**Created**: 2026-05-16  
**Spec**: [docs/specs/cache_framework/spec.md](../spec.md) | **Status**: Ready for Implementation

---

## Overview

This plan details the implementation of an in-memory caching framework using eko/gocache library to eliminate repeated database queries for related entity data. The cache layer stores preloaded entities with configurable TTLs and eviction policies, enabling O(1) retrieval while maintaining thread-safe access patterns.

### Key Design Decisions

1. **Storage Backend**: go-cache flavor with in-memory storage via `goCacheStore.NewGoCache()` - simplest deployment, no external dependencies
2. **Client Pattern**: Separated cache client interface from store implementation for testability and future backend swap capability
3. **Generic Type Safety**: Cache operations use Go generics for compile-time type safety (e.g., `Get[T](context.Context, string, *T)` )
4. **Configuration-Driven TTL**: Per-entity expiration configured via CacheConfig struct with per-relation overrides

---

## Constitution Check

### Gate 1: Test-Driven Development (TDD) ✅ PASS

**Requirement**: All implementation code must be developed using TDD - tests written before production code; all functionality verified by passing tests.

**Implementation Strategy**:
- Follow TDD cycle: Red → Green → Refactor for each component
- Write unit tests first, then implementation
- Use `go test` with coverage to verify all public functions tested
- Stress tests included to validate concurrent access patterns (15+ goroutines)
- Integration tests verify cache interacts correctly with database operations

**Testing Coverage**:
```
Phase 2: Core Implementation Tests
├── Unit Tests (T01-T04)
│   ├── NewCacheClient() initialization validation
│   ├── CacheClient interface conformance tests
│   ├── Set/Get/Del operation correctness tests
│   └── TTL expiration and refresh behavior
├── Integration Tests (T05-T07)
│   ├── Database cache miss triggers reload
│   ├── Concurrent access pattern validation
│   └── Memory leak detection under sustained load
└── Benchmark Tests (T08-T10)
    ├── Compare cache vs direct DB query latency
    ├── Measure O(1) lookup performance metrics
    └── Verify eviction efficiency at capacity limits
```

---

### Gate 2: Modular Architecture ✅ PASS

**Requirement**: Cache framework must be independent and swappable - separate package with well-defined interfaces.

**Implementation Approach**:
- Package `internal/cache` isolated from business logic
- Core client interface abstracts store implementation
- Repository layer injects CacheClient dependency (DI pattern)
- Swap to Redis backend requires only factory function change, zero code changes in handlers

---

### Gate 3: RESTful Design & Performance ✅ PASS

**Requirement**: Cache operations must optimize for fast read patterns; eliminate N+1 query problem.

**Performance Goals**:
- Get operation: O(1) time complexity (map lookup)
- Set operation: O(N) insertion cost acceptable (configurable TTL ensures freshness)
- Concurrent access: No performance degradation under 15+ simultaneous goroutines

---

## Phase 0: Research & Design ✅ Complete

**Completed Artifacts**:
- ✅ spec.md - Feature specification with requirements
- ✅ Clarifications integrated from 2026-05-16 session  
- ✅ This design document (phase 2)

**Research Findings Consolidated**:

**Decision 1: Use eko/gocache with go-cache flavor**
- **Rationale**: Zero external dependencies (in-memory), simple initialization, tested in production food ordering systems
- **Alternatives Considered**:
    - Redis-backed cache: Would require Redis server dependency, adds complexity for single-server deployment
    - sync.Map direct: eko/gocache provides better abstractions, test helpers, and configurable eviction policies

**Decision 2: Separate client-store interface pattern**
- **Rationale**: Enables swapping to Redis/memcached later without changing cache usage code
- **Impact**: Repository layer passes CacheClient interface via constructor injection, factory handles store instantiation

---

## Phase 1: Implementation Tasks ✅ Defined

### Task Breakdown (Test-Driven Development Order)

#### Phase 1a: Infrastructure Setup (T001-T002)

**Task T001**: Create internal/cache package directory structure
```bash
mkdir -p internal/cache
# Files to be created: cache_client.go, storage.go, config.go, eviction_policy.go, utils.go
```

**Task T002**: Initialize project documentation and dependency stubs
- Create README.md with usage examples
- Add go.mod comment header documenting eko/gocache imports

---

#### Phase 1b: Configuration Loading (T003)

**Task T003**: Implement CacheConfig struct and YAML/environment variable loading
```go
type CacheConfig struct {
    DefaultTTL               time.Duration `yaml:"default_ttl" json:"default_ttl"`                    // Global default TTL
    MaxItemsPerCache         int           `yaml:"max_items" json:"max_items"`                       // Eviction trigger threshold
    MaxIdleConns             int           `yaml:"max_idle_conns" json:"max_idle_conns"`            // Pool size (for Redis swap)
    TTLOverrides             map[string]time.Duration `yaml:"ttl_overrides" json:"ttl_overrides"`   // Per-entity overrides
    EvictionStrategy         string        `yaml:"eviction_strategy" json:"eviction_strategy"`       // "lru" or "fifo"
}
```

**Test Requirements (T004)**:
```go
// Unit tests for configuration loading
func TestLoadConfig_UsesDefaults(t *testing.T) {
    cfg, err := LoadConfig("")
    assert.NoError(t, err)
    assert.Equal(t, 30*time.Minute, cfg.DefaultTTL) // Default from spec
}

func TestLoadConfig_AppendsOverrides(t *testing.T) {
    overrides := map[string]time.Duration{
        "order.menu_items": 15 * time.Minute,
    }
    cfg, err := LoadConfig("", overrides)
    assert.NoError(t, err)
    assert.Equal(t, 15*time.Minute, cfg.TTLOverrides["order.menu_items"])
}

func TestCacheConfig_ValidatesRequiredFields(t *testing.T) {
    cfg := CacheConfig{DefaultTTL: 30 * time.Minute} // Missing MaxItemsPerCache
    err := cfg.Validate()
    assert.ErrorContains(t, err, "max_items_per_cache required")
}
```

---

#### Phase 1c: Core Client Interface (T005-T007)

**Task T005**: Define CacheClient interface with core operations
```go
// Package cache provides the caching interface for storing and retrieving relational entity data.
package cache

import (
    "context"
    "time"
)

// CacheClient is the primary cache interface abstracting store implementation.
// Implementation details are hidden - only methods declared here visible to clients.
type CacheClient interface {
    // Get retrieves cached value or triggers reload from DB on miss.
    // Returns pointer to dereference result (nil if expired/missed).
    Get[T any](ctx context.Context, key string, out *T) error
    
    // Set stores value with TTL and optional expiration policy.
    Set[T any](ctx context.Context, key string, value T, opts ...CacheOption[T]) error
    
    // Del removes key from cache entirely (bypasses TTL).
    Del(ctx context.Context, key string) error
    
    // Has checks if key exists in cache (O(1)).
    Has(key string) bool
    
    // FlushAll clears entire cache (useful for admin operations).
    FlushAll() error
}

// CacheOption is a functional options pattern for cache operations.
type CacheOption[T any] func(options *CacheOptions[T])

// WithTTL sets time-to-live for operation (overrides global default).
func WithTTL(d time.Duration) CacheOption[int] {
    return func(opts *CacheOptions[int]) {
        opts.TTL = d
    }
}

// WithEvictionPolicy configures eviction strategy (lru/fifo).
func WithEvictionPolicy(policy string) CacheOption[string] {
    // Implementation details...
}
```

**Test Requirements**:
```go
// Interface conformance tests - verify methods exist and have correct signatures
var _ CacheClient = (*NewGoCacheClient)(nil)

// Verify Get returns pointer type for optional result
func TestGet_ReturnsPointer(t *testing.T) {
    gocacheClient := gocache.New(config.DefaultTTL, config.DefaultTTL*2)
     gocacheStore := goCacheStore.NewGoCache(gocacheClient)
    cacheManager := cache.New[string, any](gocacheStore)
    
    var menuItems []MenuItem // output slice to be populated
    err := client.Get(context.Background(), "test_key", &menuItems)
    
    assert.ErrorIs(t, err, ErrCacheMiss)        // Cache miss triggers DB load
    assert.Nil(t, menuItems)                    // Output remains nil until loaded
}

// Verify Set accepts variadic options
func TestSet_AcceptsOptions(t *testing.T) {
    gocacheClient := gocache.New(config.DefaultTTL, config.DefaultTTL*2)
     gocacheStore := goCacheStore.NewGoCache(gocacheClient)
    cacheManager := cache.New[string, any](gocacheStore)
    
    menuItem := MenuItem{Name: "Test Item"}
    
    err := client.Set(
        context.Background(), 
        "test_key",
        menuItem,
        WithTTL(30*time.Minute),    // Override default
        WithEvictionPolicy("lru"),  // Configure eviction
    )
    
    assert.NoError(t, err)
}
```

---

**Task T006**: Implement basic cache operations (Set, Get, Has)
```go
// Implementation using eko/gocache store
func NewGoCacheClient(ctx context.Context, config CacheConfig) (cache.CacheClient, error) {
     // Initialize eko/gocache client with correct API pattern from docs:
     gocacheClient := gocache.New(config.DefaultTTL, config.DefaultTTL*2)         // Default and max TTL params
     gocacheStore := goCacheStore.NewGoCache(gocacheClient)                        // Memory store backend
     cacheManager := cache.New[string, any](gocacheStore)                         // Cache client instance
     return cacheManager, nil
}

// Core Get implementation - delegates to store.Get
func (c *GoCacheClient) Get[T any](ctx context.Context, key string, out *T) error {
    var empty T

    
    err := c.client.Get(
        ctx, 
        key,
        &empty,                    // Load missing value into output pointer
        gocache.WithWait(0),      // No wait - return immediately on miss
     )
    
    if err != nil {
         // Cache miss - could trigger DB load here if using pattern of auto-refresh
         return ErrCacheMiss
     }
     
    return nil
}

// Core Set implementation - delegates to store.Set
func (c *GoCacheClient) Set[T any](ctx context.Context, key string, value T, opts ...CacheOption[T]) error {
    options := &CacheOptions[int]{TTL: 30 * time.Minute} // Apply defaults
    
    for _, opt := range opts {
        opt(options) // Apply variadic overrides
     }
    
    err := c.client.Set(ctx, key, value, options...)
    return err
}
```

---

**Task T007**: Implement TTL-based expiration and refresh behavior
```go
// Refresh operation - reloads expired keys from database
func (c *GoCacheClient) Refresh[T any](ctx context.Context, key string, loader func() (*T, error)) error {
    if c.Has(key) {
        return nil // Already fresh - skip refresh
     }
    
    var out T
    err := c.Get(ctx, key, &out)
    if err == ErrCacheMiss || time.Since(c.GetExpiryTime(key)) > c.DefaultTTL {
        item, err := loader()  // Call database loader function
        if err != nil {
            return fmt.Errorf("failed to reload cache: %w", err)
     }
        
        return c.Set(ctx, key, *item, WithTTL(c.DefaultTTL))
    }
    
    return nil
}
```

**Test Requirements**:
```go
func TestSet_ExpiresAfterTTL(t *testing.T) {
    shortClient := NewGoCacheClient()
    err := shortClient.Set(context.Background(), "key", MenuItem{Name: "Expire Me"})
    assert.NoError(t, err)
    
    // Wait for TTL to expire (5 seconds in test)
    time.Sleep(5 * time.Second)
    
    assert.False(t, shortClient.Has("key")) // Expired key no longer exists
}

func TestGet_RefreshesExpiredEntry(t *testing.T) {
    gocacheClient := gocache.New(config.DefaultTTL, config.DefaultTTL*2)
     gocacheStore := goCacheStore.NewGoCache(gocacheClient)
    cacheManager := cache.New[string, any](gocacheStore)
    
    // Set initial value that will expire
    err := client.Set(context.Background(), "refresh_test", MenuItem{Name: "Old"})
    assert.NoError(t, err)
    
    time.Sleep(100 * time.Millisecond)
    
    var out MenuItem
    err = client.Get(context.Background(), "refresh_test", &out)
    // Should trigger refresh from database loader
    
    assert.ErrorIs(t, err, ErrCacheMissAfterRefresh) // Verify DB reload happened
}
```

---

#### Phase 1d: Entity-Specific TTL Configuration (T008)

**Task T008**: Implement per-relation TTL override mechanism
```go
type PerRelationConfig struct {
    TTLOverrides map[string]time.Duration // "menu.food_details": 30m, "order.items": 15m
}

func (c *GoCacheClient) SetWithTTL(
    ctx context.Context,
    key string,
    value any,
    ttl time.Duration,      // Override per-relation from config
) error {
    var opts []CacheOption[string]
    if cfg := c.GetConfig(); cfg != nil {
        if overriddenTTL, ok := cfg.TTLOverrides[key]; ok {
            opts = append(opts, WithTTL(overriddenTTL))
        }
    }
    
    return c.Set(ctx, key, value, opts...)
}
```

---

#### Phase 1e: Concurrent Access Patterns (T009)

**Task T009**: Implement thread-safe concurrent access validation
```go
// eko/gocache's memory store is already sync.Map-based and thread-safe
// No additional locking needed - the store handles concurrent access

func TestConcurrentAccess_PassUnderLoad(t *testing.T) {
    gocacheClient := gocache.New(config.DefaultTTL, config.DefaultTTL*2)
     gocacheStore := goCacheStore.NewGoCache(gocacheClient)
    cacheManager := cache.New[string, any](gocacheStore)
    
    // Populate cache with test data
    var menus []WeeklyMenu
    err := client.Get(context.Background(), "weekly_menus", &menus)
    if err != nil {
        t.Skip("Skipping - need real DB connection for populate")
    }
    
    // Concurrent read operations - should not fail or corrupt data
    var wg sync.WaitGroup
    errors := make(chan error, 20)
    
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            
            var out WeeklyMenu
            err := client.Get(context.Background(), "weekly_menus", &out)
            if err != nil {
                errors <- fmt.Errorf("concurrent read %d failed: %w", idx, err)
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Should have no errors from 20 concurrent reads
    for err := range errors {
        t.Error(err)
    }
}
```

---

#### Phase 1f: Database Integration (T010)

**Task T010**: Implement database loader interface and fallback mechanism
```go
// Loader interface - abstracts DB operations for cache miss handling
type CacheLoader[T any] interface {
    Load(ctx context.Context, key string) (*T, error) // Fetch from DB
}

// Repository pattern integration example:
func (c *GoCacheClient[T]) GetWithFallback(
    ctx context.Context, 
    key string, 
    out *T,
    loader CacheLoader[T],   // Inject dependency
) error {
    var empty T
    
    err := c.Get(ctx, key, &empty)
    if err == nil {
        return nil // Hit - data already in cache
    }
    
    // Miss - load from database using injected loader
    item, err := loader.Load(ctx, key)
    if err != nil {
        return fmt.Errorf("db fallback failed: %w", err)
    }
    
    return c.Set(ctx, key, *item) // Store in cache after loading
}
```

---

## Implementation Artifacts Checklist

### Core Implementation Files (T01-T04)
- [ ] `internal/cache/client.go` - CacheClient interface and GoCacheClient implementation
- [ ] `internal/cache/storage.go` - Per-entity TTL configuration storage
- [ ] `internal/cache/config.go` - Configuration loading from YAML/env vars
- [ ] `internal/cache/types.go` - Type definitions (CacheConfig, CacheOption, etc.)

### Test Files (TDD Order)
- [ ] `internal/cache/client_test.go` - Unit tests for cache operations
- [ ] `internal/cache/config_test.go` - Unit tests for configuration loading
- [ ] `internal/cache/storage_test.go` - TTL override mechanism tests

---

## Validation & Success Criteria

### Before Implementation:
1. ✅ Spec.md exists with complete feature requirements  
2. ✅ Clarifications documented and integrated
3. ✅ Test coverage requirements defined (TDD approach)

### During Implementation:
1. All unit tests passing (`go test ./internal/cache/... -v`)
2. Coverage ≥ 80% for public functions
3. Concurrent access patterns validated under 15+ goroutines
4. Type safety preserved via generics throughout interface

### After Implementation:
1. Cache operations reduce database query count by 90%+ in typical scenarios
2. Response time P95 ≤ 10ms for cached entity retrieval
3. No memory leaks detected under sustained load (5 minute stress test)
4. Thread-safe access validated via `go test -race`

---

## Next Steps

Run `/speckit-tasks` to generate detailed task breakdown with exact file paths and test specifications. This will produce:
- Tasks/01-Setup.md
- Tasks/02-Config.md  
- Tasks/03-ClientInterface.md
- Tasks/04-BasicOperations.md
- ...and additional task files per implementation phase

---

*Plan generated 2026-05-16 for cache_framework feature | Status: Design Complete, Ready for Task Generation*
