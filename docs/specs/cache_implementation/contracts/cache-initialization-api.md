# Cache Initialization API Contract

**Feature**: cache-initialization-from-food-categories  
**Version**: 1.0.0  
**Branch**: `002-cache-initialization-from-food-categories`  
**Date**: 2026-05-20

---

## Overview

This document defines the contract for the cache initialization hook pattern used to populate in-memory caches with database-driven reference data at application startup. The contract specifies how the `PostInitialize` method extends the `Client` interface to enable one-time bootstrap operations that query the database and seed the cache.

---

## Interface Contract

### Extended Client Interface

The original `Client` interface is extended with a new `PostInitialize` method for post-creation initialization:

```go
type Client interface {
     // Standard cache operations (unchanged)
    Get(ctx context.Context, key string) (any, error)
    Set(ctx context.Context, key string, value any, options ...SetOption) error
    Has(key string) bool
    
     // NEW: Post-initialization hook for database-driven population
    PostInitialize(ctx context.Context, tenantID string, db *sqlx.DB) error
}
```

---

## Method Signatures

### PostInitialize (NEW METHOD)

```go
func (c *cacheImpl) PostInitialize(ctx context.Context, tenantID string, db *sqlx.DB) error
```

#### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| ctx | `context.Context` | Propagation context for cancellation and timeouts |
| tenantID | `string` | Tenant identifier for scoped database queries (e.g., "tenant_1") |
| db | `*sqlx.DB` | Database connection pool for querying reference tables |

#### Return Value

| Type | Description |
|------|-------------|
| `error` | Returns error if database query fails or cache population encounters an issue. Non-fatal errors are logged but do not prevent application startup. |

---

## Data Models

### FoodCategory (Source Model)

The cache is populated from the `food_category` table using this data model:

```go
type FoodCategory struct {
    ID           string        // Primary key (e.g., "fc_1")
    TenantID     string        // Tenant identifier for RLS (e.g., "tenant_1")
    Name         string        // Category name (e.g., "vegetarian", "dairy-free")
    Description   *string       // Optional description
    CreatedAt    time.Time     // Record creation timestamp
    UpdatedAt    time.Time     // Last update timestamp
}
```

### Cache Key Format: `entity_type:name`

Cache keys follow the pattern `{entity_type}:{name}` where:

- **entity_type**: Normalized domain identifier using kebab-case (e.g., "food-category")
- **:**: Literal colon separator
- **name**: The actual category name in lowercase with spaces converted to hyphens (e.g., "dairy-free", "gluten-free")

#### Examples

| Database Record | Cache Key Format | Resulting Key |
|----------------|------------------|---------------|
| tenant_id="tenant_1", name="vegetarian" | food-category:vegetarian | `food-category:vegetarian` |
| tenant_id="tenant_1", name="dairy-free" | food-category:dairy-free | `food-category:dairy-free` |
| tenant_id="tenant_2", name="gluten-free" | food-category:gluten-free | `food-category:gluten-free` |

---

## Implementation Contract

### Query Pattern

The implementation must execute the following SQL query with parameterized binding:

```sql
SELECT id, tenant_id, name, description, created_at, updated_at
FROM food_category
WHERE tenant_id = $1;
```

**Parameter Binding**: `$1` = tenantID string value

### Cache Population Pattern

For each row returned by the query, call `cache.Set`:

```go
for _, cat := range categories {
     // Construct cache key using entity_type:name format
    key := fmt.Sprintf("food-category:%s", cat.Name)
    
     // Store as pointer to FoodCategory (allows nil Description support)
    value := &cat
    
     // Apply TTL from config override or default
    ttl := c.config.TTLOverrides["food_category"]
    if ttl == 0 {
        ttl = c.config.DefaultTTL
     }
    
    err := c.store.Set(ctx, key, value, store.WithExpiration(ttl))
    if err != nil {
        return fmt.Errorf("failed to cache category %s: %w", cat.Name, err)
     }
}

// Log completion for observability
log.Printf("Cache initialized with %d food categories for tenant %s", len(categories), tenantID)
```

### Logging Requirement

