# Tasks: Cache Initialization from Food Categories

**Branch**: `002-cache-initialization-from-food-categories`  
**Date**: 2026-05-20  
**Spec**: [spec.md](../002-cache-initialization-from-food-categories/spec.md)  
**Plan**: [plan.md](./plan.md)  
**Data Model**: [data-model.md](./data-model.md)

---

## Overview

These tasks implement cache initialization from the `food_category` database table at application startup using a post-initialization hook pattern. The implementation populates the in-memory cache with tenant-scoped food categories, enabling semantic search lookups via `entity_type:name` key convention (e.g., `food-category:vegetarian`).

---

## Phase 1: Setup

### Task Overview
Initialize project structure and documentation artifacts for cache initialization feature.

**Story Goal**: Prepare the codebase with necessary documentation and contracts for implementation.

**Independent Test Criteria**: All generated files exist, contain valid content matching specifications, and reference correct plan documents.

---

### T001 Create project documentation directory structure
- [x] T001 Create docs/specs/cache_implementation directory at repository root ✓ COMPLETED
   - **Description**: Establish the documentation container for this feature branch.
   - **Dependencies**: None
   - **Parallel Execution**: Can run independently alongside other setup tasks
   - **Implementation Strategy**: `mkdir -p docs/specs/cache_implementation` followed by subdirectories for contracts

---

### T002 Create integration contract document
- [x] T002 Define cache API contract in docs/specs/cache_implementation/contracts/cache-initialization-api.md ✓ COMPLETED
   - **Story Goal**: Document the cache initialization interface contract
   - **Independent Test Criteria**: Contract specifies PostInitialize method signature, parameter types, return types, error conditions, and usage pattern examples
   - **Dependencies**: None (Phase 1)
   - **Parallel Execution**: Independent of other setup tasks
   - **Implementation Strategy**: Document the Client interface extension with PostInitialize method

---

### T003 Update CLAUDE.md with implementation plan reference
- [ ] T003 Add link to docs/specs/cache_implementation/plan.md in CLAUDE.md under SPECKIT markers
   - **Story Goal**: Make implementation plan accessible to development AI agents
   - **Independent Test Criteria**: CLAUDE.md contains reference to plan.md between `<!-- SPECKIT START -->` and `<!-- SPECKIT END -->` markers
   - **Dependencies**: T002 (contract document must exist first)
   - **Parallel Execution**: Can run independently
   - **Implementation Strategy**: Edit CLAUDE.md to add implementation plan reference

---

## Phase 2: Foundational Tasks

### Task Overview
Implement core cache initialization functionality with database querying and tenant-scoped population. These tasks are blocking prerequisites for all user stories.

**Story Goal**: Establish PostInitialize hook that initializes cache with food category data from database at startup.

**Independent Test Criteria**: 
- Cache contains all N food categories from database after initialization
- Zero database queries during application runtime for cached entity types
- 100% cache hit rate for food category lookups post-initialization
- < 5 second initialization time for 1000-record dataset
- No breaking changes to existing cache API users

---

### T004 Add PostInitialize method to Client interface in internal/cache/client.go
T004 ✓ COMPLETED\n- [x] T004 Extend cache.Client interface with PostInitialize method signature
   - **Story Goal**: Define the contract for post-initialization hook
   - **Independent Test Criteria**: 
     - Client interface includes `PostInitialize(ctx context.Context, tenantID string, db *sqlx.DB) error` method
     - Get, Set, Has methods remain unchanged (backward compatibility)
     - Interface compiles without errors
   - **Dependencies**: None (Phase 2 foundational)
   - **Parallel Execution**: Must be completed before implementation tasks
   - **Implementation Strategy**: Add method signature:
     ```go
     type Client interface {
         Get(ctx context.Context, key string) (any, error)
         Set(ctx context.Context, key string, value any, options ...SetOption) error
         Has(key string) bool
         PostInitialize(ctx context.Context, tenantID string, db *sqlx.DB) error   // NEW METHOD
     }
     ```

---

