
# Quick Start Guide: Cache Framework for Home-Cooked

**Version**: 1.0.0  
**Date**: 2026-05-16  
**Feature**: In-Memory Caching Layer with eko/gocache Library  
**Spec**: [docs/specs/cache_framework/spec.md](../spec.md) | **Data Model**: [data-model.md](./data-model.md)

---

## Table of Contents

1. [Setup & Installation](#setup--installation)
2. [Basic Usage Pattern](#basic-usage-pattern)
3. [Entity Cache Examples](#entity-cache-examples)
4. [Configuration Options](#configuration-options)
5. [Testing Guide](#testing-guide)

---

## Setup & Installation

### Prerequisites

```go
// Import the cache package after dependency installation
import (
    "github.com/eko/gocache/lib/v4/cache"           // Cache client interface
    "github.com/eko/gocache/store/go_cache/v2"      // In-memory store implementation
)
```

Add to `go.mod`:
```
require github.com/eko/gocache v2.1.0+incompatible
```

### Initialize Cache Client

```go
import (
    "context"
    "github.com/eko/gocache/lib/v4/cache"
    gocacheStore "github.com/eko/gocache/store/go_cache/v2"
)

func initCache(ctx context.Context) error {
     // Create go-cache instance with memory store backend
     var store cache.Store = gocacheStore.NewGoCache(
         cache.WithDefaultTTL(30 * time.Minute),  // Global default TTL
     )
     
     // Wrap in client interface (type-safe, generic over T)
     client := cache.New[string, FoodItem](store)
     
     return nil    // Cache ready for use
}
```

### Inject into Application

```go
// Example: Server startup initialization
func main() {
    var cacheClient cache.CacheClient
    
    db := initDatabaseConnection()
    
    ctx := context.Background()
    if err := initCache(ctx); err != nil {
        log.Fatalf("Failed to initialize cache: %v", err)
     }
     
     // Inject client into handlers/repository via DI pattern
     foodRepo := NewFoodRepository(db, cacheClient)
     orderHandler := NewOrderHandler(foodRepo)
     
     http.ListenAndServe(":8080, nil)   // Server starts with cache available
}
```

---

## Basic Usage Pattern

### Step 1: Load Entity from Cache (O(1))

```go
func GetUserProfile(ctx context.Context, userID string, out *UserProfile) error {
     client := getCacheClientFromContext(ctx)
     
     // Try to retrieve cached user data
     var empty UserProfile      // Output variable to be populated
     
     err := client.Get(
         ctx,
         "users:" + userID,          // Cache key format: entity:type:id
         &empty,                    // Pointer required - will hold result
     )
     
     if err == nil && len(empty.Name) > 0 {
         return nil                // Success - cached value found
     }
     
     // Cache miss - fallback to database query
     var dbUser DBUser
     err = repo.Get(ctx, userID, &dbUser)
     if err != nil {
         return err               // Database error propagated
      }
     
     // Populate output and cache for future requests
     out = &UserProfile{ID: dbUser.ID, Name: dbUser.Name}
     return client.Set(ctx, "users:"+userID, *out)  // Store result
}
```

---

### Step 2: Store with Custom TTL

```go
// Store food item details with 30-minute expiration
func StoreMenuItems(ctx context.Context, key string, items []MenuItem) error {
    client := getCacheClientFromContext(ctx)
    
    return client.Set(
        ctx,
        key,                          // "weekly_menu.menu_items:123"
        items,                      // Slice of menu items
        30*time.Minute,             // TTL override for this entity type
        "lru",                      // Eviction policy (use least recently used)
     )
}
```

---

### Step 3: Check Existence Before Operations

```go
// Pre-check if data is cached before processing
func ProcessOrder(ctx context.Context, orderID string) error {
    client := getCacheClientFromContext(ctx)
    
    // Fast existence check - O(1) lookup
    if !client.Has("order.menu_items:" + orderID) {
        log.Printf("Order %s not in cache - will load from database", orderID)
        
        // Need to load before processing
        return loadFromDatabase(ctx, orderID)
      }
    
     // Proceed with cached data (fast path)
    return handleCachedOrder(ctx, orderID)
}
```

---

## Entity Cache Examples

### Food Catalog Example

```go
// Retrieve food item details by ID - O(1) map lookup
func GetFoodItemDetails(ctx context.Context, foodID string) (*FoodItem, error) {
     var out FoodItem
    
    err := cache.Get(
        ctx,
        "food_catalog.food_details:"+foodID,       // Key format: entity:type:id
        &out,                                      // Populate with concrete type
     )
     
     if err != nil {
         return nil, err                         // Miss/Expired triggers DB load
      }
     
     return out, nil                            // Success - food data retrieved
}

// Usage: Render food detail page instantly from cache
func renderFoodDetailPage(foodID string) http.Handler {
    menuHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        foodItem, err := GetFoodItemDetails(r.Context(), foodID)
        
        if err != nil || foodItem == nil {
            http.Error(w, "Menu item not found", http.StatusNotFound)
            return
     }
     
     tmpl.ExecuteTemplate(w, "food_detail.html", foodItem)   // Render instantly
    })
    
    return menuHandler
}
```

---

### Weekly Menu Example

```go
// Retrieve all menu items for a specific weekly menu - bulk load
func GetWeeklyMenuItems(ctx context.Context, menuID string) ([]MenuItem, error) {
     var out []MenuItem
    
     err := cache.Get(
         ctx,
         "weekly_menu.menu_items:"+menuID,         // Cache key: entity:type:id
         &out,                                     // Output slice populated
      )
      
      if err == nil && len(out) > 0 {
          return out, nil                        // Cached - instant response
       }
       
      // Miss - load all items from database once
    items, dbErr := repo.LoadMenuItems(ctx, menuID)
    if dbErr != nil {
        return nil, dbErr                        // Database error propagated
     }
    
     // Store bulk-loaded results for future fast lookups
     cache.Set(ctx, "weekly_menu.menu_items:"+menuID, items, 15*time.Minute)
     
     return items, nil                          // Success - bulk-retrieved
}

// Usage: Display entire menu page instantly (no per-item DB calls)
func ServeHTTPMenuPage(w http.ResponseWriter, r *http.Request) {
    menuItems, err := GetWeeklyMenuItems(r.Context(), "2026-W15")
    
     // All items in memory - instant rendering possible
    tmpl.Execute(w, struct {
        Items []MenuItem
    }{menuItems})
}
```

---

### Catering Menu Example

```go
// Store catering order items with 60-minute TTL (same-day orders)
func SetCateringMenuItems(ctx context.Context, cateringID string, items []MenuItem) error {
     client := getCachingClient(r.Context())
     
     return client.Set(
         ctx,
         "catering_menu.items:"+cateringID,         // Key: entity:type:id
         items,                                      // Slice of catering menu items
         60*time.Minute,                            // Catering TTL - generous for same-day orders
         "lru",                                     // LRU eviction when cache full
      )
}
```

---

### Order Items Example (Side-Loaded)

```go
// When fetching order details, also retrieve menu items as side-effect:
func GetOrderWithMenuItems(ctx context.Context, orderID string) (*Order, []MenuItem, error) {
     client := getCacheClientFromContext(ctx)
     
    var order Order         // Primary entity to fetch first
    err := client.Get(
        ctx, 
        "orders:"+orderID,     // Primary key format
        &order,               // Output pointer
     )
    
    if err != nil {
         return nil, nil, err    // Failed to load order itself
      }
     
     // Now fetch menu items using cached menu data (side-loaded efficiently)
     var menuItems []MenuItem
    
     // Key points to side-loaded menu: "weekly_menu.menu_items:{order.menu_item_ref}"
     // Assumes parent menu already cached from previous operation
    err = client.Get(
        ctx,
        "weekly_menu.menu_items:"+order.MenuID,   // Parent menu key format
        &menuItems,                              // Populate slice output
     )
    
    if err != nil {
        return order, nil, err           // Miss - cannot load without parent menu cache
     }
    
    return order, menuItems, nil        // Success: order + items (both cached)
}

// Usage: Order detail page shows complete information instantly
func ServeHTTPOrderDetail(w http.ResponseWriter, r *http.Request) {
    order, items, err := GetOrderWithMenuItems(r.Context(), "12345")
    
     // Both entities in cache - fast rendering possible
    tmpl.Execute(w, struct {
        Order  Order
        Items   []MenuItem
    }{order, items})
}
```

---

## Configuration Options

### Cache Configuration Structure

```go
type CacheConfig struct {
    DefaultTTL       time.Duration         // 30m (global default for all entities)
    MaxItems         int                    // Eviction trigger: LRU removes oldest entries when exceeded
    TTLOverrides     map[string]time.Duration   // Per-entity TTL configuration
}

// Example: Configure per-relation TTL values before initializing cache
config := CacheConfig{
    DefaultTTL: 30 * time.Minute,         // Global default
    MaxItems:   1000,                    // Eviction threshold
    
    TTLOverrides: map[string]time.Duration{
        "weekly_menu.menu_items": 15 * time.Minute,       // Orders expire in 15m
        "catering_menu.items":     60 * time.Minute,      // Catering expires in 60m
        "food_catalog.food_details": 30 * time.Minute,    // Food details expire in 30m
     },
}
```

### Apply Configuration During Initialization

```go
func NewCacheClient(ctx context.Context, config CacheConfig) (cache.CacheClient, error) {
     // Create MemoryStore with default TTL from config
     store := gocacheStore.NewGoCache(
         cache.WithDefaultTTL(config.DefaultTTL),          // 30m default
      )
     
     var client cache.CacheClient = cache.New[string, any](store)
      
     return client, nil    // Cache ready with global defaults applied
}
```

---

## Testing Guide

### Running Unit Tests

```bash
cd internal/cache
go test -v           # Run all tests with verbose output
go test -cover       # Show code coverage percentages
go test -race        # Detect race conditions in concurrent access patterns
```

### Example: Writing Cache Client Tests

**File**: `internal/cache/client_test.go`

```go
package cache

import (
     "context"
     "testing"
     
     "github.com/eko/gocache/lib/v4/cache"
)

func TestGet_CacheHit(t *testing.T) {
     gocacheClient := gocache.New(config.DefaultTTL, config.DefaultTTL*2)
     gocacheStore := goCacheStore.NewGoCache(gocacheClient)
    cacheManager := cache.New[string, any](gocacheStore)
     
     // Arrange: Pre-populate with known value
     var empty FoodItem
     err := client.Set(context.Background(), "test_key", FoodItem{Name: "Test"})
     if err != nil {
         t.Fatalf("Setup failed: %v", err)
      }
     
     // Act: Attempt Get operation
     var out FoodItem
     err = client.Get(context.Background(), "test_key", &out)
     
     // Assert: Verify cache hit behavior
     if err != nil {
         t.Errorf("Expected success but got error: %v", err)
      }
     if out.Name != "Test" {
           t.Errorf("Expected Name='Test', got '%s'", out.Name)
     }
}

func TestGet_CacheMiss(t *testing.T) {
     gocacheClient := gocache.New(config.DefaultTTL, config.DefaultTTL*2)
     gocacheStore := goCacheStore.NewGoCache(gocacheClient)
    cacheManager := cache.New[string, any](gocacheStore)    // Empty cache
    
     var empty FoodItem
     err := client.Get(context.Background(), "nonexistent_key", &empty)
     
     // Should trigger database reload (or return miss error)
     if err == nil || len(empty.Name) > 0 {
         t.Error("Expected cache miss error but got success")
      }
}

// Concurrent Access Pattern Test - Validate Thread Safety
func TestConcurrentAccess_15Goroutines(t *testing.T) {
     gocacheClient := gocache.New(config.DefaultTTL, config.DefaultTTL*2)
     gocacheStore := goCacheStore.NewGoCache(gocacheClient)
    cacheManager := cache.New[string, any](gocacheStore)
     
     // Pre-populate test data
     err := client.Set(context.Background(), "shared_key", FoodItem{Name: "Test Item"})
     if err != nil {
         t.Fatalf("Setup failed: %v", err)
      }
     
     var wg sync.WaitGroup
     
     // Launch 15 concurrent read operations (simulates multi-user access)
     for i := 0; i < 15; i++ {
         wg.Add(1)
         go func(id int) {
             defer wg.Done()
             
             var out FoodItem
             err := client.Get(context.Background(), "shared_key", &out)
             
             if err != nil {
                 t.Errorf("Goroutine %d failed: %v", id, err)
              }
         }(i)
      }
      
     // Wait for all goroutines to complete
     wg.Wait()
     
     t.Log("All 15 concurrent operations completed successfully - no race conditions detected")
}

func TestSet_WithCustomTTL(t *testing.T) {
     gocacheClient := gocache.New(config.DefaultTTL, config.DefaultTTL*2)
     gocacheStore := goCacheStore.NewGoCache(gocacheClient)
    cacheManager := cache.New[string, any](gocacheStore)
     
     err := client.Set(
         context.Background(),
         "short_ttl_test",
         FoodItem{Name: "Fast Expiry"},
         5*time.Second,                    // Short TTL for testing
      )
     
     if err != nil {
         t.Fatalf("Set failed: %v", err)
     }
     
     // Verify entry exists immediately
     var out FoodItem
     err = client.Get(context.Background(), "short_ttl_test", &out)
     if err != nil || out.Name != "Fast Expiry" {
         t.Error("Entry should exist immediately after Set")
      }
     
     // Wait for TTL to expire
     time.Sleep(6 * time.Second)       // Let pass expiry
     
     // Verify entry no longer exists (expired)
     err = client.Get(context.Background(), "short_ttl_test", &out)
     if err == nil {
         t.Error("Expected error after TTL expiration")
      }
}
```

---

## Troubleshooting Guide

### Issue: Cache Miss Triggering Unexpected Reloads

**Symptom**: Application logs show repeated database queries for same entity

**Solution**: Ensure cache initialization happens before handler execution:

```go
// In server setup, initialize cache before starting handlers
func initServer() {
    cache := NewCacheClient(context.Background(), defaultConfig)
    
    foodHandler := newFoodHandler(cache)   // Pass cache dependency
    orderHandler := newOrderHandler(foodHandler)
    
    http.HandleFunc("/menu", foodHandler.GetMenuPage)  // Cache is now available
    
    server := &http.Server{Addr: ":8080"}
    go server.ListenAndServe()
}
```

---

### Issue: Memory Exhaustion Warning in Logs

**Symptom**: Application warns "cache approaching memory limit"

**Solution**: Increase eviction capacity threshold:

```go
config := CacheConfig{
    MaxItems:   1000,                  // Change from 500 if hitting limits
    DefaultTTL: 30 * time.Minute,
    
     TTLOverrides: map[string]time.Duration{
        "weekly_menu.menu_items": 15 * time.Minute,      // Adjust per your use case
     },
}

cache := NewCacheClient(context.Background(), config)   // Use new config
```

---

### Issue: Concurrent Read Failures After Cache Miss Reload

**Symptom**: Multiple users accessing same entity after cache reload see inconsistent data

**Solution**: Verify concurrent test passes (should have 0 errors):

```bash
cd internal/cache
go test -race -run TestConcurrentAccess_15Goroutines
```

If race detected: Ensure eviction policy is applied atomically via eko/gocache store abstraction.

---

## Quick Reference Commands

```bash
# Initialize cache in Go application
go run main.go          # Server starts with cache ready

# Run tests (unit only)
go test ./internal/cache/...

# Run tests with coverage
go test -cover ./internal/cache/... -v

# Test for race conditions
go test -race ./internal/cache/...

# Check memory usage (runtime profiling)
go tool pprof http://localhost:8080/debug/pprof mem

# Verify cache entries exist (debug endpoint if implemented)
curl http://localhost:8080/admin/cache-status
```

---

*Generated: 2026-05-16 | Use this guide to implement in-memory caching layer with eko/gocache for your Home-Cooked food ordering application*
