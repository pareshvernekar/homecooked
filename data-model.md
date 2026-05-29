# Data Model: FoodItem-CRUD-Service

**Created**: 2026-05-28  
**Feature Branch**: `002-cache-initialization-from-food-categories`  
**Status**: Implementation Ready ✅  
**Specification**: [`FoodItem-CRUD-Service.spec.md`](specs/food_menu_system/FoodItem-CRUD-Service.spec.md)

---

## Table of Contents

1. [Domain Entities](#domain-entities)
2. [Database Schema](#database-schema)
3. [Entity Relationships](#entity-relationships)
4. [Cache Structure](#cache-structure)
5. [Service Interface Contracts](#service-interface-contracts)
6. [Validation Rules](#validation-rules)

---

## Domain Entities

### 1. FoodItem (Primary Entity)

**Purpose**: Represents individual menu items in a food catalog with category association and availability status.

#### Fields

| Field | Type | Nullable | Description | Source |
|-------|------|----------|-------------|--------|
| ID | UUID string | No | Primary key, auto-generated | Database |
| TenantID | string | No | Multi-tenant isolation (RLS context) | Request Context |
| Name | string | Required | Display name of the food item (100 chars max) | API Body |
| Description | `*string` | Optional | Detailed description of the item | API Body |
| Category | UUID string | Required | Reference to FoodCategory primary key | Resolved via GetCategoryByName() |
| Price | float64 | Required | Item price (must be >= 0) | API Body |
| ImageURL | `*string` | Optional | URL for item image | API Body |
| Avoidance | `*string` | Optional | Dietary avoidance info (e.g., "no-peanuts") | API Body |
| IsVegetarian | bool | Required | Boolean flag for vegetarian classification | API Body |
| AvailabilityStatus | string | Required | Current availability: "available", "unavailable", or "low_stock" | API Body |
| CreatedAt | `*time.Time` | No | Record creation timestamp | System-generated |
| UpdatedAt | `*time.Time` | No | Last update timestamp (auto-incremented) | System-generated |

#### Entity Relationships

- **TenantID** → Row-Level Security context (external reference via PostgreSQL RLS)
- **Category** → One-to-One with FoodCategory (UUID primary key reference)
- **CreatedAt/UpdatedAt** → Temporal audit trail for cache staleness detection

### 2. FoodCategory (Context Entity)

**Purpose**: Defines food category labels for organizational purposes (e.g., "vegetarian", "non-vegetarian", "dessert").

#### Fields

| Field | Type | Nullable | Description | Source |
|-------|------|----------|-------------|--------|
| ID | UUID string | No | Primary key | Database |
| TenantID | string | No | Multi-tenant isolation (RLS context) | Request Context |
| Name | string | Required | Category label (100 chars max, kebab-case preferred) | API Body |
| Description | `*string` | Optional | Category description | API Body |
| CreatedAt | `time.Time` | No | Record creation timestamp | System-generated |
| UpdatedAt | `time.Time` | No | Last update timestamp | System-generated |

#### Relationship to FoodItem

FoodItem.Category field stores the **FoodCategory.ID** (UUID) as a foreign key reference, enabling:
- Efficient database queries using primary key lookups
- Case-insensitive matching at service layer via `strings.EqualFold()`
- Data integrity through UUID constraints

### 3. Cache Entry Patterns (In-Memory Store)

#### FoodCategory Cache Structure

```
Key Format: food-category:{category-name}
Examples:
    - food-category:vegetarian
    - food-category:non-vegetarian
    - food-category:dessert
```

**Fields Cached**:
- Full `FoodCategory` entity (all fields) as pointer
- TTL: 24 hours (rare category changes)

#### FoodItem Cache Structure

```
Key Format: food-item:{item-id}
Examples:
    - food-item:a1b2c3d4-e5f6-7890-abcd-ef1234567890
    - food-item:b2c3d4e5-f6a7-8901-bcde-f12345678901
```

**Fields Cached**:
- Full `FoodItem` entity (all fields) as pointer
- Includes resolved Category UUID for DB compatibility
- TTL: 30 minutes (frequent menu updates)

---

## Service Interface Contracts

### FoodItemService Interface

**Design Principle**: All public methods return complete domain entities via pointers for consistency and proper entity lifecycle management.

```go
type FoodItemService interface {
    // Create: Accepts full entity with category NAME (string), resolves internally to UUID before persisting
    //         Returns created entity pointer (*models.FoodItem) with resolved Category ID, CreatedAt/UpdatedAt timestamps
    //         - Caller immediately receives persisted data without additional DB call
    Create(ctx context.Context, tenantID string, item *models.FoodItem) (*models.FoodItem, error)

    // GetByID: Returns complete *FoodItem entity pointer from cache or database
    //          - Checks cache first for existing entry
    //          - Detects staleness based on UpdatedAt timestamp (> 5min default = stale)
    //          - Falls back to DB if stale/miss, refreshes cache atomically
    GetByID(ctx context.Context, id string) (*models.FoodItem, error)

    // GetByTenant: Returns complete slice of entity pointers []*FoodItem
    //              Batch stale detection iterates through cached items checking each UpdatedAt
    //              If ANY item is stale (UpdatedAt > threshold), refreshes ENTIRE batch from DB
    //              Returns full list with all fields populated as pointer references
    GetByTenant(ctx context.Context, tenantID string) ([]*models.FoodItem, error)

    // Delete: Removes food item from database and invalidates cache entry using key format (food-item:{id})
    Delete(ctx context.Context, id string) error

    // Helper method to resolve category name to ID - used internally before Create/Update operations
    GetCategoryByName(categoryName, tenantID string) (string, error)
}
```

### FoodCategoryService Interface Extension

```go
// Existing methods preserved - both return pointer slices for consistency with domain entity patterns
ListCategories(tenantID string) ([]*models.FoodCategory, error)      // Returns slice of pointers
GetCategoryByID(categoryID string, tenantID string) (*models.FoodCategory, error)     // Returns pointer

// New method for name-based resolution (internal helper for FoodItemService)
// Accepts: Category NAME from API request (string)
// Returns: Resolved UUID ID (string) for database persistence in FoodItem.category field
GetCategoryByName(categoryName, tenantID string) (string, error)
```

### Implementation Notes

1. **Create**: Service accepts complete `*models.FoodItem` with category name as string, resolves internally using GetCategoryByName to UUID, persists to database, then returns the created pointer with:
   - Resolved Category ID (UUID) instead of original name string
   - Populated CreatedAt timestamp
   - Populated UpdatedAt timestamp
   - Caller receives fully persisted data immediately, no round-trip needed

2. **GetByID**: Returns a pointer to the full `*models.FoodItem` entity from cache or database - never partial data

3. **GetByTenant**: Returns a slice of complete entity pointers `[]*models.FoodItem` with all items for tenant

4. **Delete**: Accepts ID string, performs database deletion and invalidates cache entry using key format `food-item:{id}`

5. **GetCategoryByName**: Special internal method that accepts category NAME (not ID) and returns resolved UUID - used ONLY to resolve names before calling Create/Update methods

---

## Validation Rules

### FoodItem Creation/Update Fields

| Field | Validation Rule | Error Code |
|-------|-----------------|------------|
| Name | Required, 1-100 chars, alphanumeric + spaces/dashes | `INVALID_REQUEST: "name"` |
| Price | Required, >= 0.00, max 999999.99 | `INVALID_REQUEST: "price"` |
| IsVegetarian | Required boolean enum (true/false) | `INVALID_REQUEST: "is_vegetarian"` |
| AvailabilityStatus | Required enum: available/unavailable/low_stock | `INVALID_REQUEST: "availability_status"` |
| Category | Required UUID, must reference valid tenant category | `CATEGORY_NOT_FOUND: "category_id"` |
| Description | Optional text (max 1000 chars) | `-` |
| ImageURL | Optional URL string | `-` |
| Avoidance | Optional text (dietary restrictions) | `-` |

### FoodCategory Creation Fields

| Field | Validation Rule | Error Code |
|-------|-----------------|------------|
| Name | Required, 1-100 chars, kebab-case recommended | `INVALID_REQUEST: "name"` |
| Description | Optional text (max 1000 chars) | `-` |
| TenantID | Inherited via RLS context | `TENANT_NOT_FOUND` |

---

## Error Contract

| Error Code | HTTP Status | Trigger Condition |
|------------|-------------|-------------------|
| CATEGORY_NOT_FOUND | 404 | Category name not found in tenant's catalog |
| INVALID_REQUEST | 400 | Validation failure on FoodItem/FoodCategory fields |
| DATABASE_ERROR | 500 | Database operation failure (Create/Update/Delete) |
| CACHE_MISS | - | Cache unavailable, fallback to DB |

---

## Integration Points

### External Dependencies

1. **PostgreSQL Database**
   - Row-Level Security for tenant isolation
   - UUID primary keys via `gen_random_uuid()`
   - Foreign key constraints between tables

2. **Cache Layer (Redis/GoCache)**
   - In-memory key-value store abstraction
   - TTL-based expiration (configurable per entity)
   - Thread-safe operations

3. **FoodCategoryService**
   - Category name-to-ID resolution via `strings.EqualFold()`
   - Returns UUID string for database persistence

---

## Implementation Notes

### Staleness Detection

```go
const DefaultStalenessThreshold = 5 * time.Minute

func isStale(entity models.FoodItem, updatedAt *time.Time) bool {
    if updatedAt == nil {
        return true // No timestamp = definitely stale
    }
    return time.Since(*updatedAt) > DefaultStalenessThreshold
}
```

### Cache Key Validation

```go
// Validate cache key format (entity_type:identifier)
func validateCacheKey(key string) error {
    parts := strings.Split(key, ":")
    if len(parts) != 2 {
        return errors.New("invalid cache key format")
    }
    entityType, identifier := parts[0], parts[1]

    switch entityType {
    case "food-item", "food-category":
        return nil
    default:
        return fmt.Errorf("unknown entity type: %s", entityType)
    }
}
```

---

**Document Version**: 1.0  
**Last Updated**: 2026-05-28  
**Next Step**: Review architecture decisions, proceed to implementation
