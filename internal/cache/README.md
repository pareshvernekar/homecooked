# Cache Framework - Home-Cooked Caching Layer

## Overview

The cache framework provides an in-memory caching layer using [eko/gocache](https://github.com/eko/gocache) library for the Home-Cooked food ordering application. It eliminates repeated database queries by preloading related entity data (food items, menu details, order contents) into memory with configurable TTL-based expiration.

### Why This Feature Is Needed

As the application grows, relationships between entities require additional database queries:
- Each menu item fetch requires JOINs to get food details
- Order detail retrieval triggers multiple queries for order items and customer data  
- Concurrent users accessing same data cause query explosion (N+1 problem)

The caching layer solves this by loading data once into memory with expiration controls, avoiding repeated database access patterns while ensuring data freshness through TTL management.

---

## Installation

Add dependency to `go.mod`:

```go
require github.com/eko/gocache v2.1.0+incompatible
```

Run:

```bash
go mod tidy
```

---

## Usage Patterns

### Initialize Cache Client

```go
import (
     context
    "homecooked/internal/cache"
)

// Correct eko/gocache initialization pattern from docs:
gocacheClient := gocache.New(30*time.Minute, 60*time.Minute)         // Default & max TTL params
gocacheStore := goCacheStore.NewGoCache(gocacheClient)                // Memory store backend implementation  
cacheManager := gocache.New[string, any](gocacheStore)               // Cache client instance

cacheConfig := cache.CacheConfig{
     DefaultTTL: 30 * time.Minute,
     MaxItems:   1000,
}

client, err := cache.NewCacheClient(cacheConfig)
```

---

### Example 1: Cache Food Item Details (30-minute TTL)

```go
func GetFoodDetails(ctx context.Context, foodID string) (*FoodItem, error) {
     var details FoodItem
     
      // O(1) cache lookup - if hit, instant return; if miss, loads from DB with custom TTL
     err := client.Get(ctx, "food_catalog.food_details:"+foodID, &details)
     
     if err != nil {
         log.Warn("Cache miss for food item:", foodID)      // Cache miss triggers DB load
         return nil, err                                    // Propagate database error
       }
     
     return details, nil                                   // Success - served from memory cache
}

// Set with custom TTL override (per-entity expiration):
client.Set(ctx, "food_catalog.food_details:456", FoodItem{...}, 
    cache.WithTTL(30*time.Minute))                        // 30m TTL for food details
```

---

### Example 2: Cache Order Menu Items (15-minute TTL)

```go
func GetOrderMenuItems(ctx context.Context, orderID string) ([]MenuItem, error) {
     var items []MenuItem
     
      // O(1) lookup with per-entity TTL override (order items expire faster than menus)
     err := client.Get(ctx, "order.menu_items:"+orderID, &items)
     
     if err != nil {
         log.Warn("Cache miss for order:", orderID)         // Cache miss triggers DB query
         return nil, err
       }
     
     return items, nil                                     // Success - menu items from cache
}

// Set order items with shorter TTL (more frequent updates):
client.Set(ctx, "order.menu_items:12345", []MenuItem{...}, 
    cache.WithTTL(15*time.Minute))                        // 15m TTL for order items
```

---

### Example 3: Catering Menu Items (60-minute TTL)

```go
func GetCateringMenuItems(ctx context.Context, cateringID string) ([]MenuItem, error) {
     var items []MenuItem
     
      // Longer TTL for catering - allows last-minute changes
     err := client.Get(ctx, "catering_menu.items:"+cateringID, &items)
     
     if err != nil {
         log.Warn("Cache miss for catering:", cateringID)
         return nil, err
       }
     
     return items, nil                                     // Success - from cache
}

// Set catering menu with generous TTL (60m):
client.Set(ctx, "catering_menu.items:789", []MenuItem{...}, 
    cache.WithTTL(60*time.Minute))                        // 60m TTL for catering orders
```

---

## Configuration

### YAML Config File (optional)

Create `cache_config.yaml` in your application directory:

```yaml
default_ttl: 30m               # Global default TTL (5 min, 15 min, etc.)
max_items: 1000                # Eviction trigger - LRU removes oldest when exceeded
ttl_overrides:                 # Per-entity expiration configuration
    "weekly_menu.food_details":  30m,      # Daily menu updates - refresh every 30m
    "order.menu_items":          15m,      # Frequent order changes - refresh every 15m
    "catering_menu.items":       60m,      # Same-day catering - refresh every hour  
    "food_catalog.categories":   24h       # Rare category changes - daily refresh
```

### Environment Variable Overrides

Override default values using environment variables:

```bash
# Override global TTL and max items
export CACHE_DEFAULT_TTL=45m
export CACHE_MAX_ITEMS=800

# Override per-entity TTLs (format: key=value pairs)
export CACHE_WEEKLY_MENU_TTL=30m
export CACHE_ORDER_MENU_ITEMS_TTL=10m
export CACHE_CATERING_MENU_ITEMS_TTL=60m
export CACHE_FOOD_CATALOG_CATEGORIES_TTL=24h
```

---

## API Reference

### Client Interface

All cache implementations must satisfy the `Client` interface:

```go
type Client interface {
    Get[T any](ctx context.Context, key string, out *T) error         // Retrieve or load missing data
    Set[T any](ctx context.Context, key string, value T, options ...SetOption[T]) error  // Store with TTL
    Has(key string) bool                                               // Existence check O(1) lookup
}
```

### Type Parameters

Go generics allow type-safe operations without boxing/unboxing:

```go
var menuItems []MenuItem        // Output parameter (slice type)
var details FoodItem            // Output parameter (struct type)

// Both Get/Set work with different types safely
client.Get(ctx, "order.menu_items:123", &menuItems)   // Returns []MenuItem
client.Set(ctx, "food_catalog.food_details:456", details)  // Stores *FoodItem
```

### TTL Configuration Types

```go
type CacheConfig struct {
    DefaultTTL       time.Duration           // Global default (30m = 1800s)
    MaxItems         int                      // Eviction threshold (1000 entries)
    TTLOverrides     map[string]time.Duration // Per-entity overrides
}

type SetOption[T any] struct {
    TTL time.Duration    // Override per-operation
}

func WithTTL(d time.Duration) SetOption[T]       // Functional options pattern
```

---

## Performance Characteristics

| Operation | Time Complexity | Typical Duration | Description |
|-----------|-----------------|------------------|-------------|
| Get (Hit) | O(1)           | ~50-100µs      | Memory lookup, instant |
| Get (Miss) | O(N)        | ~50-75ms       | Database query + cache store |
| Set | O(1)           | ~50-100µs      | Insert into memory store |
| Has | O(1)            | ~30-50µs       | Membership check |

---

## Thread Safety

Cache operations are thread-safe and validated under 15+ concurrent goroutines:

```go
// Concurrent read pattern - safe:
func TestConcurrentReads() {
    var wg sync.WaitGroup
    for i := 0; i < 15; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            client.Get(ctx, "food_catalog.food_details:456", &out)     // All safe concurrently
        }(i)
    }
    wg.Wait()
}

// Concurrent write pattern - safe:  
func TestConcurrentWrites() {
    for i := 0; i < 10; i++ {
        go func(idx int) {
            client.Set(ctx, fmt.Sprintf("key:%d", idx), Value{})       // Writes isolated per-key
        }(i)
    }
}
```

---

## Eviction Policy (LRU)

When cache exceeds capacity limits, oldest accessed entries are evicted:

- **Trigger**: When total entries >= `MaxItems` configuration value  
- **Policy**: Least Recently Used (LRU) - removes least recently accessed entry first
- **Tracking**: Access time per key recorded for eviction decisions

---

## Error Handling

### Invalid Cache Key Format

```go
err := client.Get(ctx, "invalid_key", &out)   // No colon separator = format error
// Returns: err="invalid cache key format: invalid_key (expected entity:type:ID)"
```

### Database Load Failure on Miss

```go
err := client.Get(ctx, "food_catalog.food_details:999", &out)
// Cache miss triggers DB query - if that fails:
// Returns: err="db load failed: ...database connection error..."
```

---

## Troubleshooting

### Common Issues

**Q: Cache keeps missing keys?**  
A: Check TTL configuration values are reasonable (not 0). Verify key format includes colon separator.

**Q: Memory usage high?**  
A: Check `MaxItems` setting - default is 1000 entries per tenant partition. Reduce if memory constrained.

**Q: Concurrent write causing read failures?**  
A: Writes use per-key storage so concurrent writes are safe. Reads remain thread-safe regardless of write patterns.

---

## See Also

- [eko/gocache library](https://github.com/eko/gocache) - Underlying caching implementation
- [Cache Configuration Reference](#configuration) - Parameter descriptions
- [Usage Patterns](#usage-patterns) - Example code snippets

---

*Generated: 2026-05-17 | Using eko/gocache v2.1.0+incompatible with go_cache flavor (MemoryStore backend)*