### T005 Implement PostInitialize method in cacheImpl struct using sqlx database querying with entity_type:name key format
- [x] T005 Implement database-driven cache population in `internal/cache/client.go` ✓ COMPLETED
   - **Story Goal**: Populate in-memory cache with food category records from database
   - **Independent Test Criteria**: 
     - PostInitialize method accepts context and sqlx.DB parameters
     - Method queries database using prepared statement: `SELECT id, tenant_id, name, description, created_at, updated_at FROM food_category WHERE tenant_id = $1`
     - For each result row, calls `cache.Set(ctx, key, value, store.WithExpiration(ttl))` where key follows format `food-category:<category-name>` (e.g., `food-category:vegetarian`, `food-category:dairy-free`)
     - Value stored as `*models.FoodCategory` (pointer to reuse existing domain model)
     - Logs completion with item count using logger passed via dependency injection
   - **Dependencies**: T004 (interface extension must exist first)
   - **Parallel Execution**: Sequential after T004 completes
   - **Implementation Strategy**: Implement database query and cache population loop

---

### T006 Add PostInitialize implementation to NewCacheClient function return type
- [ ] T006 Update `internal/cache/client.go` to ensure cacheImpl struct satisfies extended Client interface
   - **Story Goal**: Complete the dependency injection pattern with full interface compliance
   - **Independent Test Criteria**: 
     - `NewCacheClient(config CacheConfig) (Client, error)` returns type implementing full Client interface including PostInitialize
     - Interface compliance check: `assert cache.Client((*cacheImpl)(nil))` passes in Go compiler
   - **Dependencies**: T004 (interface extension), T005 (implementation)
   - **Parallel Execution**: Must complete after T004 and T005
   - **Implementation Strategy**: Ensure returned `*cacheImpl` satisfies extended Client interface; no changes needed to NewCacheClient signature itself, just ensure cacheImpl has PostInitialize method

---

### T007 Wire up cache initialization at application startup in main.go after cache client creation
- [ ] T007 Add cache population call in cmd/main.go bootstrap sequence
   - **Story Goal**: Trigger cache initialization with database connection at application start
   - **Independent Test Criteria**: 
     - Application starts successfully and populates cache before serving HTTP requests
     - Cache contains all food categories from database after startup completes
     - No database queries needed for category lookups during application lifetime
   - **Dependencies**: T006 (interface must be implemented first)
   - **Parallel Execution**: Sequential dependency on T006
   - **Implementation Strategy**: Add PostInitialize call after cache client creation

---

### T008 Create unit tests for PostInitialize method in internal/cache/client_test.go
- [ ] T008 Write comprehensive test suite for cache initialization hook
   - **Story Goal**: Validate PostInitialize functionality through automated testing
   - **Independent Test Criteria**:
     - Test "PostInitialize successfully caches N records" mocks sqlx.DB returning expected rows, verifies exactly N calls to `cache.Set`, validates keys follow `entity_type:name` pattern (e.g., `food-category:vegetarian`)
     - Test "PostInitialize handles database errors gracefully" mocks query failure, verifies error returned and non-fatal behavior allows startup
     - Test "Tenant-specific isolation verified" creates Category A for tenant_1, Category B for tenant_2, verifies each tenant's cache contains only their own categories
     - All tests run without panic and pass assertion checks
   - **Dependencies**: T006 (implementation must exist)
   - **Parallel Execution**: Sequential after T006 completes
   - **Implementation Strategy**: Create table-driven tests covering success paths, error handling, and tenant isolation

---

### T009 Update config.yaml with food_category TTL override and cache initialization configuration
- [ ] T009 Extend `config/config.yaml` to include food_category TTL in ttl_overrides section
   - **Story Goal**: Enable per-entity TTL configuration for cached food categories
   - **Independent Test Criteria**: 
     - config.yaml includes `food_category: "30m"` (or other appropriate duration) under `cache.ttls:`
     - Default TTL of 30 minutes applies to all food category entries
     - Configuration loads correctly through viper.GetString() pattern
   - **Dependencies**: T007 (PostInitialize implementation must exist first)
   - **Parallel Execution**: Sequential after T007 completes
   - **Implementation Strategy**: Add `food_category: "30m"` entry to ttls section

---

## Phase 3: User Story 1 - Cache Hit Rate Optimization (P1)

### Task Overview
Ensure the cache initialization achieves maximum efficiency for food category lookups, enabling zero database queries during application runtime.

**Story Goal**: Achieve 100% cache hit rate for all food category access patterns with semantic search capability via `entity_type:name` key format.

**Independent Test Criteria**: 
- Zero database queries during application runtime for food category lookups
- Semantic search functionality: retrieve all categories by matching name pattern (e.g., find all vegetarian items)
- All cache operations use `food-category:<name>` keys and return complete FoodCategory objects immediately
- Application startup completes within acceptable time bounds (< 5s for datasets up to 1000 records per tenant)

