# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

# HomeCooked - Food Menu Management System

A REST API for managing food catalogs, weekly menus, catering menus, orders, and notifications with multi-tenancy support.

## Project Structure

```
homecooked/
├── cmd/                    # Application entry point
│   └── main.go            # Main application bootstrap
├── internal/              # Core application packages
│   ├── cache/           # Cache layer (Redis-based caching)
│   ├── config/          # Configuration management (Viper)
│   ├── database/        # Database connection & tenant context
│   ├── handlers/        # HTTP request handlers
│   ├── logger/          # Structured logging (slog)
│   ├── middleware/      # HTTP middleware (tenant isolation)
│       └── tenant/     # Tenant ID extraction middleware
│   ├── models/          # Domain models & DTOs
│   ├── repository/      # Data access layer
│   ├── server/          # HTTP server setup
│   ├── validation/      # Input validation
│   └── views/          # Response view models
├── migrations/         # Database migration scripts (empty - uses SQL files)
├── docs/
    └──specs/             # API & system specifications
│   ├── food_menu_system/     # Main API specification
│   └── cache_framework/      # Caching strategy spec
├── config/           # Application configuration
│   └── configs/     # Config YAML files (config.yaml)
├── docker-compose.yml  # Docker Compose for local dev
├── Makefile           # Build/test scripts
└── tests/            # Test directory
```

## Technology Stack

- **Language**: Go 1.26.2
- **Web Framework**: Gin (v1.12.0)
- **Database**: PostgreSQL with Row-Level Security (RLS)
- **ORM**: sqlx
- **Cache**: Redis-compatible interface (go-cache fallback via eko/gocache)
- **Config**: Viper
- **Logger**: slog (structured logging)

## Architecture Overview

### Multi-Tenancy Pattern

The application uses PostgreSQL Row-Level Security (RLS) for tenant isolation:

```sql
-- Tenant context set in session
SET app.current_tenant_id = :tenant_id;

-- All queries automatically filter by tenant
SELECT * FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1;
```

### Dependency Injection Flow

```
main.go (bootstrap)
    ↓
server.Server (router + DB + Logger)
    ↓
handlers.FoodItemHandler (business logic)
    ↓
repository.PostgreSQLFoodItemRepository (data access)
```

### Caching Strategy

Cache client supports hierarchical caching with TTL overrides:

- Default TTL: 30 minutes
- Max items per tenant: 1000
- Per-entity overrides in config (`cache.ttl_overrides`)

## Key Components

### 1. Server (`internal/server/server.go`)

```go
type Server struct {
    Router *gin.Engine
    DB     *sqlx.DB
    Logger  *logger.Logger
    Cache   cache.Client
}

// Routes configured with dependency injection
func SetupRoutes(router *gin.Engine, db, logger, handler, cache)
```

### 2. Tenant Middleware (`internal/middleware/tenant/middleware.go`)

Extracts tenant ID from `X-Tenant-ID` header and injects into request context:

```go
type TenantMiddleware func(c *gin.Context)

func TenantMiddleware(l *logger.Logger) TenantMiddleware {
    // Returns middleware that sets c.Set("tenant_id", ...)
}
```

### 3. Models (`internal/models/fooditem.go`)

FoodItem entity with tenant-scoped attributes:

```go
type FoodItem struct {
    ID              string       `json:"id"`
    TenantID        string       `json:"tenant_id"`  // RLS context
    Name            string       `json:"name"`
    Description     *string      `json:"description,omitempty"`
    Category        string       `json:"category"`
    Price           float64      `json:"price"`
    ImageURL        *string      `json:"image_url,omitempty"`
    Avoidance       *string      `json:"avoidance,omitempty"`
    IsVegetarian    bool         `json:"is_vegetarian"`
    AvailabilityStatus string     `json:"availability_status"`
    CreatedAt       *time.Time   `json:"created_at,omitempty"`
    UpdatedAt       *time.Time   `json:"updated_at,omitempty"`
}
```

### 4. Repository (`internal/repository/fooditem_repository.go`)

Interface-based repository pattern with PostgreSQL implementation:

```go
type FoodItemRepository interface {
    Create(*models.FoodItem) error
    GetByID(string) (*models.FoodItem, error)
    Update(*models.FoodItem) error
    Delete(string) error
    ListByTenant(string, int, int) ([]models.FoodItem, int64, error)
}

type PostgreSQLFoodItemRepository struct {
    DB      *sqlx.DB
    TenantID string
}
```

### 5. Handler (`internal/handlers/fooditem.go`)

Business logic with validation and caching:

```go
type FoodItemHandler struct {
    Repo        repository.FoodItemRepository
    Logger       *logger.Logger
    CacheClient cache.Client
}

// CRUD operations with validation
func (h *FoodItemHandler) CreateFoodItem(c *gin.Context)
func (h *FoodItemHandler) GetFoodItems(c *gin.Context)
func (h *FoodItemHandler) UpdateFoodItem(c *gin.Context)
func (h *FoodItemHandler) DeleteFoodItem(c *gin.Context)
```

