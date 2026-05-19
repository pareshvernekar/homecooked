# Data Model: Cache Initialization from Food Categories

**Date**: 2026-05-18  
**Feature**: cache-initialization-from-food-categories  
**Branch**: `002-cache-initialization-from-food-categories`

---

## Overview

This document defines the data structures for initializing the in-memory cache with food category records from the `food_category` database table at application startup.

---

## Database Schema Reference (Source of Truth)

From `specs/food_menu_system/database/database_schema.md`:

```sql
CREATE TABLE IF NOT EXISTS food_category (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

---

## Cache Entry Data Model

### `FoodCategory` Structure (from `internal/models/fooditem.go`)

```go
type FoodCategory struct {
    ID           string          // Database primary key (e.g., "fc_1")
    TenantID     string          // Multi-tenancy isolation context
    Name         string          // Category name (e.g., "vegetarian", "dairy-free", "jain")
    Description  *string         // Nullable; may not be populated in some datasets
    CreatedAt    time.Time       // Record creation timestamp
    UpdatedAt    time.Time       // Last update timestamp
}
```

**Note**: Using existing `FoodCategory` model from `internal/models/fooditem.go` rather than creating duplicate struct ensures type consistency across the codebase and reuses validated domain types.

**Fields Mapping**:
| Database Column | Go Field | Cache Key Component | Description |
|----------------|----------|---------------------|-------------|
| `id`            | ID        | -                   | Primary key (stored in cache value, not part of key) |
| `tenant_id`     | TenantID | -                   | Multi-tenancy isolation context (queries scoped by this) |
| `name`          | Name      | cache key           | Human-readable category name used for semantic lookup |
| `description`   | Description | -               | Optional text; null-able in Go representation |
| `created_at`    | CreatedAt | -                  | Timestamp when record was created |
| `updated_at`    | UpdatedAt | -                  | Timestamp of last modification |

---

## Cache Key Convention: `entity_type:name`

### Format Definition

Cache keys follow the pattern: `{entity_type}:{name}` where:
- **entity_type**: Normalized domain entity identifier using kebab-case (e.g., "food-category")
- **:**: Literal colon separator
- **name**: Actual database field value in lowercase with spaces converted to hyphens (e.g., "vegetarian", "dairy-free", "gluten-free", "jain-diet")

### Examples

| Database Record | Cache Key Format | Resulting Key |
|----------------|------------------|---------------|
| tenant_id="tenant_1", name="vegetarian" | food-category:vegetarian | `food-category:vegetarian` |
| tenant_id="tenant_1", name="dairy-free" | food-category:dairy-free | `food-category:dairy-free` |
| tenant_id="tenant_1", name="gluten-free" | food-category:gluten-free | `food-category:gluten-free` |
| tenant_id="tenant_2", name="jain-diet" | food-category:jain-dit | `food-category:jain-diet` |

### Why This Format?

The `entity_type:name` convention (rather than `entity_type:id`) enables:
- **Semantic search**: Look up categories by their descriptive name instead of opaque IDs like "fc_7"
- **Readable logging**: Stack traces show "food-category:dairy-free" instead of "food-category:fc_7"
- **Developer-friendly debugging**: No need to join id back to a database table
- **UI integration**: Easy mapping to frontend filters and dropdowns

### Key Validation Rules

1. Entity type must be lowercase kebab-case (e.g., "food-category", not "FoodCategory")
2. Name component must be:
   - Lowercase
   - Spaces converted to hyphens ("non vegetarian" → "non-vegetarian")
   - Non-alphanumeric characters replaced with hyphens ("Jain Diet" → "jain-diet")
3. Leading/trailing whitespace trimmed from name
4. Empty names after normalization rejected

---

## Sample Data: Tenant 1 Categories (Food Category Model Examples)

```go
categories := []FoodCategory{
     {
        ID:           "fc_1",
        TenantID:     "tenant_1",
        Name:         "vegetarian",
        Description:  str("Vegetarian-based items only"),
        CreatedAt:    time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
        UpdatedAt:    time.Date(2026, 3, 1, 14, 30, 0, 0, time.UTC),
     },
     {
        ID:           "fc_2",
        TenantID:     "tenant_1",
        Name:         "non-vegetarian",
        Description:  str("Meat and seafood items"),
        CreatedAt:    time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
        UpdatedAt:    time.Date(2026, 3, 1, 14, 30, 0, 0, time.UTC),
     },
     {
        ID:           "fc_3",
        TenantID:     "tenant_1",
        Name:         "vegan",
        Description:   nil, // Nullable; may not be populated in some datasets
        CreatedAt:     time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
        UpdatedAt:     time.Date(2026, 3, 1, 14, 30, 0, 0, time.UTC),
     },
     {
        ID:           "fc_4",
        TenantID:     "tenant_1",
        Name:         "gluten-free",
        Description:   str("Items without gluten-containing ingredients"),
        CreatedAt:     time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
        UpdatedAt:     time.Date(2026, 3, 1, 14, 30, 0, 0, time.UTC),
     },
}
```

---

## Database Query Pattern

### SELECT Statement

```sql
SELECT id, tenant_id, name, description, created_at, updated_at
FROM food_category
WHERE tenant_id = $1;
```

**Parameters**: `$1` = tenant_id string (from connection context)

**Result Set**: One row per category for the given tenant

---

## Cache Entry Structure Details

### go-cache Value Format

Stored value type: `*FoodCategory` (pointer to reuse existing model from `internal/models/fooditem.go`)

```go
type CacheEntry struct {
    Key          string         // "food-category:vegetarian" (entity_type:name format)
    Value        interface{}    // *FoodCategory (pointer for nil description support)
    AccessedAt   time.Time      // Timestamp of last cache access (for eviction tracking)
    TTL          time.Duration  // Configurable per entity; e.g., 30 minutes
}
```

### TTL Configuration Pattern

**Per-entity override from config.yaml**:
```yaml
cache:
  ttls:
    food_category: "30m"   # All food category entries expire after 30 min
