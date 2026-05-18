
# Cache Data Model & Entity Relationships

**Date**: 2026-05-16  
**Feature Branch**: 20260516-174330-cache-framework  
**Spec**: [specs/cache_framework/spec.md](../spec.md) | **Design Plan**: [plan/design.md](./design.md)

---

## Entity Cache Key Format Definitions

### Primary Key Structure

All cache keys follow consistent format: `{entity_type}:{entity_id}`

```
Format: <entity_type>:<id>
Example 1: "weekly_menu.food_details:123"    // Menu item 123 with food details
Example 2: "order.menu_items:456"            // Order 456's menu items (slice)
```

### Entity Types Mapped to Cache Keys

| Database Table | Entity Type Name | Cache Key Format | Purpose |
|----------------|------------------|------------------|----------|
| `food_item` | food_details | `food_catalog.food_details:{id}` | Store food item details by ID for menu lookups |
| `menu` | weekly_menu | `weekly_menu.food_details:{id}` | Weekly menu items with associated food data |
| `order_line_items` | order_menu_items | `order.menu_items:{order_id}` | Order's menu items (returns slice) |
| `catering_order` | catering_menu | `catering_menu.items:{id}` | Catering order menu items |

---

## Data Structures

### 1. Cache Client Configuration (`CacheConfig`)

```go
type CacheConfig struct {
    DefaultTTL       time.Duration        // 30m (global default)
    MaxItems         int                   // Eviction threshold when exceeded
    TTLOverrides     map[string]time.Duration  // Per-entity TTL overrides
}
```

**Field Definitions**:

| Field | Type | Default | Description |
|-------|-------|---------|-------------|
| `DefaultTTL` | time.Duration | `30*time.Minute` | Global default expiration for all cache entries |
| `MaxItems` | int | `1000` | Eviction trigger - when exceeded, LRU policy removes oldest entries |
| `TTLOverrides` | map[string]time.Duration | empty (none) | Per-entity TTL configuration: `"order.menu_items"` → `15m`, etc. |

---

### 2. Cache Key Components

#### Entity ID String

```go
type EntityKey struct {
    Type        string  // "weekly_menu.food_details" or "order.menu_items"  
    ID          string   // Actual database ID (string type)
}
// Example: EntityKey{"weekly_menu.food_details", "123"}
```

**Valid Examples**:
```go
key := "weekly_menu.food_details:123"        // food_id 123 for weekly menu
key := "order.menu_items:456"                // order 456's menu items  
key := "catering_menu.items:789"             // catering order 789
```

---

### 3. Cache Entry Payloads (Stored Values)

#### Food Details Entry (`FoodItem`)

When `Get("food_catalog.food_details:123", &food)` succeeds:

```go
type FoodItem struct {
    ID          string   `json:"id"`           // Primary key from database
    Name        string   `json:"name"`         // e.g., "Grilled Salmon"
    Description string   `json:"description"`  // "Fresh Atlantic salmon, lemon-dressed"
    Price       float64  `json:"price"`        // 29.99
    CategoryID  string   `json:"category_id"`  // Reference to food_category table
}

// Cache stores this struct directly:
client.Set(ctx, "food_catalog.food_details:123", FoodItem{...})
```

#### Menu Items Slice (`[]Menu`)

When `Get("weekly_menu.menu_items", &menus)` succeeds (slice retrieval):

```go
type MenuItem struct {
    MenuID      string   // Reference to parent menu
    FoodID      string   // Reference to food_catalog
    Name        string   // Display name: "Pan-Seared Salmon"
    Price       float64  // Unit price
    Quantity    int       // How many ordered
    PreparedAt  time.Time // When served (optional)
}

// Cache stores slice as any type:
client.Set(ctx, "weekly_menu.menu_items", []MenuItem{...})
```

---

## Entity Relationships & Caching Strategy

### One-to-Many Relationships

```
┌─────────────────┐        Many:One        ┌──────────────────┐
│   Menu (1)      │ ◄──────────────►       │  Food Items(n)   │
│ weekly_menu     │                        │ food_catalog     │
└─────────────────┘                        └──────────────────┘

Caching Strategy: Load all menu items once, store in cache
Key Format: "weekly_menu.menu_items:{menu_id}"
Return Type: []Menu  // Slice of menu items for this menu
```

### Many-to-One Relationships