## API Endpoints

### Food Items API
- `POST /api/v1/food-items` - Create food item
- `GET /api/v1/food-items` - List with pagination
- `PUT /api/v1/food-items/:id` - Update
- `DELETE /api/v1/food-items/:id` - Delete

### Categories API (placeholder)
- `POST /api/v1/categories` - Create category
- `GET /api/v1/categories` - List categories

### Weekly Menus API (placeholder)
- `POST /api/v1/weekly-menus` - Placeholder stubs

### Catering Menus API (placeholder)
- `POST /api/v1/catering-menus` - Placeholder stubs

### Orders API (placeholder)
- `POST /api/v1/orders` - Placeholder stubs

### Notifications API (placeholder)
- `POST /api/v1/notifications` - Placeholder stubs

## Common Patterns

### Error Responses

```go
// Standardized error response format
type ErrorResponse struct {
    Success   bool         `json:"success"`
    ErrorCode  string       `json:"error_code"`
    Message    string       `json:"message"`
    Detail     interface{}  `json:"details,omitempty"`
    Timestamp  time.Time    `json:"timestamp"`
}

// Predefined error codes in internal/views/errorresponse.go
const (
    INVALID_REQUEST           ErrorCode = "INVALID_REQUEST"
    VALIDATION_ERROR          ErrorCode = "VALIDATION_ERROR"
    NOT_FOUND                 ErrorCode = "NOT_FOUND"
    FORBIDDEN                 ErrorCode = "FORBIDDEN"
    DATABASE_ERROR            ErrorCode = "DATABASE_ERROR"
    // ... more codes in views/errorresponse.go
)
```

### Response Views (`internal/views/`)

- `errorresponse.go` - ErrorResponse, ValidationError types
- `successresponse.go` - SuccessResponse, PaginationInfo

## Database Schema

Key tables:
- `food_item` - Catalog items with tenant_id column (RLS)
- Categories table (placeholder)
- Weekly menus table (placeholder)
- Catering menus table (placeholder)
- Orders table (placeholder)
- Notifications table (placeholder)

### Row-Level Security Policy

```sql
CREATE POLICY tenant_isolation ON food_item
USING (current_setting('app.current_tenant_id'::text, true) = current_setting('app.current_tenant_id'::text, false));
```

## Configuration

### Application Config (`config/config.yaml`)
```yaml
cache:
  enabled: true
  global_ttl: "30m"
  max_items_per_tenant: 1000
  ttls:
    menu.weekly: "30m"
    order.menu_items: "15m"
```

### Environment Variables
- `TENANT_ID` - Tenant identifier (defaults to "1")
- Database connection via Viper

## Build & Run Commands

```bash
# Build
make build                    # Build Go binary
make run                      # Run compiled binary

# Development
make local                    # Run in dev mode with debug output
make server                   # Alias for make local

# Tests
make test                     # Run all tests + lint
make unit-test                # Unit tests only
make integration-test         # Integration tests (requires PostgreSQL)

# Lint
make lint                     # golangci-lint

# Database migrations
make migrate                  # Run all migrations
make migrate-up               # Apply migrations forward
make migrate-down             # Rollback one migration
make migrate-status           # Show migration status

# Docker
docker-compose -f docker-compose.prod.yml up --build    # Build & run
sleep 20                                                 # Wait for DB schema load
curl http://localhost:8080/health                       # Health check
```

## Testing Patterns

### Handler Tests (`internal/handlers/fooditem_test.go`)
- Mock repository pattern
- Mock cache client
- Gin test context with httptest
- Tenant ID injection via middleware simulation

```go
func TestUpdateFoodItem_Success(t *testing.T) {
    mockRepo := &MockFoodItemRepository{}
    handler := NewFoodItemHandler(mockRepo, logger, &MockCacheClient{})
    
    req := httptest.NewRequest(http.MethodPut, "/api/v1/food-items/:id", ...)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set(middleware.TenantIDKey, "tenant-1")  // Inject tenant ID
    
    handler.UpdateFoodItem(c)
}
```

### Repository Tests (`internal/repository/fooditem_repository_test.go`)
- `file=:memory:` SQLite database for integration testing
- Real PostgreSQL tests via initDBWithPostgres

## Notes & Implementation Status

- Current branch: `implementation-phase-4`
- Core FoodItem CRUD is implemented with validation and caching
- Other endpoints (categories, weekly menus, catering, orders, notifications) are placeholder stubs
- Caching framework built but not fully wired for all operations
- Multi-tenancy via PostgreSQL RLS

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
<!-- SPECKIT END -->

<!-- 
SPECKIT REFERENCE: docs/specs/cache_implementation/plan.md
Implementation Plan for cache-initialization-from-food-categories feature
Updated: 2026-05-18
-->
