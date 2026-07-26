# Feature Specification: Food Category CRUD API with Cache Integration

**Feature Branch**: `002-food-category-crud-api`  
**Created**: 2026-07-04  
**Status**: Draft  
**Input**: User description: "Implement the CRUD API for FoodCategory using the FoodCategory service and FoodCategory repository. FoodCategory Create/Update/Delete should update the FoodCategory cache and FoodCategory read should check the FoodCategory cache for updated entries or refresh the cache in case of stale items"

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Admin Creates Food Categories (Priority: P1)

As an admin, I want to create food categories so that food items can be properly organized and users can filter by dietary preferences.

**Why this priority**: This is a foundational feature required for all food item operations. Without categories, food items cannot be properly associated with dietary requirements (vegetarian, non-vegetarian, vegan).

**Independent Test**: Can be fully tested by having an admin authenticated, sending a POST request to create a category, and verifying the category is stored in the database and returns from cache without database queries.

**Acceptance Scenarios**:

1. **Given** an admin is authenticated with valid JWT token, **When** they send a POST request with a valid category name (e.g., "Vegetarian") and description to `/api/v1/categories` (without id field), **Then** the system should auto-generate a primary key in the repository, create the category in the database, update the cache with the new entry, and return HTTP 200 with the created category (including generated ID).

2. **Given** a category already exists with the same name (case-insensitive), **When** an admin sends a POST request to create it again without specifying an id, **Then** the system should detect the duplicate, log a warning, continue operation gracefully, and return HTTP 200 with the existing category data (not 201).

3. **Given** no category exists for the given name, **When** an admin sends a POST request to create it, **Then** the system should auto-generate an id in the repository layer, persist to database, and return HTTP 200 with the new category including generated id.

---

### User Story 2 - Admin Retrieves Food Categories (Priority: P1)

As an admin, I want to retrieve all food categories for my tenant so that I can view available options and check cache performance.

**Why this priority**: This is required for displaying menu options to users and validating food item data integrity. It's the primary read operation for category management.

**Independent Test**: Can be fully tested by listing cached categories without hitting the database, verifying cache freshness by comparing timestamps, and ensuring stale cache entries are automatically refreshed on subsequent requests.

**Acceptance Scenarios**:

1. **Given** food categories exist in the cache with a known timestamp, **When** an admin sends a GET request to `/api/v1/categories`, **Then** the system should return cached data without hitting the database if the cache is fresh (within TTL).

2. **Given** food categories have expired from cache (TTL elapsed), **When** an admin sends a GET request, **Then** the system should detect stale data and automatically refresh/repopulate the cache from the database before returning results to ensure data accuracy.

3. **Given** an admin is viewing categories after another admin created new ones for the same tenant, **When** the previous request cached old data, **Then** subsequent requests should recognize the update pattern (create followed by read) and refresh stale entries automatically.

---

### User Story 3 - Admin Updates Food Categories (Priority: P2)

As an admin, I want to update existing food categories so that I can fix typos, improve descriptions, or toggle visibility without recreating records.

**Why this priority**: Update operations are common maintenance tasks that admins will perform regularly. Supporting it reduces redundant database writes and maintains data integrity.

**Independent Test**: Can be fully tested by sending a PUT request to update category details, verifying the cache is updated with new values before returning the response, and ensuring old cached versions are replaced entirely.

**Acceptance Scenarios**:

1. **Given** a food category exists in the database and has been cached, **When** an admin sends a PUT request to update its name or description, **Then** the system should update the cache with new values before returning the response, ensuring users see updated data immediately.

2. **Given** categories are stale in cache from earlier operations, **When** an update occurs on an existing category and then another read happens, **Then** the refresh pattern (update followed by read) should trigger automatic cache invalidation and repopulation.

---

### User Story 4 - Admin Deletes Food Categories (Priority: P2)

As an admin, I want to delete inactive food categories so that users won't see them in menu selections.

**Why this priority**: Deletion is a core CRUD operation required for complete category lifecycle management. Soft deletion preserves data integrity while removing categories from active selection.

**Independent Test**: Can be fully tested by sending a DELETE request, verifying database row is marked inactive, confirming cache entry is cleared, and ensuring subsequent reads no longer return the deleted category.

**Acceptance Scenarios**:

1. **Given** a food category exists in the database and is cached for tenant viewing, **When** an admin sends a DELETE request to mark it as inactive, **Then** the system should soft-delete from database (set is_active=false), clear the cache entry entirely, and return HTTP 204 No Content.