---

### T010 [P1] Implement semantic search pattern for cached food categories in handlers
- [ ] T010 Add semantic search capability using `entity_type:name` cache keys in `internal/handlers/fooditem.go`
   - **Story Goal**: Enable filtering and searching of cached food categories by descriptive name attributes
   - **Independent Test Criteria**: 
     - Handler can retrieve category by name key: `cache.Get(ctx, "food-category:vegetarian")` returns `*models.FoodCategory` immediately from cache
     - Multiple semantic queries work: `food-category:dairy-free`, `food-category:gluten-free`, `food-category:jain-diet` all return cached values
     - No database round-trips when serving categorization data for food items
   - **Dependencies**: T008 (unit tests must pass before feature usage)
   - **Parallel Execution**: Sequential after T008
   - **Implementation Strategy**: Update handlers to use `entity_type:name` cache keys

---

### T011 [P1] Validate cache hits work correctly after initialization in integration tests
- [ ] T011 Create integration test scenario validating PostInitialize functionality and cache hit behavior
   - **Story Goal**: Verify end-to-end cache initialization achieves expected cache hit rate metrics
   - **Independent Test Criteria**: 
     - Start application, trigger HTTP request for food items
     - Observe zero database queries in logs for category resolution
     - Response includes properly categorized items with instant latency (cache hit)
     - Cache statistics confirm all requested categories were served from cache
   - **Dependencies**: T010 (semantic search implementation must exist)
   - **Parallel Execution**: Sequential after T010 completes
   - **Implementation Strategy**: Use httptest or integration test framework to simulate startup scenario and verify no database access for cached lookups

---

## Phase 4: User Story 2 - Tenant-Scoped Cache Isolation (P2)

### Task Overview
Ensure each tenant's cache instance contains only their own food category data, preventing cross-tenant data leakage.

**Story Goal**: Maintain strict tenant isolation in cache population with per-tenant database queries and separate cache instances.

**Independent Test Criteria**: 
- Tenant 1 cache contains only categories created for tenant_1
- Tenant 2 cache contains only categories created for tenant_2
- Cross-tenant queries return empty results (not data from other tenants)
- Row-Level Security policies enforced at database query level before cache population

---

### T012 [P2] Implement per-tenant PostInitialize method that isolates cache by tenant ID
- [ ] T012 Extend `PostInitialize` to accept tenantID parameter and scope database queries per tenant
   - **Story Goal**: Ensure each tenant's cache instance populates only their own food categories
   - **Independent Test Criteria**: 
     - Method signature: `PostInitialize(ctx context.Context, tenantID string, db *sqlx.DB) error` (tenantID param added)
     - Database query scopes to specific tenant: `WHERE tenant_id = $1` parameterized query
     - Tenant A's cache does not contain Tenant B's categories even if both tenants exist in same database
   - **Dependencies**: T009 (config update must include TTL override)
   - **Parallel Execution**: Sequential after T009 completes
   - **Implementation Strategy**: Update method signature to include tenantID parameter and ensure WHERE clause properly scopes queries

---

### T013 [P2] Update main.go to call PostInitialize with tenant context for each Server instance
- [ ] T013 Modify `cmd/main.go` bootstrap to pass tenant ID when initializing cache per tenant server instance
   - **Story Goal**: Wire up tenant-specific cache population during multi-tenant application startup
   - **Independent Test Criteria**: 
     - Each Server instance receives correct tenantID in PostInitialize call
     - Tenant isolation maintained: Server A's cache initialized with query scoped to tenant_A, Server B's cache with query scoped to tenant_B
   - **Dependencies**: T012 (per-tenant method signature must exist)
   - **Parallel Execution**: Sequential after T012 completes
   - **Implementation Strategy**: Update main.go to pass tenantID to PostInitialize for each tenant

---

### T014 [P2] Add tenant isolation unit tests for PostInitialize method
- [ ] T014 Write test validating cross-tenant data leakage does not occur in cache
   - **Story Goal**: Prevent information disclosure between tenants through cache population logic
   - **Independent Test Criteria**: 
     - Create Category A for tenant_1, Category B for tenant_2 using mocked database
     - Verify tenant_1's cache contains only A (not B)
     - Verify tenant_2's cache contains only B (not A)
     - Mock SQL query shows WHERE tenant_id = $1 parameter with correct tenant ID bound
   - **Dependencies**: T013 (per-tenant call pattern must exist)
   - **Parallel Execution**: Sequential after T013 completes
   - **Implementation Strategy**: Create unit tests validating tenant isolation