```

**Applied in PostInitialize**:
```go
ttl := c.config.TTLOverrides["food_category"]
if ttl == 0 {
     ttl = c.config.DefaultTTL // Fallback to global default (30m)
}
cache.Set(ctx, key, value, store.WithExpiration(ttl))
```

---

## Cache Client Interface Extension

### Extended Interface (with PostInitialize):

```go
type Client interface {
    Get(ctx context.Context, key string) (any, error)
    Set(ctx context.Context, key string, value any, options ...SetOption) error
    Has(key string) bool
    PostInitialize(ctx context.Context, db *sqlx.DB) error
}

// Functional option for TTL configuration
type SetOption func(*Option) error

type Option struct {
    TTL time.Duration
}

func WithTTL(d time.Duration) SetOption {
     return func(o *Option) error {
         o.TTL = d
         return nil
      }
}
```

---

## Integration Point: main.go Bootstrap Sequence

**Proposed Location**: After cache client initialization, before HTTP server starts

```go
// 1. Create cache client (existing pattern)
cacheClient, err := cache.NewCacheClient(config.CacheConfig)
if err != nil {
    log.Fatal("failed to initialize cache: " + err.Error())
}

// 2. Initialize cache with food categories from DB (new hook)
if err := cacheClient.PostInitialize(ctx, db); err != nil {
    log.Printf("warning: cache initialization failed for some entities: %v", err)
     // Note: non-fatal warning allows app to start; missing keys will miss
}

// 3. Proceed with server setup (cache already ready)
s := &server.Server{
    Router: router,
    DB:     db,
    Logger: logger,
    Cache:  cacheClient, // Already warmed up
}

// Start server...
```

---

## Testing Strategy

### Unit Test Scenarios

1. **PostInitialize successfully caches N records**
   - Mock sqlx.DB returning expected rows
   - Verify exactly N calls to `cache.Set`
   - Verify keys follow `entity_type:name` pattern (e.g., "food-category:vegetarian")
   - Verify TTL applied correctly

2. **PostInitialize handles database errors gracefully**
   - Mock DB query failure
   - Verify error returned (or logged with warning)
   - Non-fatal behavior allows app startup

3. **Tenant-specific isolation verified**
   - Create Category A for tenant_1, Category B for tenant_2
   - Verify tenant_1's cache contains only A
   - Verify tenant_2's cache contains only B

4. **Cache lookups return cached values after initialization**
   - Simulate PostInitialize completing
   - Call cache.Get with known key
   - Verify immediate return from store (no DB query)

---

## Assumptions

1. Database connection is available at startup (application blocks until DB connects)
2. Cache client implementation wraps eko/gocache store interface
3. Per-tenant cache instance already created before PostInitialize call
4. SQL queries can use prepared statements via sqlx to prevent SQL injection
5. Missing/failed entries are logged as warnings but don't block startup

---

## Dependencies

1. `sqlx` library (already in use for ORM operations)
2. `config.CacheConfig` structure with TTL override support
3. Database connection pool available for queries
4. Main.go bootstrap sequence accessible for hook injection

