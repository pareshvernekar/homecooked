# Quick Start Guide: FoodItem-CRUD-Service Implementation

**Created**: 2026-05-28  
**Feature Branch**: `002-cache-initialization-from-food-categories`  
**Target**: Internal services/fooditem service implementation  
**Specification**: [`FoodItem-CRUD-Service.spec.md`](specs/food_menu_system/FoodItem-CRUD-Service.spec.md)

---

## Table of Contents

1. [Overview](#overview)
2. [Project Structure](#project-structure)
3. [Implementation Tasks](#implementation-tasks)
4. [Code Examples](#code-examples)
5. [Testing Strategy](#testing-strategy)
6. [Integration Checklist](#integration-checklist)

---

## Overview

This document provides a quick start guide for implementing the **FoodItem-CRUD-Service** feature, which adds service-layer business logic for food item management with integrated caching and category resolution.

### What This Feature Delivers

1. **CRUD Operations**: Create, Read (single & batch), Update, Delete
2. **Category Resolution**: Auto-resolve category names to UUIDs using FoodCategoryService
3. **Cache Integration**: Cache-aside pattern with stale detection
4. **Multi-Tenancy**: Full tenant isolation via PostgreSQL RLS

### Key Dependencies

```
FoodItemService requires:
┌─────────────────────────────────────────────┐
│ • FoodItemRepository      (DB access layer)  │
│ • Logger                  (structured logs)   │
│ • Cache Client            (TypedClient[T])    │
│ • FoodCategoryService     (name→ID resolver) │
└─────────────────────────────────────────────┘
```

---

## Project Structure

### New Files to Create

```
internal/services/fooditem/
├── service.go         # Main service implementation
└── service_test.go    # Unit tests
```

### Updated Files

```
internal/services/foodcategory/
└── service.go       # Add GetCategoryByName method
```

---

## Implementation Tasks

### Task 2: Create FoodItemService Structure and Dependency Injection

**File**: `internal/services/fooditem/service.go`

**What to implement**: Service struct with constructor exposing complete CRUD methods that return pointers to full entities

**Acceptance Criteria**:
- [ ] Define `FoodItemService` struct with fields
- [ ] Implement constructor `NewFoodItemService(repo, logger, cache)` returning pointer
- [ ] Inject repository, logger, and cache client as dependencies
- [ ] Expose complete entity methods - all operations work with full pointers:

```go
type FoodItemService struct {
    repository   repo.FoodItemRepository
    logger        *logger.Logger
    cacheClient   cache.TypedClient[models.FoodItem]
}

func NewFoodItemService(
    repo repo.FoodItemRepository,
    l *logger.Logger,
    c cache.TypedClient[models.FoodItem],
) *FoodItemService {
    return &FoodItemService{
        repository:   repo,
        logger:       l,
        cacheClient:  c,
      }
}

// Create returns pointer to created entity with resolved Category ID and timestamps
func (s *FoodItemService) Create(ctx context.Context, tenantID string, item *models.FoodItem) (*models.FoodItem, error) { ... }

// GetByID returns pointer to single entity - complete data never partial
func (s *FoodItemService) GetByID(ctx context.Context, id string) (*models.FoodItem, error) { ... }

// GetByTenant returns slice of pointers to all entities - batch stale detection supported
func (s *FoodItemService) GetByTenant(ctx context.Context, tenantID string) ([]*models.FoodItem, error) { ... }

func (s *FoodItemService) Delete(ctx context.Context, id string) error { ... }
}
```

**Important**: All methods must return pointer types for consistency:
- `Create` → returns `(*models.FoodItem, error)` - not nil or value receiver
- `GetByID` → returns `(*models.FoodItem, error)` - complete entity, never partial data
- `GetByTenant` → returns `([]*models.FoodItem, error)` - slice of pointers to all entities

This ensures callers can safely access all fields and pass references between layers.

---

### Task 1: Add GetCategoryByName to FoodCategoryService

**File**: `internal/services/foodcategory/service.go`

**What to implement**: Method that resolves category name → UUID for use in FoodItem creation/update

**Acceptance Criteria**:
- [ ] Accepts categoryName (string) and tenantID (string) as parameters
- [ ] Lists all categories for tenant using existing repository's ListByTenant() method (respects RLS)
- [ ] Finds matching category using case-insensitive comparison via `strings.EqualFold()`
- [ ] Returns category ID (UUID string) on success - caller uses this for DB persistence
- [ ] Returns descriptive error if category not found
- [ ] Caches the resolved category entity if cache client available (Write-Through pattern)

**Example Pattern**:
```go
func (s *FoodCategoryService) GetCategoryByName(categoryName, tenantID string) (string, error) {
    ctx := context.Background()
    
    // Fetch all categories for this tenant (RLS ensures only tenant's data)
    categories, err := s.repository.ListByTenant(tenantID)
    if err != nil {
        return "", fmt.Errorf("failed to fetch categories: %w", err)
    }
    
    // Iterate and find by name (case-insensitive comparison)
    for _, cat := range categories {
        if strings.EqualFold(cat.Name, categoryName) {
            s.logger.Info(ctx, "GetCategoryByName: Category resolved",
                      "category_name", cat.Name,
                      "resolved_id", cat.ID)
            
            // Cache the category entity (Write-Through pattern - similar to existing ListCategories caching)
            if s.cacheClient != nil {
                cacheErr := s.cacheClient.Set(ctx, cat.GetCacheKey(), cat)
                if cacheErr != nil {
                    s.logger.Error(ctx, "GetCategoryByName: Failed to cache food category",
                              "key", cat.GetCacheKey(), "error", cacheErr)
                    // Don't fail on cache error - keep returning ID
                }
            }
            
            return cat.ID, nil
        }
    }
    
    s.logger.Error(ctx, "GetCategoryByName: Category not found in tenant's catalog",
             "category_name", categoryName, "tenant_id", tenantID)
    return "", fmt.Errorf("category '%s' not found in tenant %s's catalog", categoryName, tenantID)
}
```

---

### Task 2: Create FoodItemService Structure and Dependency Injection

**File**: `internal/services/fooditem/service.go`

**What to implement**: Service struct with constructor exposing complete CRUD methods

**Acceptance Criteria**:
- [ ] Define `FoodItemService` struct with fields
- [ ] Implement constructor `NewFoodItemService(repo, logger, cache)` returning pointer
- [ ] Inject repository, logger, and cache client as dependencies
- [ ] Expose complete entity methods - all operations work with full pointers:

```go
type FoodItemService struct {
    repository   repo.FoodItemRepository
    logger        *logger.Logger
    cacheClient   cache.TypedClient[models.FoodItem]
}

func NewFoodItemService(
    repo repo.FoodItemRepository,
    l *logger.Logger,
    c cache.TypedClient[models.FoodItem],
) *FoodItemService {
    return &FoodItemService{
        repository:   repo,
        logger:       l,
        cacheClient:  c,
    }
}

// Complete entity methods with proper signatures
func (s *FoodItemService) Create(ctx context.Context, tenantID string, item *models.FoodItem) (*models.FoodItem, error) { ... }

func (s *FoodItemService) GetByID(ctx context.Context, id string) (*models.FoodItem, error) { ... }

func (s *FoodItemService) GetByTenant(ctx context.Context, tenantID string) ([]*models.FoodItem, error) { ... }

func (s *FoodItemService) Delete(ctx context.Context, id string) error { ... }
```

---

### Task 3: Implement Create with Category Resolution and Caching

**What to implement**: Full CRUD Create operation with category name resolution → UUID conversion

**Acceptance Criteria**:
- [ ] Accept complete `*models.FoodItem` with category NAME (string) as parameter
- [ ] Call GetCategoryByName() to resolve string name → UUID internally before persisting
- [ ] Update item.Category field with resolved UUID
- [ ] Set CreatedAt timestamp to current time
- [ ] Persist to database using repository's Create() method
- [ ] Populate cache with created entity (Write-Through pattern)
- [ ] **Return the created entity pointer** with resolved category ID and timestamps populated

**Code Example**:
```go
func (s *FoodItemService) Create(
    ctx context.Context,
    tenantID string,
    item *models.FoodItem,
) (*models.FoodItem, error) {
    s.logger.Info(ctx, "CreateFoodItem: Starting creation flow",
        "tenant_id", tenantID,
        "category_name", item.Name, // Note: name as received from client
    )
    
    // Step 1: Resolve category NAME to UUID using FoodCategoryService
    categoryID, err := s.GetCategoryByName(item.Category, tenantID)
    if err != nil {
        return nil, fmt.Errorf("category resolution failed: %w", err)
    }
    
    s.logger.Info(ctx, "CreateFoodItem: Category name resolved to ID",
        "name", item.Category, // String name received from client
        "resolved_id", categoryID, // UUID for DB
    )
    
    // Step 2: Update with resolved ID before persisting
    item.Category = categoryID // Now stores UUID, not string name
    
    now := time.Now().UTC()
    item.CreatedAt = &now
    item.UpdatedAt = &now
    
    s.logger.Debug(ctx, "CreateFoodItem: Prepared item with resolved ID",
        "item_id", item.ID,
        "category_id", item.Category,
        "created_at", item.CreatedAt,
        "updated_at", item.UpdatedAt,
    )
    
    // Step 3: Persist to database (repository handles RLS via tenantID)
    createdItem, err := s.repository.Create(item)
    if err != nil {
        return nil, fmt.Errorf("database operation failed: %w", err)
    }
    
    s.logger.Info(ctx, "CreateFoodItem: Persisted successfully",
        "item_id", createdItem.ID,
        "category_id", createdItem.Category,
    )
    
    // Step 4: Populate cache with the CREATED entity (Write-Through pattern)
    if s.cacheClient != nil {
        cacheKey := fmt.Sprintf("food-item:%s", createdItem.ID)
        setErr := s.cacheClient.Set(ctx, cacheKey, *createdItem, cache.WithTTL(30*time.Minute))
        
        if setErr != nil {
            s.logger.Error(ctx, "CreateFoodItem: Failed to cache food item",
                    "cache_key", cacheKey,
                    "error", setErr)
            // Return success - cache is optional layer
        } else {
            s.logger.Debug(ctx, "CreateFoodItem: Caches populated with created entity",
                    "cache_key", cacheKey,
                    "ttl", 30*time.Minute)
        }
    }
    
    // Step 5: Return the fully persisted entity with resolved Category ID and timestamps
    return createdItem, nil
}
```

---

### Task 4: Implement GetByID with Cache-Aside Pattern and Stale Detection

**What to implement**: Single item retrieval with cache-first approach and staleness detection

**Acceptance Criteria**:
- [ ] Check cache first for existing entry using `TypedClient.Get()` with key format `food-item:{id}`
- [ ] Detect staleness based on UpdatedAt timestamp compared to StalenessThreshold (5min default)
- [ ] If cache miss or stale detected: fall back to repository's GetByID() method
- [ ] After DB retrieval, refresh cache with fresh data atomically
- [ ] Return the entity pointer - complete data never returned in partial form

**Code Example**:
```go
func (s *FoodItemService) GetByID(
    ctx context.Context,
    id string,
) (*models.FoodItem, error) {
    s.logger.Info(ctx, "GetFoodItemByID: Retrieving food item",
        "item_id", id)
    
    // Step 1: Try cache first (cache-aside pattern)
    if s.cacheClient != nil {
        cacheKey := fmt.Sprintf("food-item:%s", id)
        cachedItem, err := s.cacheClient.Get(ctx, cacheKey)
        
        if err == nil {
            s.logger.Debug(ctx, "GetFoodItemByID: Cache hit for food item",
                "item_id", id,
                "cache_key", cacheKey)
            
            // Step 2: Check staleness based on UpdatedAt timestamp
            isStale := s.isStale(&cachedItem)
            
            if !isStale {
                s.logger.Debug(ctx, "GetFoodItemByID: Cache entry fresh - returning from cache",
                    "item_id", id,
                    "stale_threshold_ms", defaultStalenessThreshold.Milliseconds())
                
                // Fresh data - return directly
                return &cachedItem, nil
            } else {
                s.logger.Info(ctx, "GetFoodItemByID: Cache entry stale - will refresh from database",
                    "item_id", id,
                    "stale_threshold_ms", defaultStalenessThreshold.Milliseconds())
                
                err = fmt.Errorf("cache stale") // Treat as miss
            }
        } else {
            s.logger.Debug(ctx, "GetFoodItemByID: Cache miss for food item",
                "item_id", id,
                "error", err)
        }
    }
    
    // Step 3: Fall back to database (cache miss or stale data)
    var foodItem models.FoodItem
    if err := s.repository.GetByID(id, &foodItem); err != nil {
        return nil, fmt.Errorf("database operation failed: %w", err)
    }
    
    s.logger.Info(ctx, "GetFoodItemByID: Retrieved fresh data from database",
        "item_id", foodItem.ID,
        "updated_at", foodItem.UpdatedAt,
    )
    
    // Step 4: Populate cache with fresh data (cache-aside write-back pattern)
    if s.cacheClient != nil {
        cacheKey := fmt.Sprintf("food-item:%s", id)
        setErr := s.cacheClient.Set(ctx, cacheKey, foodItem, cache.WithTTL(30*time.Minute))
        
        if setErr != nil {
            s.logger.Error(ctx, "GetFoodItemByID: Failed to populate cache",
                "item_id", id,
                "cache_key", cacheKey,
                "error", setErr)
        } else {
            s.logger.Debug(ctx, "GetFoodItemByID: Cache populated with fresh data",
                "item_id", id,
                "cache_key", cacheKey)
        }
    }
    
    return &foodItem, nil
}

// Helper method for staleness detection
func (s *FoodItemService) isStale(item *models.FoodItem) bool {
    if item.UpdatedAt == nil {
        return true // No timestamp = definitely stale
    }
    return time.Since(*item.UpdatedAt) > defaultStalenessThreshold
}
```

---

### Task 5: Implement GetByTenant with Batch Stale Detection and Cache Refresh

**What to implement**: Multi-item retrieval for entire tenant's items with batch-level stale detection

**Acceptance Criteria**:
- [ ] Check cache first using key format `food-item:{tenantID}` to fetch all items
- [ ] Iterate through cached items checking each `UpdatedAt` timestamp against StalenessThreshold
- [ ] If ANY single item is stale (or all miss from cache): refresh ENTIRE batch from DB
- [ ] After database refresh, populate all cache entries atomically (loop through results)
- [ ] Return complete list of entity pointers `[]*models.FoodItem` - never return partial/stale data

**Code Example**:
```go
func (s *FoodItemService) GetByTenant(
    ctx context.Context,
    tenantID string,
) ([]*models.FoodItem, error) {
    s.logger.Info(ctx, "GetFoodItemsByTenant: Retrieving all food items for tenant",
        "tenant_id", tenantID)
    
    // Step 1: Try cache first (batch key by tenant)
    if s.cacheClient != nil {
        cacheKey := fmt.Sprintf("food-item:%s", tenantID)
        cachedItems, err := s.cacheClient.Get(ctx, cacheKey)
        
        if err == nil && len(cachedItems) > 0 {
            s.logger.Debug(ctx, "GetFoodItemsByTenant: Cache hit for tenant",
                "tenant_id", tenantID,
                "cache_key", cacheKey,
                "items_count", len(cachedItems))
            
            // Step 2: Iterate through cached items to detect stale entries
            now := time.Now()
            var staleIndices []int
            
            for i := range cachedItems {
                isStale := s.isStale(&cachedItems[i])
                
                if isStale {
                    staleIndices = append(staleIndices, i)
                    s.logger.Info(ctx, "GetFoodItemsByTenant: Found stale item",
                        "tenant_id", tenantID,
                        "stale_index", i,
                        "item_id", cachedItems[i].ID)
                }
            }
            
            // If ANY items are stale, refresh entire batch
            if len(staleIndices) > 0 {
                s.logger.Info(ctx, "GetFoodItemsByTenant: Stale entries detected - refreshing entire batch from database",
                    "tenant_id", tenantID,
                    "stale_count", len(staleIndices),
                    "total_cached", len(cachedItems))
                
                err = fmt.Errorf("batch stale detected") // Treat as full miss
            } else {
                s.logger.Debug(ctx, "GetFoodItemsByTenant: All cached items are fresh - returning from cache",
                    "tenant_id", tenantID,
                    "items_count", len(cachedItems))
                
                // Fresh data - return directly
                return cachedItems, nil
            }
        } else if err != nil {
            s.logger.Debug(ctx, "GetFoodItemsByTenant: Cache miss for tenant",
                "tenant_id", tenantID)
        }
    }
    
    // Step 3: Fall back to database (cache miss or stale data)
    var foodItems []*models.FoodItem
    if err := s.repository.ListByTenant(tenantID, &foodItems); err != nil {
        return nil, fmt.Errorf("database operation failed: %w", err)
    }
    
    s.logger.Info(ctx, "GetFoodItemsByTenant: Retrieved fresh data from database",
        "tenant_id", tenantID,
        "items_count", len(foodItems),
    )
    
    // Step 4: Populate all cache entries atomically (refresh entire batch)
    if s.cacheClient != nil {
        for i := range foodItems {
            cacheKey := fmt.Sprintf("food-item:%s", foodItems[i].ID)
            setErr := s.cacheClient.Set(ctx, cacheKey, *foodItems[i], cache.WithTTL(30*time.Minute))
            
            if setErr != nil {
                s.logger.Error(ctx, "GetFoodItemsByTenant: Failed to cache individual item",
                    "tenant_id", tenantID,
                    "item_id", foodItems[i].ID,
                    "error", setErr)
            } else {
                s.logger.Debug(ctx, "GetFoodItemsByTenant: Cached individual item",
                    "tenant_id", tenantID,
                    "item_id", foodItems[i].ID)
            }
        }
        
        // Delete tenant-level wildcard entry if it exists (prevents serving stale batch data)
        _ = s.cacheClient.Delete(cacheKey)
    }
    
    return foodItems, nil
}

func (s *FoodItemService) isStale(item *models.FoodItem) bool {
    if item.UpdatedAt == nil {
        return true
    }
    return time.Since(*item.UpdatedAt) > defaultStalenessThreshold
}
```

---

### Task 6: Implement Delete with Cache Invalidation

**What to implement**: Remove food item from database and invalidate cache entry

**Acceptance Criteria**:
- [ ] Accept ID string, perform database deletion using repository's Delete() method
- [ ] Validate item exists before deletion attempt
- [ ] Invalidate cache entry by calling `Delete()` with key format `food-item:{id}`
- [ ] Return error if deletion fails (entity not found in DB)

**Code Example**:
```go
func (s *FoodItemService) Delete(
    ctx context.Context,
    id string,
) error {
    s.logger.Info(ctx, "DeleteFoodItem: Starting deletion",
        "item_id", id)
    
    // Step 1: Validate item exists in database before attempting deletion
    var existingItem models.FoodItem
    if err := s.repository.GetByID(id, &existingItem); err != nil {
        return fmt.Errorf("food item not found or database error: %w", err)
    }
    
    s.logger.Info(ctx, "DeleteFoodItem: Item exists - proceeding with deletion",
        "item_id", id,
        "category_id", existingItem.Category,
        "tenant_id", existingItem.TenantID,
    )
    
    // Step 2: Delete from database (repository handles tenant validation via RLS)
    deleteErr := s.repository.Delete(id)
    
    if deleteErr != nil {
        return fmt.Errorf("database operation failed: %w", deleteErr)
    }
    
    s.logger.Info(ctx, "DeleteFoodItem: Food item deleted successfully from database",
        "item_id", id,
    )
    
    // Step 3: Invalidate cache entry using key format 'food-item:{id}'
    if s.cacheClient != nil {
        cacheKey := fmt.Sprintf("food-item:%s", id)
        deleteErr := s.cacheClient.Delete(cacheKey)
        
        if deleteErr != nil {
            return fmt.Errorf("failed to invalidate cache: %w", deleteErr)
        }
        
        s.logger.Info(ctx, "DeleteFoodItem: Cache entry invalidated successfully",
            "item_id", id,
            "cache_key", cacheKey,
        )
    } else {
        s.logger.Debug(ctx, "DeleteFoodItem: No cache client available - skipping cache invalidation",
            "item_id", id)
    }
    
    return nil
}
```

---

## Testing Strategy

### Unit Test Structure

```go
package fooditem

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/pareshvernekar/homecooked/internal/cache"
    logger "github.com/pareshvernekar/homecooked/internal/logger"
    models "github.com/pareshvernekar/homecooked/internal/models"
)

// Mock types for testing
type MockRepository struct {
    CreateFunc  func(item *models.FoodItem) error
    GetByIDFunc func(id string, item *models.FoodItem) error
    ListByTenantFunc func(tenantID string, offset, limit int) ([]models.FoodItem, int64, error)
    DeleteFunc  func(id string) error
}

// Test Create - Success case with category resolution
func TestCreate_Success(t *testing.T) {
    // Arrange
    mockRepo := &MockRepository{
        CreateFunc: func(item *models.FoodItem) error {
            // Verify item has resolved UUID (not name string) in Category field
            if item.Category != expectedCategoryUUID {
                t.Errorf("Expected category UUID '%s', got '%s'", expectedCategoryUUID, item.Category)
            }
            return nil
        },
    }
    
    service := NewFoodItemService(mockRepo, logger.NewLogger(), nil)
    ctx := context.Background()
    
    // Create test data with category NAME (not UUID)
    testItem := &models.FoodItem{
        Name: "Test Dish",
        Category: "vegetarian", // <-- String name, not UUID
        Price: 19.99,
        IsVegetarian: true,
        AvailabilityStatus: "available",
    }
    
    tenantID := "test-tenant-123"
    
    // Act
    createdItem, err := service.Create(ctx, tenantID, testItem)
    
    // Assert
    if err != nil {
        t.Fatalf("Create failed unexpectedly: %v", err)
    }
    
    if createdItem == nil {
        t.Fatal("Expected non-nil created item")
    }
    
    // Verify returned item has resolved UUID (not original name string)
    if createdItem.Category != expectedCategoryUUID {
        t.Errorf("Expected category UUID, got: %s", createdItem.Category)
    }
}

// Test GetByID - Cache Hit
func TestGetByID_CacheHit(t *testing.T) {
    // Arrange - setup mock that returns cached item instead of querying DB
    service := NewFoodItemService(nil, nil, &MockCacheClient{
        Items: map[string]models.FoodItem{"test-item-123": {ID: "test-item-123", Name: "Cachetest"}}
    })
    
    ctx := context.Background()
    reqId := "test-item-123"
    
    // Act
    item, err := service.GetByID(ctx, reqId)
    
    // Assert - Verify cache.Get was called (not repository.GetByID)
}

// Test GetByID - Cache Stale, DB Refresh
func TestGetByID_CacheStale_DBRefresh(t *testing.T) {
    // Arrange - setup stale item in mock cache (old UpdatedAt timestamp)
    // Act - should fall back to DB
    // Assert - Verify repository.GetByID was called with fresh data
}

// Test GetByTenant - Batch Stale Detection
func TestGetByTenant_BatchStaleDetection(t *testing.T) {
    // Arrange - setup mock with 5 items, one stale (5min+ ago)
    // Act - should refresh entire batch from DB
    // Assert - Verify ListByTenant called instead of cache return
}

// Test Delete - Success
func TestDelete_Success(t *testing.T) {
    // Arrange - item exists in mock database
    service := NewFoodItemService(mockRepo, logger.NewLogger(), nil)
    ctx := context.Background()
    
    // Act
    err := service.Delete(ctx, "test-item-123")
    
    // Assert - DB deleted AND cache invalidated (cache.Delete called)
}

// Test Create - Category Not Found
func TestCreate_CategoryNotFound(t *testing.T) {
    // Arrange - mock returns error when GetCategoryByName is invoked
    service := NewFoodItemService(mockRepo, logger.NewLogger(), nil)
    
    ctx := context.Background()
    testItem := &models.FoodItem{
        Name: "Test Dish",
        Category: "nonexistent-category", // Doesn't exist in tenant's catalog
        Price: 19.99,
    }
    
    // Act
    _, err := service.Create(ctx, tenantID, testItem)
    
    // Assert - Returns error mentioning category not found
}

// Test Delete - Item Not Found
func TestDelete_ItemNotFound(t *testing.T) {
    // Arrange - item doesn't exist in mock DB
    service := NewFoodItemService(mockRepo, logger.NewLogger(), nil)
    ctx := context.Background()
    
    // Act
    err := service.Delete(ctx, "nonexistent-id-123")
    
    // Assert - Returns error mentioning item not found
}

func TestIsStale_NoUpdatedAt(t *testing.T) {
    var item models.FoodItem // No fields set = nil UpdatedAt
    
    stale := s.isStale(&item)
    if !stale {
        t.Error("Expected item without UpdatedAt to be marked as stale")
    }
}

// Helper function for staleness detection (used in tests too)
func (s *FoodItemService) isStale(item *models.FoodItem) bool {
    // Same implementation as service
    if item.UpdatedAt == nil {
        return true
    }
    return time.Since(*item.UpdatedAt) > defaultStalenessThreshold
}

```

---

## Integration Checklist

### Before Pull Request

- [ ] **Code Compilation**: `go build ./...` - no compilation errors
- [ ] **Unit Tests Pass**: `go test -v ./services/fooditem/...` - all tests green
- [ ] **Linting Passes**: `golangci-lint run` - no violations
- [ ] **GetCategoryByName Works**: Category name resolution returns correct UUIDs
- [ ] **Cache Operations**: Set/Delete/Has methods called correctly with proper keys

### Functionality Verification

- [ ] Create food item with category name → stored with resolved UUID, CreatedAt/UpdatedAt populated
- [ ] Get single item from cache (when fresh) - returns complete entity
- [ ] Get single item from DB (when stale) - returns complete entity with cache refresh
- [ ] Get all items - returns complete list of entities when no staleness detected
- [ ] Get all items - returns complete list after batch stale detection and refresh
- [ ] Delete removes from both DB and cache
- [ ] Category not found error includes descriptive message with valid suggestions

---

## Common Patterns & Gotchas

### Cache Key Format Validation

```go
// Always validate cache key format before use
func validateCacheKey(key string) error {
    parts := strings.Split(key, ":")
    if len(parts) != 2 {
        return errors.New("invalid cache key format: expected entity_type:identifier")
    }
    entityType, identifier := parts[0], parts[1]
    
    switch entityType {
    case "food-item", "food-category":
        // Valid entityType
        return nil
    default:
        return fmt.Errorf("unknown entity type: %s", entityType)
    }
}

// Use in service:
if err := validateCacheKey(cacheKey); err != nil {
    s.logger.Error(ctx, "Invalid cache key format", "key", cacheKey, "error", err)
    return nil, err
}
```

---

**Document Version**: 1.0  
**Last Updated**: 2026-05-28  
**Next Step**: Review architecture decisions, proceed to implementation