2. **Given** categories are stale in cache, **When** they are deleted via DELETE operation followed by a GET list request, **Then** the refresh pattern should detect the deletion and rebuild the complete cached dataset from database to exclude deleted items.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST expose RESTful CRUD endpoints for food category management:
     - `POST /api/v1/categories` - Create a new food category (id auto-generated by repo)
     - `GET /api/v1/categories` - List all food categories for tenant
     - `PUT /api/v1/categories/{id}` - Update an existing food category by ID
     - `DELETE /api/v1/categories/{id}` - Delete/deactivate a food category

- **FR-002**: System MUST enforce multi-tenant isolation where each tenant can only manage their own categories through the `X-Tenant-ID` header, ensuring no cross-tenant data leakage.

- **FR-003**: System MUST validate input data at the handler level before delegating to service layer, including name length (1-100 characters), required field checks, and sanitization.

- **FR-004**: System MUST implement semantic caching using `entity_type:name` key format (e.g., `food_category:vegetarian`) for all read operations to enable efficient cache lookups by name.

- **FR-005**: System MUST propagate cache updates for Create and Update operations immediately, ensuring subsequent reads return fresh data without stale entries.

- **FR-006**: System MUST implement cache refresh logic for GET operations that detects stale entries (expired TTL) and automatically repopulates from database before serving results.

- **FR-007**: System MUST propagate cache deletion for Delete operations, removing the specific entry to prevent accidental retrieval of deleted data.

- **FR-008**: System MUST return comprehensive error responses with status codes, error codes, messages, and timestamps for all failed operations.

- **FR-009**: System MUST support pagination for list operations with configurable page size (default 20, max 100) and offset/limit parameters.

- **FR-010**: System MUST return complete category objects (not just IDs) from all endpoints to support downstream filtering and validation operations by name.

---

### Key Entities

- **FoodCategory Entity**: Represents a food category grouping used for organizing food items by dietary requirements or themes.
     - Key attributes: `id` (VARCHAR), `tenant_id` (VARCHAR), `name` (VARCHAR, unique per tenant), `description` (TEXT, optional), `is_active` (BOOLEAN), `created_at` (TIMESTAMP), `updated_at` (TIMESTAMP)
     - Relationships: Many-to-many with FoodItem through `food_item.category_id` foreign key; cached independently for efficient lookups

- **FoodCategory Cache Entry**: In-memory representation keyed by `entity_type:name` pattern enabling semantic search.
     - Format: `food_category:{lowercase_name}` (e.g., `food_category:vegetarian`, `food_category:chinese`)
     - TTL: Configurable per tenant or global default (30 minutes)
     - Contains full entity object for database persistence requirements

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Create, Update, and Delete operations MUST propagate cache changes synchronously before returning responses, ensuring zero-latency consistency for downstream reads.

- **SC-002**: GET operations MUST check cache freshness before every read request; if entries are stale (expired or outdated), the system MUST refresh from database automatically without user-perceived delay.

- **SC-003**: All CRUD operations MUST enforce tenant isolation with 100% accuracy - no cross-tenant data leakage verified through row-level security and cache key scoping.

- **SC-004**: Cache hit rate for GET operations must be >= 50% on repeated requests (second request within TTL), measured as database queries avoided per operation count.

- **SC-005**: Cache refresh logic must automatically detect stale entries based on TTL expiration and re-populate from database without manual cache invalidation calls.

- **SC-006**: List operations (GET all categories) MUST fetch fresh data whenever any tenant performs a Create/Update/Delete, triggering automatic cache refresh for the entire dataset.

---

## Assumptions

1. **Cache Refresh Pattern**: When a read operation follows a write operation from another admin request within TTL window, the system treats it as stale and refreshes from database automatically.

2. **Cache Key Format**: All cached entries use `entity_type:name` convention where entity_type is `food_category` and name is lowercase slugified category name (e.g., `food_category:vegetarian`, `food_category:chinese`).

3. **Multi-Tenancy Enforcement**: PostgreSQL Row-Level Security (RLS) policies filter database queries by tenant ID; cache keys include tenant context to prevent cross-tenant caching.

4. **Soft Delete Pattern**: Delete operations set `is_active=false` rather than removing records permanently, preserving referential integrity for existing food items.

5. **Caching Strategy**:
     - GET operations: Check cache first → return cached if fresh → refresh from DB and return if stale
     - POST/PUT operations: Write to DB → update cache immediately → return response
     - DELETE operations: Soft delete from DB → remove cache entry