The implementation MUST log completion with item count using logger passed via dependency injection.

---

## Error Handling Contract

### Non-Fatal Errors (Logged but Don't Block Startup)

| Condition | Behavior |
|-----------|----------|
| Database query fails | Log warning, return error, allow app to start |
| Cache store Set fails for one entry | Log error, continue with remaining entries if possible |
| Tenant has no categories in DB | Log informational message, return nil (cache stays empty for this tenant) |

### Fatal Errors (Log and Return - Should Never Occur)

| Condition | Behavior |
|-----------|----------|
| Invalid database connection passed | Log fatal error, fail initialization |
| nil sqlx.DB provided | Fail with clear panic-recover or error |
| Context cancelled during population | Return context cancellation error |

---

## Usage Pattern

### Standard Cache Client Creation (Unchanged)

```go
cacheClient, err := cache.NewCacheClient(config.CacheConfig)
if err != nil {
    log.Fatal("failed to initialize cache: " + err.Error())
}
```

### Post-Initialization Call (NEW)

After creating the cache client and before server setup, call `PostInitialize`:

```go
// Initialize cache with food categories from DB (new hook)
if err := cacheClient.PostInitialize(ctx, tenantID, db); err != nil {
    log.Printf("warning: cache initialization failed for some entities: %v", err)
    // Non-fatal warning allows app to start; missing keys will cause cache misses
}

// Proceed with server setup (cache already warmed up)
s := &server.Server{
    Router: router,
    DB:     db,
    Logger: logger,
    Cache:  cacheClient, // Already contains cached categories
}
```

---

## TTL Configuration Contract

### Per-Entity TTL Override Key Format

Per-entity TTL overrides use the same `entity_type:name` key convention:

```yaml
cache:
  enabled: true
  global_ttl: "30m"
  max_items_per_tenant: 1000
  ttls:
    food_category: "30m"   # NEW ENTRY - all food categories expire after 30 min
    menu.weekly: "30m"
    order.menu_items: "15m"
```

The implementation must retrieve the TTL using:

```go
ttl := c.config.TTLOverrides["food_category"]
if ttl == 0 {
    ttl = c.config.DefaultTTL // Fallback to global default (30m)
}
```

---

## Thread Safety Contract

### Locking Requirements

The `PostInitialize` method must be thread-safe and follow the existing locking pattern:

```go
func (c *cacheImpl) PostInitialize(ctx context.Context, tenantID string, db *sqlx.DB) error {
     // Acquire write lock before cache population
    c.mu.Lock()
    defer c.mu.Unlock()
    
     // Query database
    var categories []models.FoodCategory
    err := db.Get(&categories, `SELECT id, tenant_id, name, description, created_at, updated_at 
                                  FROM food_category WHERE tenant_id = $1`, tenantID)
    if err != nil {
        return err
     }
    
     // Populate all entries while holding lock
    for _, cat := range categories {
        key := fmt.Sprintf("food-category:%s", cat.Name)
        value := &cat
        
        ttl := c.config.TTLOverrides["food_category"]
        if ttl == 0 {
            ttl = c.config.DefaultTTL
         }
        
        err = c.store.Set(ctx, key, value, store.WithExpiration(ttl))
        if err != nil {
            return err
         }
     }
    
     // Log completion
    log.Printf("Cache initialized with %d food categories", len(categories))
    return nil
}
```

---

## Testing Contract

### Unit Test Requirements

The implementation MUST include unit tests covering:

1. **Successful Cache Population**
    - Mock database returning N category records
    - Verify exactly N calls to `cache.Set` with correct `entity_type:name` keys
    - Verify TTL applied correctly from config

2. **Error Handling - Database Failures**
    - Mock query failure returning error
    - Verify error returned (non-fatal) allows application startup

3. **Tenant Isolation Verification**
    - Create Category A for tenant_1, Category B for tenant_2
    - Verify each cache instance contains only its own categories
    - Verify SQL queries show `WHERE tenant_id = $1` parameterization

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-20 | Initial contract definition for cache initialization hook pattern |

---

**End of Contract Document**