---

## Phase 5: Polish & Cross-Cutting Concerns

### Task Overview
Complete remaining implementation details including configuration updates and documentation.

---

### T015 Update config.yaml with food_category TTL override
- [ ] T015 Add `food_category` entry to `cache.ttls:` section in configuration file
   - **Story Goal**: Enable configurable TTL for all cached food category entries
   - **Independent Test Criteria**: 
     - Config file contains: `food_category: "30m"` (default TTL) or higher
     - LoadConfig() function picks up TTL override through viper.GetString("cache.ttls.food_category")
     - PostInitialize uses overridden TTL when storing categories
   - **Dependencies**: T014 (tenant isolation tests must pass)
   - **Parallel Execution**: Sequential after T014
   - **Implementation Strategy**: Add `food_category: "30m"` entry to ttls section in config.yaml

---

### T016 Create quickstart guide for developers using cache initialization API
- [ ] T016 Document usage pattern in docs/specs/cache_implementation/quickstart.md
   - **Story Goal**: Help new developers understand and use PostInitialize hook correctly
   - **Independent Test Criteria**: 
     - Document explains how to call PostInitialize at startup with tenant context
     - Example code shows calling pattern with database connection injection
     - Explains `entity_type:name` key convention for semantic search capability
   - **Dependencies**: T015 (config update must exist)
   - **Parallel Execution**: Sequential after T015 completes
   - **Implementation Strategy**: Create comprehensive documentation with examples

---

## Summary

| Phase | Task Count | Priority | Status |
|-------|------------|----------|--------|
| T001-T003 Setup | 3 | - | In Progress |
| T004-T009 Foundational | 6 | P0 (blocking) | Pending |
| T010-T011 User Story 1 - Cache Hit Rate | 2 | P1 | Pending |
| T012-T014 User Story 2 - Tenant Isolation | 3 | P2 | Pending |
| T015-T016 Polish | 2 | - | Pending |
| **TOTAL** | **16** | - | - |

---

## Implementation Strategy

### MVP Scope (Phase 4, T004-T009)
The minimal implementable solution includes:
- Extended Client interface with PostInitialize method signature
- PostInitialize implementation in cacheImpl using sqlx database querying
- Application startup wiring to trigger cache initialization
- Config update with food_category TTL override

### Completion Order
1. Setup tasks (T001-T003) - can run in parallel
2. Foundational tasks (T004-T009) - sequential dependency chain, must complete before user stories
3. User Story 1 tasks (T010-T011) - cache hit rate optimization
4. User Story 2 tasks (T012-T014) - tenant isolation verification
5. Polish tasks (T015-T016) - documentation and configuration finalization

---

## Dependencies Graph

```
Phase 1: Setup
└── T001, T002, T003 (parallel execution)

Phase 2: Foundational (blocking for all phases)
┌─────────┬────────────┬──────────┐
│ T004     │ → T006      │ → T007    │
│          │             │           │
│       ───┼────────────┤           │
│ T005     │ ← ──────────┼ ← ─────┼─────────┐
│          └──────────────────┴────┘         │
                                          (blocks all user stories)

Phase 3: P1 User Story (Cache Hit Rate)
├── T010 → T011
│   Dependency: PostInitialize working correctly (T007, T008)

Phase 4: P2 User Story (Tenant Isolation)
├── T012 → T013 → T014
│   Dependency: Per-tenant method signature (T012), wiring (T013), tests (T014)

Phase 5: Polish
├── T015, T016 (final documentation and configuration updates)
```

---

## Test Scenarios Summary

| Scenario | Task ID | Priority | Independent Test Criteria |
|----------|---------|----------|---------------------------|
| Cache initialization works correctly | T008 | P0 | Mock DB returns expected rows, verifies N Set calls with correct key format `entity_type:name` |
| Database errors handled gracefully | T008 | P0 | Query failure returns error but allows app startup (non-fatal) |
| Cache hit rate = 100% for categories | T011 | P1 | Zero DB queries after initialization, all category lookups instant |
| Tenant A cannot see Tenant B's data | T014 | P2 | Cross-tenant queries return empty; WHERE clause parameterized correctly |
| Configuration applies TTL override | T015 | - | viper.GetString("cache.ttls.food_category") returns configured value |