```
┌──────────────┐     One:Many      ┌───────────────┐     Many:One      ┌──────────────┐
│   Order(1)   │ ─────────────►    │ Menu Items(n) │ ◄──────────────►  │   Menu       │
│ order_menu   │                   │ menu_items    │                        │ (parent)   │
└──────────────┘                   └───────────────┘                        └──────────────┘

Caching Strategy: 
- When loading menu items for order, don't load parent menu separately  
- Parent menu already cached from previous operation

Key Format: "order.menu_items:{order_id}"
Return Type: []Menu  // Retrieved as side-effect of other operations
```

---

## TTL Configuration Matrix

| Entity Relation Type | Default TTL | Override TTL | Rationale |
|---------------------|-------------|-------------|-----------|
| `weekly_menu.food_details` | 30 minutes | - | Weekly menus updated daily; 30m allows buffer for late edits |
| `catering_menu.items` | 60 minutes | 60 minutes | Catering same-day orders; generous margin for last-minute changes |
| `order.menu_items` | 15 minutes | 15 minutes | Orders change frequently (cancellations, additions); quick refresh needed |
| `food_catalog.categories` | 24 hours | - | Categories rarely change; daily cache refresh sufficient |

---

## Cache Client API Specification

### Core Operations

#### 1. Get Operation (Single Value)

```go
// Retrieves cached value or triggers reload from DB on miss
// Returns pointer to dereference result (nil if expired/missed)

func Get[T any](
    ctx context.Context,        // Context for cancellation/deadline propagation  
    key string,                // Cache key: "entity_type:id"
    out *T,                   // Pointer to output value (populated on success)
) error {
    // O(1) map lookup - if exists, return cached value
    // If miss/expired, reload from database and store
    
    var empty T
    if cached := memoryStore.Get(ctx, key); cached != nil {
        *out = *cached.(T)       // Type assertion to concrete type
        return nil               // Success - hit in cache
    }
    
    // Cache miss - load from database
    item, err := dbLoader.Load(ctx, key)  // External dependency
    if err != nil {
        return fmt.Errorf("db miss: %w", err)
    }
    
    memoryStore.Set(ctx, key, item)      // Store after loading
    *out = item                          // Populate output
    
    return nil                           // Success - loaded and cached
}
```

**Success Criteria**:
- O(1) complexity for map lookup
- Pointer type returned allows conditional usage: `if val != nil { use(val) }`
- Database fallback automatic on miss (TDD requires test coverage)

---

#### 2. Get Operation (Slice Retrieval)

```go
func Get[T any](
    ctx context.Context,
    key string,
    out *[]T,                 // Slice pointer - nil if miss/empty result
) error {
    var empty []T
    
    if cached := memoryStore.Get(ctx, key); cached != nil {
        slice := cached.([]T)      // Type assertion to slice
        *out = slice               // Populate output slice
        return nil                // Success - slice retrieved from cache
    }
    
    // Miss: load all matching items from database
    items, err := db.QueryAll(ctx, entityType, filterConditions)
    if err != nil {
        return fmt.Errorf("db query failed: %w", err)
    }
    
    memoryStore.Set(ctx, key, items)      // Bulk store all results
    *out = items                          // Return slice
    
    return nil                           // Success - bulk-loaded and cached
}
```

**Example Usage**:
```go
var menuItems []MenuItem          // Output variable (will be populated)
err := cache.Get(ctx, "weekly_menu.menu_items", &menuItems)
if err == nil && len(menuItems) > 0 {
    // Cache hit - display all items instantly
} else {
    // Cache miss - show loading indicator or empty menu
}
```

---

#### 3. Set Operation with TTL

```go
func Set[T any](
    ctx context.Context,
    key string,
    value T,                    // Concrete value to store
    ttl time.Duration,          // Time-to-live (overrides global default)
    policy string,              // Eviction policy ("lru" or "fifo")
) error {
    // Store in memory cache with configured TTL
    
    options := []cache.Option[string]{
        cache.WithTTL(ttl),                  // Override global default
        cache.WithEvictionPolicy(policy),    // LRU: remove least recently used when full
    }
    
    return memoryStore.Set(ctx, key, value, options...)
}
```

**Example**: Storing food item with 30-minute TTL
```go
err := client.Set(
    ctx, 
    "food_catalog.food_details:456",
    FoodItem{ID: "456", Name: "Pasta Primavera", Price: 18.99},
    30*time.Minute,                      // Override default - food items expire in 30m
    "lru",                               // Use LRU eviction when cache full
)
```

---

#### 4. Has Operation (Existence Check)

