# Research & Architecture Document: FoodItem-CRUD-Service

**Created**: 2026-05-28  
**Status**: Updated Analysis Complete  
**Purpose**: Document architecture decisions and integration patterns for the FoodItem-CRUD-Service implementation

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Dependency Injection Pattern](#dependency-injection-pattern)
3. [Repository Layer Analysis](#repository-layer-analysis)
4. [Caching Strategy](#caching-strategy)
5. [Service Integration Patterns](#service-integration-patterns)
6. [Cache Key Naming Conventions](#cache-key-naming-conventions)
7. [Error Handling Standards](#error-handling-standards)
8. [Implementation Decisions](#implementation-decisions)

---

## Architecture Overview

### Current Architecture Stack

```
┌─────────────────┐     ┌──────────────┐      ┌───────────────┐
│   Gin Router     │────▶│  Handler      │─────▶│    Service      │
│    (HTTP Layer)   │◀────│ (Business     │─────▶│               │
└─────────────────┘      │   Logic Layer)│      └──────┬───────┘
                         └──────────────┘              │
                                                        ▼
                                              ┌───────────────┐
                                              │ Repository    │
                                              │   (Data       │
                                              │    Access)    │
                                              └──────┬───────┘
                                                       │
                ┌─────────────────────────────────────┼─────────────┐
                │                                     │             │
      ┌─────────▼────────┐              ┌─────────────▼──────────┐   │
      │    Cache         │◀─────────────│  (Invalidate/Update)  │   │
      │  (Redis/GoCache) │               └───────────────────────┘   │
      └──────────────────┘                                            │
                     ▲                                                ▼
                ┌───────────────────┐                    ┌───────────────┐
                │    DB             │◀══════════════════▶│    DB         │
                │ (PostgreSQL/RLS)  │                    │ (PostgreSQL/  │
                └───────────────────┘                    │   RLS)        │
```

### Project Structure

The existing codebase follows a clean architecture with separation of concerns:

- **cmd/** - Application entry point and bootstrap
- **internal/cache/** - Cache layer abstraction (generic interfaces)
- **internal/handlers/** - HTTP request handling and validation
- **internal/logger/** - Structured logging with slog
- **internal/models/** - Domain models and DTOs
- **internal/repository/** - Data access layer (interface + implementation)
- **internal/services/** - Business logic layer
- **internal/views/** - Response view models

---

## Dependency Injection Pattern

### Pattern Analysis

The existing codebase uses explicit constructor injection following the dependency inversion principle:

```go
// From internal/handlers/fooditem.go
type FoodItemHandler struct {
    Repo       repository.FoodItemRepository
    Logger      *logger.Logger
    CacheClient cache.TypedClient[models.FoodItem]
}

// Factory function for constructor injection
func NewFoodItemHandler(
    repo repository.FoodItemRepository,
    l *logger.Logger,
    c cache.TypedClient[models.FoodItem]
) *FoodItemHandler {
    return &FoodItemHandler{
        Repo:       repo,
        Logger:     l,
        CacheClient: c,
    }
}
```

### Service Dependency Injection Pattern (Template from FoodCategoryService)

```go
type FoodCategoryService struct {
    repository   repo.FoodCategoryRepository
    logger       *logger.Logger
    cacheClient  cache.TypedClient[models.FoodCategory]
}

func NewFoodCategoryService(
    repo repo.FoodCategoryRepository,
    l *logger.Logger,
    c cache.TypedClient[models.FoodCategory]
) *FoodCategoryService {
    return &FoodCategoryService{
        repository:   repo,
        logger:       l,
        cacheClient:  c,
    }
}
```

### Decision for FoodItemService

Following the same pattern from `FoodCategoryService`, the `FoodItemService` will:

1. **Inject Dependencies**: Repository, Logger, and Cache Client via constructor
2. **Follow Interface Pattern**: All dependencies are interfaces (not concrete types)
3. **Graceful Degradation**: Check for nil cache client before use

```go
type FoodItemService struct {
    repository   repo.FoodItemRepository
    logger       *logger.Logger
    cacheClient  cache.TypedClient[models.FoodItem]
}

func NewFoodItemService(
    repo repo.FoodItemRepository,
    l *logger.Logger,
    c cache.TypedClient[models.FoodItem]
) *FoodItemService {
    return &FoodItemService{
        repository:   repo,
        logger:       l,
        cacheClient:  c,
    }
}
```

---

## Repository Layer Analysis

### FoodItemRepository Interface (Existing)

```go
// From internal/repository/fooditem_repository.go
type FoodItemRepository interface {
    Create(foodItem *models.FoodItem) error
    GetByID(id string) (*models.FoodItem, error)
    Update(foodItem *models.FoodItem) error
    Delete(id string) error
    ListByTenant(tenantID string, offset, limit int) ([]models.FoodItem, int64, error)
}
```

### PostgreSQLFoodItemRepository Implementation (Existing)

Key observations:
1. **RLS Integration**: Uses `current_setting('app.current_tenant_id')` in queries
2. **sqlx ORM**: Leverages sqlx for type-safe database operations
3. **Parameterized Queries**: Prevents SQL injection via $N placeholders
4. **Time Tracking**: Maintains CreatedAt and UpdatedAt for staleness detection

### FoodCategoryRepository Interface (Existing)

```go
type FoodCategoryRepository interface {
    ListByTenant(tenantID string) ([]models.FoodCategory, error)
}
```

### Gap: Missing GetCategoryByName Method

**Current State**: The existing `FoodCategoryService` does NOT have a `GetCategoryByName` method.

**Requirement from spec**: FoodItemService needs to resolve category names to IDs for database persistence.

**Integration Plan**: Add `GetCategoryByName` method to `FoodCategoryService`:

```go
// GetCategoryByName retrieves a food category by its name (case-insensitive)
func (s *FoodCategoryService) GetCategoryByName(categoryName, tenantID string) (string, error) {
    ctx := context.Background()
    s.logger.Info(ctx, "GetCategoryByName: Resolving category name to ID", 
         "category_name", categoryName, "tenant_id", tenantID)
    
    categories, err := s.repository.ListByTenant(tenantID)
    if err != nil {
        s.logger.Error(ctx, "GetCategoryByName: Failed to retrieve food categories", 
             "error", err)
        return "", err
     }
    
    for _, cat := range categories {
         // Case-insensitive comparison
        if strings.EqualFold(cat.Name, categoryName) {
            s.logger.Info(ctx, "GetCategoryByName: Category resolved", 
                 "category_id", cat.ID, "name", cat.Name)
            
             // Optionally cache the resolution for performance
            if s.cacheClient != nil {
                cacheKey := cat.GetCacheKey()
                 _ = s.cacheClient.Set(ctx, cacheKey, cat)
             }
            return cat.ID, nil
         }
     }
    
    s.logger.Error(ctx, "GetCategoryByName: Category name not found", 
         "category_name", categoryName)
    return "", fmt.Errorf("category '%s' not found in tenant %s's catalog", categoryName, tenantID)
}
```

---

## Caching Strategy

### Cache Layer Interface (Existing - Generic Pattern)

From `internal/cache/client.go`:

```go
// TypedClient[T] interface for type-safe cache operations
type TypedClient[T any] interface {
    Get(ctx context.Context, key string) (T, error)
    Set(ctx context.Context, key string, value T, options ...SetOption) error
    Has(key string) bool
    PostInitialize(ctx context.Context, tenantID string, repos map[string]any) error
}

// Cache Key Format: entity_type:entity_identifier
// Examples:
// - "food-category:vegetarian" (by name)
// - "food-category:c5a1f2e3..."  (by ID)
// - "food-item:a7b3c9d1..."        (by ID)
```

### Cache Key Convention Pattern

The existing codebase uses the pattern `entity_type:identifier`:

| Entity Type | Cache Key Format | Example Keys |
|-------------|------------------|--------------|
| FoodCategory | `food-category:{name}` | `"food-category:vegetarian"`, `"food-category:non-vegetarian"` |
| FoodItem | `food-item:{id}` | `"food-item:a1b2c3d4-e5f6..."` |

### TTL Configuration Strategy

From `internal/cache/config.go`:

```go
type CacheConfig struct {
    DefaultTTL       time.Duration
    MaxItems         int
    TTLOverrides     map[string]time.Duration
}

// Default configuration (from config.yaml)
var defaultTTLOverrides = map[string]time.Duration{
    "weekly_menu.food_details":      30 * time.Minute, // Daily menu updates - refresh every 30m
    "order.menu_items":              15 * time.Minute, // Frequent order changes - refresh every 15m
    "catering_menu.items":           60 * time.Minute, // Same-day catering - refresh every hour
    "food_catalog.categories":       24 * time.Hour,   // Rare category changes - daily refresh
}

// For FoodItem - proposed addition:
"food-item:*": 30 * time.Minute // Default for all items
```

### Cache Operation Patterns (from existing code)

#### Pattern 1: Write-Through (FoodCategoryService.ListCategories)

```go
if s.cacheClient != nil && len(categories) > 0 {
    for _, catItem := range categories {
        _ = s.cacheClient.Set(ctx, catItem.GetCacheKey(), catItem)
    }
}
```

#### Pattern 2: Cache-Aside with Staleness Detection (FoodItemService)

```go
// Read flow: Check cache → Detect staleness → Refresh from DB
func (s *FoodItemService) GetItems(tenantID string) ([]models.FoodItem, error) {
     // Step 1: Try cache
    if s.cacheClient != nil {
        cachedItems, err := s.cacheClient.Get(ctx, "food-item:*")
        if err == nil && !isStale(cachedItems) {
            return cachedItems, nil
         }
     }
    
     // Step 2: Fallback to DB for stale cache
    items, err := s.repository.ListByTenant(tenantID)
    if err != nil {
        return nil, err
     }
    
     // Step 3: Refresh cache with fresh data
    if s.cacheClient != nil {
        for _, item := range items {
             _ = s.cacheClient.Set(ctx, "food-item:"+item.ID, item)
         }
     }
    
    return items, nil
}

// Update/Delete flow: Invalidate affected entries
func (s *FoodItemService) Create(item models.FoodItem) error {
    err := s.repository.Create(&item)
    
     // Invalidate cache entry for the newly created item
    if s.cacheClient != nil {
        key := fmt.Sprintf("food-item:%s", item.ID)
         _ = s.cacheClient.Set(ctx, key, item)
     }
    
    return err
}
```

---

## Service Integration Patterns

### Pattern 1: Name-to-ID Resolution (Category Lookup)

**Use Case**: FoodItemService receives category names from API, but database requires IDs.

**Solution Flow**:
```
┌──────────────┐      ┌──────────────────────┐      ┌───────────────┐      ┌───────────────┐
│ FoodItemReq   │───▶  │ FoodCategoryService   │───▶  │ GetByID       │───▶│ ID for DB       │
│ (Name Only)    │      │ (name → ID resolver)    │      │              │      │ persistence    │
└──────────────┘      └──────────────────────┘      └───────────────┘      └───────────────┘
```

**Implementation**:
1. FoodItemService receives `categoryName` from HTTP request
2. Calls `FoodCategoryService.GetCategoryByName(categoryName, tenantID)`
3. Receives back resolved `categoryID`
4. Persists to database with ID instead of name

### Pattern 2: Cache Invalidation by Operation Type

**Use Case**: Different cache invalidation patterns based on operation type (Create/Read/Update/Delete).

**Cache-Aside Pattern (for Get operations)**:
```go
func (s *FoodItemService) GetByID(id string, tenantID string) (*models.FoodItem, error) {
    ctx := context.Background()
    
    // Step 1: Check cache for existing entry
    if s.cacheClient != nil {
        key := fmt.Sprintf("food-item:%s", id)
        cachedItem, err := s.cacheClient.Get(ctx, key)
        
        if err == nil {
            s.logger.Debug(ctx, "Cache hit for food item", "id", id, "cache_key", key)
            return cachedItem, nil
        }
        
        s.logger.Debug(ctx, "Cache miss - loading from database", "id", id)
    }
    
    // Step 2: Cache miss - load from DB
    item, err := s.repository.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("food item not found with id %s", id)
    }
    
    // Step 3: Populate cache with fresh data (cache-aside)
    if s.cacheClient != nil {
        key := fmt.Sprintf("food-item:%s", id)
        _ = s.cacheClient.Set(ctx, key, item, cache.WithTTL(30*time.Minute))
    }
    
    return item, nil
}
```

**Staleness Detection for Batch Reads (Get All)**:
```go
func (s *FoodItemService) GetByTenant(tenantID string) ([]models.FoodItem, error) {
    ctx := context.Background()
    
    // Step 1: Check cache for existing entries
    if s.cacheClient != nil {
        // Cache key pattern: "food-item:*" for tenant-specific batch retrieval
        // Implementation varies by cache system - may need to check TTL or use metadata
        
        cachedItems, err := s.cacheClient.Get(ctx, "food-item:"+tenantID)
        if err == nil {
            // Check if stale based on UpdatedAt timestamps
            now := time.Now()
            for _, item := range cachedItems {
                if item.UpdatedAt != nil && now.Sub(*item.UpdatedAt) > 5*time.Minute {
                    s.logger.Info(ctx, "Detected stale cache - refreshing from database", 
                        "stale_time", now.Sub(*item.UpdatedAt))
                    break // Found one stale item, need full refresh
                }
            }
            
            if !hasStaleItems(cachedItems) {
                return cachedItems, nil
            }
        }
    }
    
    // Step 2: Database query for fresh data
    items, err := s.repository.ListByTenant(tenantID)
    if err != nil {
        return nil, err
    }
    
    // Step 3: Refresh all cache entries atomically
    if s.cacheClient != nil {
        for _, item := range items {
            key := fmt.Sprintf("food-item:%s", item.ID)
            _ = s.cacheClient.Set(ctx, key, item, cache.WithTTL(30*time.Minute))
        }
        
        // Optionally delete tenant-level wildcard entry if exists
        _ = s.cacheClient.Delete("food-item:"+tenantID)
    }
    
    return items, nil
}
```

**Delete Operation Pattern**:
```go
func (s *FoodItemService) Delete(id string, tenantID string) error {
    ctx := context.Background()
    
    err := s.repository.Delete(id)
    if err != nil {
        return fmt.Errorf("failed to delete food item with id %s", id)
    }
    
    // Step 2: Invalidate cache entry (delete key)
    if s.cacheClient != nil {
        key := fmt.Sprintf("food-item:%s", id)
        _ = s.cacheClient.Delete(key)
        s.logger.Info(ctx, "Food item deleted from database and cache invalidated", 
            "id", id, "cache_key", key)
    }
    
    return nil
}
```

---

## Cache Key Naming Conventions

### FoodCategory Cache Keys

Format: `food-category:{category-name}`

Examples:
- `"food-category:vegetarian"`
- `"food-category:non-vegetarian"`
- `"food-category:dessert"`

Used for:
- Storing category listings in cache
- Retrieving specific category by name (case-insensitive)

### FoodItem Cache Keys

Format: `food-item:{item-id}` or `food-item:{tenantID}:{item-id}`

Examples:
- `"food-item:a1b2c3d4-e5f6..."`
- `"food-item:tenant1:a1b2c3d4-e5f6..."`

Used for:
- Individual item retrieval
- Batch cache invalidation by operation (Create/Update/Delete)

---

## Error Handling Standards

### Standardized Error Responses (From internal/views/errorresponse.go)

```go
type ErrorResponse struct {
    Success   bool          `json:"success"`
    ErrorCode string       `json:"error_code"`
    Message    string        `json:"message"`
    Detail     interface{}   `json:"details,omitempty"`
    Timestamp  time.Time     `json:"timestamp"`
}

// Predefined Error Codes
const (
    SUCCESS                   ErrorCode = "SUCCESS"
    INVALID_REQUEST           ErrorCode = "INVALID_REQUEST"
    VALIDATION_ERROR          ErrorCode = "VALIDATION_ERROR"
    NOT_FOUND                 ErrorCode = "NOT_FOUND"
    FORBIDDEN                 ErrorCode = "FORBIDDEN"
    DATABASE_ERROR            ErrorCode = "DATABASE_ERROR"
    CATEGORY_NOT_FOUND        ErrorCode = "CATEGORY_NOT_FOUND"  // For missing categories
)
```

### Logging Patterns (From internal/logger/logger.go)

```go
// Structured logging with slog
type Logger struct {
    logger *slog.Logger
}

// Usage pattern: Context + Fields
func (l *Logger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
    l.logger.Log(ctx, slog.LevelInfo, msg, keysAndValues...)
}

// Examples:
l.Info(ctx, "CreateFoodItem: Starting flow", 
     "tenant_id", tenantID,
     "category_name", categoryName)

l.Debug(ctx, "Cache hit/miss detection", 
     "hit_miss", "hit",
     "stale_threshold_ms", 300000)

l.Error(ctx, "Database operation failed", 
     "error", err,
     "query", query)
```

### Error Handling Strategy for FoodItemService

1. **Validation Errors** → Return 400 with ValidationErrorDetail
2. **Category Not Found** → Return 404 with descriptive message and suggestions
3. **Database Errors** → Return 500 with DATABASE_ERROR code
4. **Cache Errors** → Log error but continue (graceful degradation)

---

## Implementation Decisions

### ID-1: Service Layer Positioning

**Decision**: Place `FoodItemService` in `internal/services/fooditem/service.go` following the same pattern as FoodCategoryService.

**Rationale**: 
- Keeps business logic separate from handlers
- Enables testability through interface abstraction
- Follows established project patterns

### ID-2: Category Resolution Strategy

**Decision**: Implement `GetCategoryByName` method in FoodCategoryService that returns category ID (string) for database persistence.

**Alternative Considered**: Could return struct with ID + Name, but simple string return is cleaner and sufficient for our use case.

```go
// Current approach - returns just ID:
func GetCategoryByName(name, tenantID string) (string, error) { ... }

// Database requires: category_id (VARCHAR primary key)
// API receives: "vegetarian" (category name)
```

### ID-3: Cache Staleness Detection Strategy

**Decision**: Use time-based staleness detection with configurable TTL threshold.

**Rationale**: 
- Simple implementation without cache versioning complexity
- Consistent with existing cache TTL configuration
- Easy to tune based on business requirements

**Implementation Approach**:
```go
// For each item, compare cached UpdatedAt vs current time
const StalenessThreshold = 5 * time.Minute

func isStale(item models.FoodItem) bool {
    if item.UpdatedAt == nil {
        return true // No timestamp means definitely stale
    }
    return time.Since(*item.UpdatedAt) > StalenessThreshold
}
```

### ID-4: Error Messages for Missing Category

**Decision**: Include suggested valid categories in error message.

**Example Response**:
```json
{
    "success": false,
    "error_code": "CATEGORY_NOT_FOUND",
    "message": "Category 'steak' not found in tenant's catalog",
    "details": {
      "suggested_categories": ["vegetarian", "non-vegetarian", "vegans"]
    }
}
```

### ID-5: Cache Invalidation Pattern (Updated)

**Correction**: When FoodItem updates change a category, invalidate the *FoodItem* cache entry - not the FoodCategory cache. The category invalidation should happen at the FoodCategoryService level (only that service modifies categories).

**Implementation**:
```go
func (s *FoodItemService) Update(ctx context.Context, id string, updates models.FoodItemUpdateRequest) error {
    existingItem, err := s.repository.GetByID(id)
    foodItem := applyUpdates(existingItem, updates)
    
    err = s.repository.Update(foodItem)
    if err != nil {
        return err
    }
    
    // Only invalidate the specific FoodItem cache entry
    if s.cacheClient != nil {
        key := fmt.Sprintf("food-item:%s", foodItem.ID)
        _ = s.cacheClient.Delete(key)
    }
    
    return nil
}
```

### ID-6: Service File Organization

**Decision**: Split into separate files for better maintainability.

**Structure**:
```
internal/services/fooditem/
├── service.go          # Core CRUD operations (Get, Create, Update, Delete)
└── service_test.go     # Unit tests
internal/services/foodcategory/
├── service.go          # Categories + GetCategoryByName helper
└── service_test.go     # Category tests
```

**Rationale**: 
- Each domain concept gets its own file
- Service-specific tests alongside implementation
- Easier to navigate and modify independently
- Follows Go idiomatic project structure

---

## Integration Checklist

### Phase 1: Foundation (Ready Now)

- [x] Understand existing patterns from FoodCategoryService
- [x] Analyze repository layer interfaces
- [x] Document cache client interface and conventions
- [ ] Add `GetCategoryByName` to FoodCategoryService
- [ ] Create new FoodItemService with CRUD operations

### Phase 2: Service Implementation (Pending)

- [ ] Implement Create operation with category resolution
- [ ] Implement Read operation with stale detection
- [ ] Implement Update operation with cache invalidation
- [ ] Implement Delete operation with cache removal

### Phase 3: Testing & Integration (Pending)

- [ ] Unit tests for FoodItemService operations
- [ ] Integration tests with mock repositories
- [ ] Cache behavior validation tests
- [ ] Category resolution edge case tests

---

## Files to Create/Modify

### New Files

1. **internal/services/fooditem/service.go** - Main service implementation
2. **internal/services/fooditem/service_test.go** - Unit tests

### Modified Files

1. **internal/services/foodcategory/service.go** - Add `GetCategoryByName` method

---

## References

- **Specification**: `docs/specs/food_menu_system/FoodItem-CRUD-Service.spec.md`
- **FoodCategory Service Template**: `internal/services/foodcategory/service.go`
- **Models**: `internal/models/fooditem.go`, `internal/models/foodcategory.go`
- **Repositories**: `internal/repository/fooditem_repository.go`, `internal/repository/foodcategory_repository.go`
- **Cache Interface**: `internal/cache/client.go`
- **Logger**: `internal/logger/logger.go`
- **Error Views**: `internal/views/errorresponse.go`

---

**Document Status**: Research & Architecture Complete ✅  
**Next Step**: Generate implementation tasks using /speckit-tasks
