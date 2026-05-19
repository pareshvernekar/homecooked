# Implementation Plan: cache-initialization-food-categories

**Branch**: `002-cache-initialization-from-food-categories` | **Date**: 2026-05-18 | **Spec**: [spec.md](../002-cache-initialization-from-food-categories/spec.md)
**Input**: Feature specification for initializing in-memory cache with food category data from database at application startup

---

## Summary

**Primary Requirement**: Automatically populate the in-memory cache with all food category records from the `food_category` table when the application starts up.

**Technical Approach**: Implement a post-initialization bootstrap method on the cache client (`PostInitialize`) that connects to the database, retrieves all food category records per tenant, and stores them in the cache using `entity_type:id` key convention (e.g., `food-category:fc_1`). This ensures reference data is available immediately for downstream operations without database round-trips during runtime. Uses existing `FoodCategory` model from `models/fooditem.go` to maintain consistency with domain types.

---

## Technical Context

**Language/Version**: Go 1.26.2  
**Primary Dependencies**: 
- eko/gocache (in-memory cache store)
- go-cache library (gocachepatrickmn/go-cache)
- sqlx (database querying)
- viper (configuration management)

**Storage**: PostgreSQL with Row-Level Security (RLS) for tenant isolation  
**Testing**: go.test, table-driven tests, integration tests with real database

**Target Platform**: Linux server / Docker containerized deployment  
**Project Type**: web-service (REST API backend)  

**Performance Goals**: 
- Cache initialization completes within 5 seconds for datasets up to 1000 records per tenant
- Zero database queries during application runtime for food category lookups
- Startup time increase < 10% of total boot time

**Constraints**: 
- Tenant-scoped operations due to Row-Level Security (RLS) in PostgreSQL
- Cache key format must be `entity_type:name` for semantic search (e.g., `food-category:vegetarian`, `food-category:dairy-free`) to enable lookups by descriptive name rather than opaque IDs
- Must integrate with existing cache client interface without breaking dependency injection pattern
- TTL configuration from application config (default 30 minutes global, can be overridden per entity)

**Scale/Scope**: Multi-tenant food menu management system; each tenant has isolated cache instance; expected ~10-50 food categories per tenant initially

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Gates Determined:
1. **Singleton Pattern**: ✓ Required for bootstrap operation (single initialization)
2. **Interface Compatibility**: ✓ Must satisfy cache.Client interface (Get, Set, Has methods)
3. **Tenant Isolation**: ✓ Each tenant has isolated cache instance due to RLS
4. **Performance Boundaries**: ✓ < 5s init time, < 10% startup overhead

### Justifications:
- Bootstrap operation must be idempotent and thread-safe using sync.RWMutex pattern already present in cacheImpl
- Post-initialization hook preserves dependency injection while adding one-time setup capability
- Tenant isolation maintained through per-tenant cache instances

---

## Project Structure

### Documentation (this feature)

```text
specs/cache_implementation/
├── plan.md               # This file (/speckit.plan command output)
├── research.md           # Phase 0 output (/speckit.plan command)
├── data-model.md         # Phase 1 output (cache entry models, key conventions)
├── quickstart.md         # Phase 1 output (usage examples for developers)
├── contracts/            # Phase 1 output (cache API contracts)
│   └── cache-initialization-api.md
└── tasks.md              # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
homecooked/
├── cmd/
│   └── main.go                 # Application bootstrap entry point
├── internal/
│   ├── cache/
│   │   ├── client.go          # Cache implementation with PostInitialize hook
│   │   ├── config.go          # Cache configuration (CacheConfig)
│   │   └── client_test.go     # Unit tests for new PostInitialize method
│   ├── database/
│   │   └── connection.go      # Database connection pool
│   ├── handlers/
│   │   └── fooditem.go        # Handlers that will benefit from cached categories
│   ├── models/
│   │   └── fooditem.go        # FoodItem model referencing food_category
│   └── server/
│       └── server.go          # Server initialization
├── config/
│   └── configs/
│       └── config.yaml        # Cache configuration (global_ttl, max_items)
└── specs/cache_implementation/
    ├── plan.md                # This file
    ├── research.md            # Research findings
    ├── data-model.md          # Cache entry data models
    ├── quickstart.md          # Developer quick reference
    ├── contracts/             # API contracts
    └── tasks.md              # Implementation tasks (speckit.tasks output)
```

**Structure Decision**: Following existing architecture pattern with dependency injection. New `PostInitialize` method added to cache client interface and implemented in cacheImpl struct. Post-initialization logic will be wired up at application startup in main.go after cache client creation.

---

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| Per-tenant cache instances | Row-Level Security requires tenant-scoped database operations; cache must mirror this isolation to avoid data leakage between tenants | Single shared cache would violate RLS contract; each tenant query returns different results based on current_setting |

---

## Implementation Phases

### Phase 0: Research & Analysis

**Output**: `research.md` with findings on cache initialization patterns

**Research Tasks**:
1. Best practices for database-driven cache population at startup in Go applications
2. Thread-safety patterns for post-initialization bootstrap hooks
3. Tenant-scoped cache population strategies with RLS databases
4. Integration pattern for adding PostInitialize hook without breaking existing API

### Phase 1: Design & Contracts

**Output**: `data-model.md`, `contracts/cache-initialization-api.md`, `quickstart.md`

**Deliverables**:
1. Define CacheableFoodCategory model structure matching database schema
2. Define cache key convention documentation (`entity_type:name` format for semantic search capability)
3. Specify PostInitialize method signature and lifecycle
4. Create integration contract for cache initialization at bootstrap time
5. Update CLAUDE.md with plan reference

### Phase 2: Implementation Tasks

**Output**: `tasks.md` (generated by /speckit.tasks)

**Task Breakdown**:
1. Add PostInitialize method to Client interface
2. Implement PostInitialize in cacheImpl using sqlx database querying with entity_type:name key format
3. Wire up cache initialization at application startup in main.go after cache client creation
4. Create tenant-aware loop to initialize cache per tenant connection, storing categories indexed by name (e.g., "vegetarian", "dairy-free")
5. Add logging for observability (item count, errors)
6. Update config.yaml with food_category TTL override

---

## Success Metrics

- ✅ Cache contains all N food categories from database after startup
- ✅ Zero database queries during application runtime for cached entity types
- ✅ 100% cache hit rate for food category lookups post-initialization
- ✅ < 5 second initialization time for 1000-record dataset
- ✅ No breaking changes to existing cache API users

---

## Dependencies

- Database connection pool must be available at application startup
- Cache client interface already defined and implemented
- Application bootstrap code (main.go) identifies injection point for PostInitialize

---

## Next Steps

1. Run `/speckit.clarify` to validate requirements before Phase 1 design
2. Execute Phase 1 to produce data models and API contracts
3. Generate tasks with `/speckit.tasks` for implementation team
4. Review and execute implementation plan
