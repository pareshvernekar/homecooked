# Research: Cache Initialization Patterns for Go Applications

**Date**: 2026-05-18  
**Feature**: cache-initialization-from-food-categories  
**Branch**: `002-cache-initialization-from-food-categories`

---

## Research Summary

This document consolidates findings from research on database-driven cache initialization at application startup, specifically for the food menu management system's reference data (food categories).

---

## 1. Database-Driven Cache Population Patterns

### Decision: Implement PostInitialize pattern with sqlx querying

**Rationale**: 
The existing eko/gocache library provides a clean interface for adding post-initialization hooks. The Go ecosystem supports this well through the standard database/sql abstraction, and sqlx already exists in the project for ORM operations.

**Alternatives Considered**:
- **Redis key-value stores**: Already implemented as cache fallback, but would require startup sync from PostgreSQL; adds Redis network overhead at boot.
- **Application-level sync map with lazy loading**: Simpler but defeats cache purpose during high concurrency; database pressure persists.
- **Connection pool integration**: Leverage existing DB connection lifecycle to trigger cache warmup naturally; clean separation of concerns.

**Selected Approach**: PostInitialize hook that executes sqlx queries after cache client initialization, following the pattern:
```go
func (c *cacheImpl) PostInitialize(ctx context.Context, db *sqlx.DB) error {
     // Query DB and populate cache entries
}
```

---

## 2. Thread-Safety for Bootstrap Hooks

### Decision: Use existing sync.RWMutex in cacheImpl, execute PostInitialize after write lock release

**Rationale**: 
The cache client already uses mutex for thread safety. Post-initialization should occur after the initial setup is complete, ensuring consistency across all operations that follow.

**Alternatives Considered**:
- **Copy of mutex specifically for bootstrap**: Redundant; existing locks cover this case.
- **Async goroutine for initialization**: Risky for startup sequence; main thread needs cache ready.
- **Channel-based signaling**: Overkill for single-threaded bootstrap operation.

---

## 3. Tenant-Scoped Cache Population with RLS

### Decision: Iterate through tenants, populate per-tenant caches independently

**Rationale**: 
PostgreSQL Row-Level Security (RLS) ensures each tenant can only see their own data. Cache must mirror this isolation since the underlying store (go-cache) doesn't natively enforce RLS semantics.

**Alternatives Considered**:
- **Single shared cache with tenant context in key**: Would require rewriting every handler to set context before DB queries; violates existing middleware pattern.
- **Per-tenant cache instance keyed by tenant ID**: Already architected this way via dependency injection per tenant; just need to wire up initialization logic.
- **Hybrid approach**: Critical tenant data in shared cache, reference data partitioned; adds complexity for marginal benefit since all entity lookups are tenant-scoped anyway.

**Why It Works**: The Server struct is instantiated per tenant (one instance = one cache). PostInitialize receives the specific tenant context and can execute tenant-specific queries against the shared connection pool.

---

## 4. Integration Pattern: Adding PostInitialize Without Breaking API

### Decision: Extend Client interface minimally; maintain existing Get/Set/Has methods

**Rationale**: 
The dependency injection pattern relies on `Client` interface compliance. Adding a post-init method requires interface evolution but can be done backwards-compatible via functional options pattern if needed.

**Alternatives Considered**:
- **Functional options wrapper**: Creates abstraction layer between interface and implementation; adds indirection for zero-cost inline hook.
- **Second-level interface (PostInit) passed alongside**: Maintains clean separation but complicates constructor signature.
- **Interface extension in v2 pattern**: Project is early version; breaking change acceptable for cleaner design.

**Chosen**: Direct interface extension. The main server bootstrap happens once, so no concurrent initialization conflicts exist:
```go
type Client interface {
    Get(ctx context.Context, key string) (any, error)
    Set(ctx context.Context, key string, value any, options ...SetOption) error
    Has(key string) bool
    PostInitialize(ctx context.Context, db *sqlx.DB) error
}

func NewCacheClient(config CacheConfig) (Client, error) {
     // ... initialization
     return cache, nil
}

func (c *cacheImpl) PostInitialize(ctx context.Context, db *sqlx.DB) error {
     // Bootstrap routine implementation
}
```

---

## 5. Configuration Integration for TTL Overrides

### Decision: Extend TTLOverrides with food_category entry using existing viper config pattern

**Rationale**: 
The project uses viper for configuration loading with YAML files. Adding a new TTL override requires no code changes beyond viper.GetString() parsing.

**Alternatives Considered**:
- **Env-based TTL overrides**: Inconsistent with existing yaml config pattern in `/config/config.yaml`.
- **Command-line flags only**: Less flexible than YAML; harder to update without rebuild.
- **Hardcoded default per entity type**: Reduces configurability; goes against documented flexibility.

**Chosen**: Extend `cache.ttls` section in config.yaml with `food_category: "30m"` (or other duration). TTL is applied during PostInitialize using the same pattern as runtime Set operations.

---

## Key Takeaways

1. **PostInitialize hook pattern** fits existing eko/gocache implementation cleanly
2. **Thread-safety** leveraged from existing sync.RWMutex; no new synchronization needed
3. **Tenant isolation** naturally supported by per-instance cache; RLS queries scoped to tenant_id
4. **Minimal API breaking change**; interface extension acceptable at current version
5. **Configuration extensibility** already in place via viper + YAML pattern

---

## Next Steps

- Proceed to Phase 1: Define FoodCategory model structure, finalize `entity_type:name` key convention documentation
- Implement PostInitialize method using sqlx queries and TTL override from config
- Wire up cache initialization at application startup (main.go) after NewCacheClient call