6. **Existing Infrastructure**:
     - Service layer exists at `internal/services/foodcategory/service.go` with PostInitialize method
     - Repository layer exists at `internal/repository/foodcategory_repository.go` with PostgreSQL implementation
     - Cache client supports `entity_type:name` key pattern and TTL configuration
     - Logger infrastructure available for structured logging

7. **Tenant Context**: Tenant ID extracted from `X-Tenant-ID` header via middleware and injected into all repository and cache operations.

---

## API Contract

### Request Headers (All Operations)

| Header | Type | Description | Example |
|--------|------|-------------|---------|
| X-Tenant-ID | string | Tenant identifier for multi-tenancy isolation | `tenant-123` |
| Content-Type | string | Required for POST/PUT operations | `application/json` |

### 1. Create Category (POST /api/v1/categories)

#### Request

```http
POST /api/v1/categories
X-Tenant-ID: tenant-123
Content-Type: application/json

{
    "name": "Vegetarian",        // Required
    "description": "Plant-based food options",   // Optional
    "is_active": true    // Optional, defaults to true
}
```

**Note**: The `id` field is NOT accepted in POST requests. Primary keys are auto-generated by the PostgreSQL repository using UUID v4 convention and returned in the response.

#### Success Response (200 OK) - New Category Created

```json
{
       "success": true,
       "error_code": "",
       "message": "Food category created successfully",
       "data": {
         "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",  // Auto-generated by PostgreSQL repo
         "tenant_id": "tenant-123",
         "name": "vegetarian",   // Normalized to lowercase
         "description": "Plant-based food options",
         "is_active": true,
         "created_at": 1719888000000,
         "updated_at": 1719888000000
       }
}
```

#### Success Response (200 OK) - Category Already Exists (Idempotent Duplicate Handling)

```json
{
       "success": true,
       "error_code": "",
       "message": "Food category already exists",
       "data": {
         "id": "existing-category-id-from-db",   // Existing ID returned (not generated)
         "tenant_id": "tenant-123",
         "name": "vegetarian",
         "description": "Plant-based food options",
         "is_active": true,
         "created_at": 1719888000000,     // Original creation timestamp preserved
         "updated_at": 1719888600000       // Updated timestamp on duplicate create
       }
}
```

**Note**: Returns HTTP 200 (not 201) for all successful creates - this is intentional idempotent behavior. Duplicate submissions return existing data instead of creating a new entry.

#### Error Response (400 Bad Request)

```json
{
       "success": false,
       "error_code": "VALIDATION_ERROR",
       "message": "Invalid category data",
       "details": {
         "field": "name",
         "reason": "Name must be between 1 and 100 characters"
       },
       "timestamp": "2026-07-04T10:30:00.000Z"
}
```

#### Cache Behavior

- **On New Create**: Cache entry created at `food_category:{lowercase_name}` with TTL from config after successful DB insertion, returns 200 OK
- **On Duplicate**: Logs warning, returns existing cached or database data (no new cache write needed), returns 200 OK  
- **Error Path**: No cache modification on validation errors; id field rejected if provided in request body

---

### 2. List Categories (GET /api/v1/categories)

#### Request

```http
GET /api/v1/categories
X-Tenant-ID: tenant-123
```

#### Success Response (200 OK) - From Cache (Hit)

```json
{
       "success": true,
       "error_code": "",
       "message": "Food categories retrieved successfully",
       "data": [
         {
           "id": "cat-uuid-1",
           "tenant_id": "tenant-123",
           "name": "vegetarian",
           "description": "Plant-based options",
           "is_active": true,
           "created_at": 1719888000000,
           "updated_at": 1719888000000
         },
         {
           "id": "cat-uuid-2",
           "tenant_id": "tenant-123",
           "name": "chinese",
           "description": null,
           "is_active": true,
           "created_at": 1719888500000,
           "updated_at": 1719888500000
         }
       ],
       "cache_info": {
         "source": "cache",
         "cached_at": 1719888000000
       }
}
```

#### Success Response (200 OK) - From Database (Miss/Stale Refresh)

```json
{
       "success": true,
       "error_code": "",
       "message": "Food categories retrieved successfully",
       "data": [
         {
           "id": "cat-uuid-1",
           "tenant_id": "tenant-123",
           "name": "vegetarian",
           "description": "New updated description",
           "is_active": true,
           "created_at": 1719888000000,
           "updated_at": 1719888500000    // Fresh timestamp from DB
         }
       ],
       "cache_info": {
         "source": "database_refresh",
         "refreshed_at": 1719888600000,
         "previous_cache_age_seconds": 3600
       }
}
```

#### Error Response (400 Bad Request)