```go
func Has(key string) bool {
    // O(1) map lookup - check if key exists
    if cached := memoryStore.Get(context.Background(), key); cached != nil {
        return true                    // Key exists and not expired
    }
    return false                      // Miss or expired
}

// Usage pattern:
if cache.Has("order.menu_items:123") {
    fmt.Println("Order data already loaded - can proceed with processing")
} else {
    // Need to load before proceeding
    menuItems, err := db.LoadMenuItemsForOrder(orderID)
    cache.Set(context.Background(), key, menuItems, 15*time.Minute)
}
```

---

#### 5. Flush All Operation (Admin Operations)

```go
func FlushAll() error {
    // Clear entire memory store - bypasses TTL checks
    return memoryStore.Flush(ctx)     // eko/gocache provides this method
}

// Usage: Admin clears cache after menu update
client.FlushAll()
```

---

## Concurrent Access Pattern Guarantees

### Thread-Safety Model

**Implementation Guarantee**: 
```go
var memoryStore gocache.Store = goCacheStore.NewGoCache(
    gocache.WithDefaultTTL(30*time.Minute),
)

// eko/gocache's MemoryStore uses sync.Map internally
// Provides concurrent read safety automatically:
// - Multiple goroutines can safely call Get() simultaneously
// - No additional locking required by application code
```

**Test Coverage Requirement**:
```go
func TestConcurrentAccess_15Goroutines(t *testing.T) {
    var wg sync.WaitGroup
    
    // Populate with test data
    client.Set(context.Background(), "test_key", FoodItem{Name: "Test"})
    
    // Launch 15 concurrent read goroutines
    for i := 0; i < 15; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            
            var out FoodItem
            err := client.Get(context.Background(), "test_key", &out)
            if err != nil {
                t.Errorf("Goroutine %d read failed: %v", idx, err)
                panic("race condition detected")
            }
        }(i)
    }
    
    wg.Wait()  // Ensure all goroutines complete
    
    t.Log("All 15 concurrent operations completed without race conditions")
}
```

---

## Database Integration Model

### Loader Interface Pattern

```go
type CacheLoader[T any] interface {
    Load(ctx context.Context, key string) (*T, error)
}

// Example: Food Item Loader (implements CacheLoader[FoodItem])
type FoodLoader struct {
    DB *sqlx.DB  // Database connection
}

func (l *FoodLoader) Load(ctx context.Context, id string) (*FoodItem, error) {
    var item FoodItem
    
    query := `SELECT id, name, description, price, category_id 
              FROM food_item 
              WHERE id = $1`
    
    err := l.DB.Get(&item, query, id)
    if err != nil {
        return nil, fmt.Errorf("failed to load food item %s: %w", id, err)
    }
    
    return &item, nil  // Return pointer for cache client consumption
}

// Example integration usage:
func (c *GoCacheClient[FoodItem]) Get(ctx context.Context, key string, out *FoodItem) error {
    var empty FoodItem
    
    if !c.Has(key) {
         item, err := foodLoader.Load(ctx, key)  // Injected dependency
         if err != nil {
             return err
         }
         return c.Set(ctx, key, *item)   // Store after loading
    }
    
    return c.GetCached(ctx, key, out)    // Hit in cache
}
```

---

## Summary: Entity Cache Key Index

```
┌───────────────────────────────────────────────────────────────────────────────┐
│                         CACHE KEY INDEX TABLE                                  │
├─────────────────────────┬───────────────────────────────────────────────────────────────────┤
│ Database Table          │ Cache Key Format                                                    │
│ (Database Column)       │                                                                     │
├─────────────────────────┼───────────────────────────────────────────────────────────────────┤
│ food_item.id            │ food_catalog.food_details:id_string                                │
│ menu.id                 │ weekly_menu.menu_items:id_string                                   │
│ catering_order.id       │ catering_menu.items:id_string                                       │
│ order_line_item.order_id│ order.menu_items:order_id_string (side-loaded)                    │
├─────────────────────────┴───────────────────────────────────────────────────────────────┤
│ TTL Configuration per Entity Type (in CacheConfig.TTLOverrides):                │
│ - "food_catalog.food_details": 30 * time.Minute                                   │
│ - "weekly_menu.menu_items": 15 * time.Minute                                      │
│ - "catering_menu.items": 60 * time.Minute                                        │
└───────────────────────────────────────────────────────────────────────────────────┘
```

---

*Generated: 2026-05-16 | Entities: 3 main types (food_catalog, weekly_menu, catering_menu) mapped to cache keys for O(1) retrieval via eko/gocache MemoryStore*