```json
{
       "success": false,
       "error_code": "INVALID_REQUEST",
       "message": "Missing tenant ID in request headers",
       "details": {
         "field": "X-Tenant-ID",
         "reason": "Required header not provided"
       },
       "timestamp": "2026-07-04T10:35:00.000Z"
}
```

#### Cache Behavior

- **Cache Hit**: Return cached data directly, log cache hit with timestamp, include `cache_info.source: "cache"`
- **Cache Miss (New Key)**: Fetch from database, populate cache, return fresh data, log cache miss with TTL info
- **Stale Data Detection**: If cache entry expired (TTL elapsed) or update pattern detected from recent operations, auto-refresh from DB before returning
- **Refresh Trigger**: Automatic refresh occurs when `cache_info.source: "database_refresh"` indicates database was queried

---

### 3. Update Category (PUT /api/v1/categories/{id})

#### Request

```http
PUT /api/v1/categories/cat-uuid-456
X-Tenant-ID: tenant-123
Content-Type: application/json

{
     "name": "Modified Vegetarian",
     "description": "Updated plant-based options description",
     "is_active": true
}
```

#### Success Response (200 OK)

```json
{
       "success": true,
       "error_code": "",
       "message": "Food category updated successfully",
       "data": {
         "id": "cat-uuid-456",
         "tenant_id": "tenant-123",
         "name": "Modified Vegetarian",
         "description": "Updated plant-based options description",
         "is_active": true,
         "created_at": 1719888000000,
         "updated_at": 1719888600000    // New timestamp
       },
       "cache_info": {
         "source": "database",
         "cache_updated": true,
         "evicted_stale_entries": ["food_category:vegetarian"]
       }
}
```

#### Error Response (404 Not Found)

```json
{
       "success": false,
       "error_code": "NOT_FOUND",
       "message": "Food category not found",
       "details": {
         "field": "id",
         "value": "cat-uuid-invalid"
       },
       "timestamp": "2026-07-04T10:40:00.000Z"
}
```

#### Cache Behavior

- **Immediate Update**: After successful database update, cache entry is refreshed with new values before response is returned
- **Cache Key**: Updated at `food_category:{lowercase_name}` to reflect current state
- **Stale Prevention**: Ensures subsequent reads (GET operations) immediately see updated data without TTL delay
- **Eviction Flag**: Logs any stale cache entries that were replaced during update operation

---

### 4. Delete Category (DELETE /api/v1/categories/{id})

#### Request

```http
DELETE /api/v1/categories/cat-uuid-456
X-Tenant-ID: tenant-123
```

#### Success Response (204 No Content)

```
HTTP/1.1 204 No Content
Cache-Control: no-cache
```

#### Error Response (404 Not Found)

```json
{
       "success": false,
       "error_code": "NOT_FOUND",
       "message": "Food category not found",
       "details": {
         "field": "id",
         "value": "cat-uuid-invalid"
       },
       "timestamp": "2026-07-04T10:45:00.000Z"
}
```

#### Cache Behavior

- **Immediate Deletion**: Cache entry removed synchronously after successful database soft-delete
- **Key Pattern**: Remove cache entries matching `food_category:{lowercase_name}` pattern for deleted categories
- **Stale Prevention**: Ensures GET operations no longer return deleted data from cache
- **List Impact**: Subsequent list requests will reflect updated dataset without stale entries

---

## Non-Functional Requirements

1. **Performance**: Cache operations should complete within 10ms for read operations when hit rate >= 50%
2. **Reliability**: Database fallback must succeed within 500ms timeout, with graceful error handling
3. **Scalability**: Cache client supports max 1000 items per tenant as configured
4. **Observability**: All cache hits/misses/refreshes logged at DEBUG level with key patterns and timestamps
5. **Security**: Tenant isolation enforced at database (RLS) and cache (key scoping) layers
6. **Idempotency**: Create operations tolerate duplicate submissions gracefully

---

## Testing Strategy

### Unit Tests

**Handler Tests**:
- Mock `FoodCategoryService` with configurable response injection
- Mock `cache.Client` to track cache operations (Set, Delete, Get calls)
- Test scenarios:
     - Successful create → cache Set verified
     - Duplicate create → warning logged, existing data returned
     - Successful read (cache hit) → no DB call
     - Successful read (cache miss/stale) → DB refresh + cache populate
     - Successful update → cache Set verified after response
     - Delete → cache Delete verified
     - Validation errors → proper error responses

**Cache Synchronization Tests**:
- Verify POST triggers `cache.Set()` before response
- Verify PUT/DELETE trigger `cache.Set()/Delete()` before response
- Verify GET checks expiration: `if !cache.IsExpired(key) { return cached } else { refreshFromDB() }`

---

## Edge Cases

1. **Cache Staleness Detection**: How to identify stale entries?
     - Via TTL expiration (client tracks expiration timestamp)
     - Via update pattern detection (recent create/update before read)
     - Implementation: `if cacheEntry == nil || IsExpired(cacheEntry) { refreshFromDB() }`

2. **Concurrent Updates**: Multiple admins updating same category simultaneously?
     - Database handles concurrent updates via row locks
     - Cache is updated on each write, ensuring eventual consistency
     - Read operations will detect staleness and refresh if TTL expired

3. **Partial Data Updates**: What if only description changes, not name?
     - Update handler extracts all fields from request
     - Missing fields in request → no change to existing value
     - Cache key unchanged (name-based), but cache value updated with new data

4. **Cache Refresh Trigger**: When exactly does automatic refresh occur?
     - On GET if TTL has elapsed since last populate
     - After any CREATE/UPDATE operation, next GET should refresh
     - Implementation: `if time.Since(cacheEntry.CreatedAt) > config.global_ttl { refresh() }`

---

## Dependencies & Assumptions

- **Existing Files**:
     - `internal/services/foodcategory/service.go` - Service layer with PostInitialize method
     - `internal/repository/foodcategory_repository.go` - Repository interface and PostgreSQL implementation
   
- **Dependencies on Existing Features**:
     - Tenant context middleware (X-Tenant-ID header extraction) ✓ Complete
     - Logger service ✓ Complete
     - Cache client with entity_type:name key pattern ✓ Complete
     - Error response views (`internal/views/errorresponse.go`) ✓ Complete

- **Assumptions**:
     - Cache client supports `Set(ctx, key, value)` for create/update operations
     - Cache client supports `Delete(ctx, tenantID, keyPattern)` for delete operations  
     - Cache client supports `Get(ctx, key)` to check freshness and detect staleness
     - TTL configuration available in cache config (30 minutes default)

---

## Implementation Checklist

- [ ] Create FoodCategory handler (`internal/handlers/foodcategory.go`)
     - Implement `CreateCategory`, `ListCategories`, `UpdateCategory`, `DeleteCategory` methods
     - Wire service dependency injection with repository and cache client
     - Add HTTP route handlers using Gin framework
   
- [ ] Set up HTTP routes (`internal/server/server.go`)
     - POST `/api/v1/categories`
     - GET `/api/v1/categories`
     - PUT `/api/v1/categories/{id}`
     - DELETE `/api/v1/categories/{id}`
   
- [ ] Implement Cache Synchronization Logic
     - **POST Create**: Update cache with new entry → `s.cacheClient.Set(ctx, key, item)` after successful DB insert
     - **GET Read (List)**: Check cache freshness → if stale/expired → refresh from DB; if fresh → return cached
     - **PUT Update**: Update cache after DB update → `s.cacheClient.Set(ctx, key, updated_item)`
     - **DELETE**: Remove cache entry → `s.cacheClient.Delete(ctx, tenantID, categoryKey)`

- [ ] Implement Cache Refresh Logic for Read Operations
     - Detect stale entries via TTL expiration check: `s.cacheClient.IsExpired(ctx, key)` or `s.cacheClient.Get(ctx, key)` returns nil/old data
     - Auto-refresh pattern: When GET request follows recent CREATE/UPDATE from same tenant, refresh cache automatically
     - Cache invalidation trigger: After any mutation operation, mark related cache keys as stale for next read

- [ ] Add Input Validation in Handler Layer
     - Validate required fields before delegating to service
     - Call `category.IsValid()` from model validation
     - Return appropriate error responses with detail messages

- [ ] Create Unit Tests (`internal/handlers/foodcategory_test.go`)
     - Mock repository and cache client for handler tests
     - Test all CRUD operations with various scenarios
     - Verify cache sync behavior in tests (mock tracking)

---

## Completion Criteria

✅ Feature branch `002-food-category-crud-api` created  
✅ Specification file created at `specs/002-food-category-crud-api/spec.md`  
✅ All functional requirements documented  
✅ Success criteria defined and measurable  
✅ Cache synchronization strategy defined for all CRUD operations  
✅ Cache refresh logic defined for read operations with stale detection  
✅ Implementation checklist complete  

**Key Updates Applied**:
- POST `/api/v1/categories` no longer accepts `id` field in request body (auto-generated by PostgreSQL repo)
- All successful CREATE operations return HTTP 200 OK (not 201 Created) - idempotent behavior
- Duplicate submissions to POST return existing category data with 200 OK status
