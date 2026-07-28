# Food Menu System Implementation Tasks

## Overview
This document outlines the implementation tasks for the Food Menu System, a REST API for managing food catalogs, weekly menus, catering menus, orders, and notifications with multi-tenancy support.

**Related Documents**: 
- [Traceability Matrix](TRACEABILITY.md) - Maps tasks to code files and tests
- [Specification](spec.md) - Feature requirements
- [Implementation Plan](plan.md) - Design details

## Phase 1: Setup Phase
### Phase 1.1: Project Initialization
- [X] T001 Create project structure as defined in plan.md
  - **Implementation**: Directory structure created
  - **Code Reference**: [cmd/main.go](cmd/main.go), [internal/](../../internal/)
  - **Test Reference**: None required
  - ✓ Created internal/models, internal/validation, internal/repository, internal/services, internal/middleware
  - ✓ Created cmd/, pkg/, migrations/, scripts/, tests/, docs/, config/, logs/ directories
  - ✓ Created internal/api, internal/config, internal/database, internal/utils, internal/tenant directories

- [X] T002 Initialize Git repository
  - **Implementation**: [.gitignore](../../.gitignore), [README.md](../../README.md)
  - **Code Reference**: Git repository initialized
  - **Test Reference**: None required
  - ✓ Created .gitignore file with Go and project-specific rules
  - ✓ Set up initial commit with README.md

- [X] T003 Set up Go environment
  - **Implementation**: [go.mod](../../go.mod), [go.sum](../../go.sum)
  - **Code Reference**: [Makefile](../../Makefile), [.golangci.yml](../../.golangci.yml)
  - **Test Reference**: None required
  - ✓ Created project structure as defined in plan.md
  - ✓ Initialized Git repository with .gitignore
  - ✓ Created README.md with project overview
  - ✓ Installed Go tools (golangci-lint, delve)
  - ✓ Configured Go modules and dependency management
  - ✓ Created basic project directories (cmd, internal, migrations, tests, etc.)

### Phase 1.2: Environment Setup
- [X] T004 Install required Go packages
  - **Implementation**: [go.mod](../../go.mod) - Dependencies: gin-gonic/gin, go-sql-driver/mysql, jmoiron/sqlx
  - **Code Reference**: All packages compiled and linked
  - **Test Reference**: None required
  - ✓ Added gin-gonic/gin, go-sql-driver/mysql, jmoiron/sqlx to go.mod
  - ✓ Installed other necessary dependencies

- [✅] T005 Create configuration files
  - **Implementation**: [internal/config/config.go](../../internal/config/config.go)
  - **Code Reference**: [.env](../../.env) - Environment variables template
  - **Test Reference**: [internal/config/config_test.go](../../internal/config/config_test.go)
  - ✓ Created .env file with tenant-specific settings
  - ✓ Implemented configuration management for environment variables

- [✅] T006 Implement tenant context middleware
  - **Implementation**: [internal/middleware/middleware.go](../../internal/middleware/middleware.go)
  - **Code Reference**: Extracts tenant ID from X-Tenant-ID header
  - **Test Reference**: [internal/middleware/middleware_test.go](../../internal/middleware/middleware_test.go)
  - ✓ Created middleware to extract tenant ID from headers
  - ✓ Implemented tenant context for request handling

### Phase 1.3: Build and Lint Setup

- [✅] T007 Create Makefile for build and run tasks
  - **Implementation**: [Makefile](../../Makefile)
  - **Code Reference**: build, run, test, migrate, lint targets
  - **Test Reference**: None required
  - ✓ Created Makefile with build, run, test, migrate, and lint targets
  - ✓ Added commands for local development and deployment

- [✅] T008 Set up linting with golangci-lint
  - **Implementation**: [.golangci.yml](../../.golangci.yml)
  - **Code Reference**: Configured linter with errcheck, staticcheck rules
  - **Test Reference**: None required (linting itself is verification)
  - ✓ Installed golangci-lint
  - ✓ Configured .golangci.yml with project-specific linting rules (errcheck, staticcheck)
  - ✓ Added lint target to Makefile

## Phase 2: Database Setup Phase
### Phase 2.1: Database Schema Implementation
- [ ] T011 Create database tables with tenant_id columns
  - **Implementation**: PENDING - Migration files needed in `migrations/`
  - **Code Reference**: Should implement tables: tenant, user, food_item, category, weekly_menu, catering_menu, menu_item, order, order_item, notification
  - **Test Reference**: Should include SQL schema tests
  - **Spec Reference**: [spec.md](spec.md#database-schema)
  - Implement tenants, food_item, categories, menus, menu_items, orders, order_items, notifications tables
  - Add tenant_id column to all tables as foreign key

- [ ] T012 Set up row-level security policies
  - **Implementation**: PENDING - RLS SQL scripts in `scripts/`
  - **Code Reference**: PostgreSQL RLS policies for tenant isolation
  - **Test Reference**: RLS verification tests
  - **Spec Reference**: [spec.md](spec.md#tenant-model)
  - Configure PostgreSQL row-level security for tenant isolation
  - Implement database-level permissions for tenant data

### Phase 2.2: ORM Setup
- [ ] T013 Set up database connection with tenant context
  - **Implementation**: PARTIAL - [internal/database/database.go](../../internal/database/database.go) exists
  - **Code Reference**: Database connection pooling with tenant context
  - **Test Reference**: [internal/database/database_test.go](../../internal/database/database_test.go)
  - **Status**: Needs tenant context implementation
  - Configure database connection pool with tenant-specific settings
  - Implement connection pooling with tenant isolation

- [ ] T014 Implement repository pattern
  - **Implementation**: PARTIAL - FoodItem repository only
  - **Code Reference**: [internal/repository/fooditem_repository.go](../../internal/repository/fooditem_repository.go)
  - **Test Reference**: [internal/repository/fooditem_repository_test.go](../../internal/repository/fooditem_repository_test.go)
  - **Missing**: Category, User, Menu, MenuItem, Order, OrderItem, Notification repositories
  - Create tenant-aware repository interfaces and implementations
  - Implement CRUD operations for all entities

### Phase 2.3: Migrations
- [ ] T015 Create migration scripts
  - **Implementation**: PENDING - Migration runner needed
  - **Code Reference**: `migrations/` directory with numbered migration files
  - **Test Reference**: Migration verification tests
  - Create migration scripts for schema changes with tenant support
  - Set up golang-migrate for migration management

- [ ] T016 Apply initial migrations
  - **Implementation**: PENDING - Initial migration runner
  - **Code Reference**: Migration execution in main.go
  - **Test Reference**: Migration success verification
  - **Dependency**: Requires T015
  - Apply initial migrations with tenant-specific data
  - Create tenant management tables (tenants, tenant_roles)

## Phase 3: Data Models and Validation Phase
### Phase 3.1: Data Models Implementation
- [ ] T021 Define Go structs for all entities
  - **Implementation**: PARTIAL - FoodItem only
  - **Code Reference**: [internal/models/fooditem.go](../../internal/models/fooditem.go)
  - **Test Reference**: None
  - **Missing**: Tenant, User, Category, WeeklyMenu, CateringMenu, MenuItem, Order, OrderItem, Notification models
  - **Spec Reference**: [spec.md](spec.md#data-models)
  - Implement models for Tenant, FoodItem, Category, Menu, MenuItem, Order, OrderItem, Notification
  - Define database model mappings

- [ ] T022 Create DTOs for data transfer
  - **Implementation**: PENDING - DTO file needed
  - **Code Reference**: Should be `internal/models/dtos.go` or similar
  - **Test Reference**: DTO serialization tests
  - **Spec Reference**: [spec.md](spec.md#data-models) - Entity JSON examples
  - Define Data Transfer Objects for API requests/responses
  - Implement serialization/deserialization logic

### Phase 3.2: Validation Setup
- [ ] T023 Set up validation for input data
  - **Implementation**: PENDING - Validation package needed
  - **Code Reference**: Should be `internal/validation/validator.go`
  - **Test Reference**: Validation rule tests
  - **Spec Reference**: [spec.md](spec.md#data-models) - Field requirements
  - Configure go-playground/validator
  - Define validation rules for each entity

- [ ] T024 Implement error handling
  - **Implementation**: PARTIAL - Response models exist
  - **Code Reference**: [internal/views/errorresponse.go](../../internal/views/errorresponse.go), [internal/views/successresponse.go](../../internal/views/successresponse.go)
  - **Test Reference**: None
  - **Spec Reference**: [api_specification.md](api_specification.md#error-responses)
  - Define error types and error handling strategy
  - Implement consistent error responses

## Phase 4: API Endpoints Implementation Phase
### Phase 4.1: API Design
- [✅] T031 Define API endpoints using REST conventions
   - **Implementation**: COMPLETE
   - **Code Reference**: [api_specification.md](api_specification.md) - Full API documentation
   - **Test Reference**: None (documentation)
   - ✓ Created comprehensive API documentation in `api_specification.md`
   - ✓ Defined request/response schemas for all entities (FoodItems, Categories, WeeklyMenus, CateringMenus, MenuItems, Orders, Notifications)
   - ✓ Documented authentication and authorization requirements
   - ✓ Specified error response formats and rate limiting

- [ ] T032 Create API contract specifications
  - **Implementation**: PENDING - Contract files needed
  - **Code Reference**: Should be `docs/specs/food_menu_system/contracts/`
  - **Test Reference**: Contract verification tests
  - **Spec Reference**: [api_specification.md](api_specification.md)
  - Document all endpoints with request/response examples
  - Define authentication and authorization requirements

### Phase 4.2: Route Configuration
- [ ] T033 Set up router with Gin
  - **Implementation**: PARTIAL - Basic server setup exists
  - **Code Reference**: [internal/server/server.go](../../internal/server/server.go), [cmd/main.go](../../cmd/main.go)
  - **Test Reference**: None
  - **Status**: Needs complete route configuration for all endpoints
  - Configure routes for all endpoints
  - Implement middleware for authentication, logging, etc.

- [ ] T034 Implement middleware stack
  - **Implementation**: PARTIAL - Tenant middleware complete
  - **Code Reference**: [internal/middleware/middleware.go](../../internal/middleware/middleware.go)
  - **Test Reference**: [internal/middleware/middleware_test.go](../../internal/middleware/middleware_test.go)
  - **Missing**: Request logging, error handling, CORS middleware
  - Create middleware for request logging, error handling
  - Implement tenant context middleware

### Phase 4.3: Endpoint Implementation
- [ ] T035 Implement Food Items endpoints
  - **Implementation**: PARTIAL - GET implemented
  - **Code Reference**: [internal/handlers/fooditem.go](../../internal/handlers/fooditem.go)
  - **Test Reference**: [internal/handlers/fooditem_test.go](../../internal/handlers/fooditem_test.go)
  - **Endpoints**: GET /api/food-items, POST /api/food-items, PUT /api/food-items/{id}, DELETE /api/food-items/{id}
  - **Spec Reference**: [api_specification.md](api_specification.md#food-items-endpoints)
  - CRUD operations for food items (/api/food-items)
  - Add validation and error handling

- [ ] T036 Implement Categories endpoints
  - **Implementation**: PENDING - Handler not created
  - **Code Reference**: Should be `internal/handlers/category.go`
  - **Test Reference**: Should be `internal/handlers/category_test.go`
  - **Endpoints**: GET /api/categories, POST /api/categories, PUT /api/categories/{id}, DELETE /api/categories/{id}
  - **Spec Reference**: [api_specification.md](api_specification.md)
  - CRUD operations for categories (/api/categories)
  - Add validation and error handling

- [ ] T037 Implement Weekly Menus endpoints
  - **Implementation**: PENDING - Handler not created
  - **Code Reference**: Should be `internal/handlers/weekly_menu.go`
  - **Test Reference**: Should be `internal/handlers/weekly_menu_test.go`
  - **Endpoints**: GET /api/weekly-menus, POST /api/weekly-menus, PUT /api/weekly-menus/{id}, DELETE /api/weekly-menus/{id}
  - **Spec Reference**: [api_specification.md](api_specification.md)
  - CRUD operations for weekly menus (/api/weekly-menus)
  - Add validation and error handling

- [ ] T038 Implement Catering Menus endpoints
  - **Implementation**: PENDING - Handler not created
  - **Code Reference**: Should be `internal/handlers/catering_menu.go`
  - **Test Reference**: Should be `internal/handlers/catering_menu_test.go`
  - **Endpoints**: GET /api/catering-menus, POST /api/catering-menus, PUT /api/catering-menus/{id}, DELETE /api/catering-menus/{id}
  - **Spec Reference**: [api_specification.md](api_specification.md)
  - CRUD operations for catering menus (/api/catering-menus)
  - Add validation and error handling

- [ ] T039 Implement Menu Items endpoints
  - **Implementation**: PENDING - Handler not created
  - **Code Reference**: Should be `internal/handlers/menu_item.go`
  - **Test Reference**: Should be `internal/handlers/menu_item_test.go`
  - **Endpoints**: GET /api/menus/{menuType}/menu-items, POST /api/menus/{menuType}/menu-items, etc.
  - **Spec Reference**: [api_specification.md](api_specification.md)
  - CRUD operations for menu items (/api/menus/{menuType}/menu-items)
  - Add validation and error handling

- [ ] T040 Implement Orders endpoints
  - **Implementation**: PENDING - Handler not created
  - **Code Reference**: Should be `internal/handlers/order.go`
  - **Test Reference**: Should be `internal/handlers/order_test.go`
  - **Endpoints**: GET /api/orders, POST /api/orders, PUT /api/orders/{id}, DELETE /api/orders/{id}
  - **Spec Reference**: [api_specification.md](api_specification.md)
  - CRUD operations for orders (/api/orders)
  - Add validation and error handling

- [ ] T041 Implement Notifications endpoints
  - **Implementation**: PENDING - Handler not created
  - **Code Reference**: Should be `internal/handlers/notification.go`
  - **Test Reference**: Should be `internal/handlers/notification_test.go`
  - **Endpoints**: GET /api/notifications, POST /api/notifications, PUT /api/notifications/{id}, DELETE /api/notifications/{id}
  - **Spec Reference**: [api_specification.md](api_specification.md)
  - CRUD operations for notifications (/api/notifications)
  - Add validation and error handling

## Phase 5: Business Logic and Services Phase
### Phase 5.1: Services Layer
- [ ] T051 Create service interfaces
  - **Implementation**: PENDING - Service interfaces needed
  - **Code Reference**: Should be `internal/services/interfaces.go` or similar
  - **Test Reference**: Mock service tests
  - **Spec Reference**: [spec.md](spec.md#core-features)
  - Define interfaces for FoodItemService, MenuService, OrderService, NotificationService

- [ ] T052 Implement business logic
  - **Implementation**: PENDING - Service implementations needed
  - **Code Reference**: Should be `internal/services/` directory
  - **Test Reference**: Service unit tests
  - **Spec Reference**: [plan.md](plan.md#phase-5-business-logic-and-services-phase)
  - Implement menu creation and management logic
  - Implement order processing logic
  - Implement notification logic

### Phase 5.2: Authentication and Authorization
- [ ] T053 Implement JWT-based authentication
  - **Implementation**: PENDING - Auth package needed
  - **Code Reference**: Should be `internal/auth/jwt.go`
  - **Test Reference**: JWT token tests
  - **Spec Reference**: [spec.md](spec.md#user-roles)
  - Set up JWT token generation and validation
  - Implement token refresh mechanism

- [ ] T054 Set up role-based access control
  - **Implementation**: PENDING - RBAC middleware needed
  - **Code Reference**: Should be `internal/middleware/rbac.go`
  - **Test Reference**: RBAC authorization tests
  - **Spec Reference**: [spec.md](spec.md#user-roles) - Admin and User roles
  - Define user roles and permissions
  - Implement role-based authorization

## Phase 6: Testing Phase
### Phase 6.1: Unit Testing
- [ ] T061 Write unit tests for services
  - **Implementation**: PENDING - Service tests needed
  - **Code Reference**: Should be `internal/services/*_test.go`
  - **Test Reference**: Service unit tests in `*_test.go` files
  - **Spec Reference**: [plan.md](plan.md#phase-5-testing) → Unit Testing
  - Test all service methods with mock dependencies
  - Ensure high test coverage

- [ ] T062 Write unit tests for handlers
  - **Implementation**: PENDING - Additional handler tests
  - **Code Reference**: Should be `internal/handlers/*_test.go`
  - **Test Reference**: Handler unit tests (partial: [fooditem_test.go](../../internal/handlers/fooditem_test.go) exists)
  - **Spec Reference**: [plan.md](plan.md#phase-5-testing) → Unit Testing
  - Test all API handlers with mock services
  - Ensure proper error handling

### Phase 6.2: Integration Testing
- [ ] T063 Test API endpoints with mock database
  - **Implementation**: PENDING - Integration test suite
  - **Code Reference**: Should be `tests/integration/` directory
  - **Test Reference**: Integration tests for all CRUD operations
  - **Spec Reference**: [plan.md](plan.md#phase-5-testing) → Integration Testing
  - Test all CRUD operations with test data
  - Verify database interactions

- [ ] T064 Test service integrations
  - **Implementation**: PENDING - Service integration tests
  - **Code Reference**: Should be `tests/integration/services/` directory
  - **Test Reference**: Service integration tests
  - **Spec Reference**: [plan.md](plan.md#phase-5-testing) → Integration Testing
  - Test business logic with mock repositories
  - Verify transaction handling

## Phase 7: Deployment and CI/CD Phase
### Phase 7.1: Docker Setup
- [ ] T071 Create Dockerfile for application
  - **Implementation**: PENDING - Dockerfile needed
  - **Code Reference**: `Dockerfile` at project root
  - **Test Reference**: Docker build and run tests
  - Define multi-stage build process
  - Configure environment variables

- [ ] T072 Create Dockerfile for database
  - **Implementation**: PENDING - Database Dockerfile
  - **Code Reference**: `Dockerfile.db`
  - **Test Reference**: Database container tests
  - Configure PostgreSQL/MySQL with tenant isolation
  - Set up database initialization scripts

- [ ] T073 Set up docker-compose.yml
  - **Implementation**: PARTIAL - File exists
  - **Code Reference**: [docker-compose.yml](../../docker-compose.yml)
  - **Test Reference**: Docker Compose verification
  - **Dependency**: Requires T071, T072
  - Configure application and database services
  - Define network and volume configurations

### Phase 7.2: CI/CD Pipeline
- [ ] T074 Set up CI pipeline
  - **Implementation**: PENDING - GitHub Actions workflow
  - **Code Reference**: `.github/workflows/ci.yml`
  - **Test Reference**: CI pipeline execution logs
  - Configure GitHub Actions for building and testing
  - Add test coverage checks

- [ ] T075 Set up CD pipeline
  - **Implementation**: PENDING - GitHub Actions deployment
  - **Code Reference**: `.github/workflows/cd.yml`
  - **Test Reference**: CD pipeline execution logs
  - **Dependency**: Requires T074
  - Configure deployment to cloud provider
  - Set up environment-specific configurations

## Phase 8: Monitoring and Maintenance
### Phase 8.1: Monitoring Setup
- [ ] T081 Set up application monitoring
  - **Implementation**: PENDING - Monitoring package
  - **Code Reference**: Should include Prometheus metrics in handlers/services
  - **Test Reference**: Metrics collection tests
  - Configure Prometheus for metrics collection
  - Set up Grafana dashboards

- [ ] T082 Configure monitoring alerts
  - **Implementation**: PENDING - Alert configuration
  - **Code Reference**: Alert rules configuration
  - **Test Reference**: Alert trigger tests
  - Set up alerts for critical issues
  - Define alert thresholds

### Phase 8.2: Logging Setup
- [ ] T083 Set up centralized logging
  - **Implementation**: PARTIAL - Logger exists
  - **Code Reference**: [internal/logger/logger.go](../../internal/logger/logger.go)
  - **Test Reference**: Logging output tests
  - Configure ELK Stack or CloudWatch
  - Set up log retention policies

### Phase 8.3: Security Audits
- [ ] T084 Regular security audits
  - **Implementation**: PENDING - Security audit procedures
  - **Code Reference**: Security policies and procedures
  - **Test Reference**: Security testing checklist
  - Conduct penetration testing
  - Update security policies

## Dependencies
- [ ] T999 Install all required dependencies
  - **Implementation**: [go.mod](../../go.mod), [go.sum](../../go.sum)
  - Install Go tools and packages
  - Set up development environment

## Testing
- [ ] T998 Run all tests
  - **Implementation**: [Makefile](../../Makefile) - test target
  - Execute unit tests
  - Execute integration tests
  - Verify test coverage

## Documentation
- [ ] T997 Document deployment procedures
  - **Implementation**: Should be in `docs/deployment.md`
  - Create deployment guides
  - Document API documentation
  - Document troubleshooting steps

## User Stories
### Admin Features
- [ ] T991 Create and manage food items
- [ ] T992 Create weekly menus
- [ ] T993 Create catering menus
- [ ] T994 Manage menu items
- [ ] T995 Process orders
- [ ] T996 Send notifications

### User Features
- [ ] T999 View menus
- [ ] T990 Place orders
- [ ] T995 Receive notifications

## Parallel Tasks
- [ ] T999 Implement error handling across all endpoints
- [ ] T998 Set up logging throughout the application
- [ ] T997 Implement consistent error responses

## Dependencies Between Tasks
- Database schema implementation (T011) must be completed before ORM setup (T013)
- ORM setup (T013) must be completed before data models implementation (T021)
- Data models implementation (T021) must be completed before API endpoints implementation (T035-T041)
- Authentication setup (T053-T054) must be completed before full API implementation

## Notes
- All tasks should include proper error handling and validation
- Follow Go best practices for code organization and style
- Ensure all code follows the project's coding standards
- Implement proper logging throughout the application

## References
- Database schema: [database/database_schema.md](database/database_schema.md)
- API contracts: [spec.md](spec.md)
- Project plan: [plan.md](plan.md)
- **Traceability Matrix**: [TRACEABILITY.md](TRACEABILITY.md) - Complete mapping of tasks to code files and tests

---

## Tag System and Traceability Guide

### Task ID Format
Each task follows the format: `T###` (e.g., T001, T035, T999)
- Task IDs are sequential within phases
- Ranges: 001-008 (Phase 1), 011-016 (Phase 2), 021-024 (Phase 3), 031-041 (Phase 4), 051-054 (Phase 5), 061-064 (Phase 6), 071-075 (Phase 7), 081-084 (Phase 8)

### Implementation and Test File References
Each task includes:
1. **Status**: Completion status (✅ Complete, 🔄 Partial, ❌ Not Started)
2. **Implementation**: Code files that implement the task
3. **Code Reference**: Specific functions/features in those files
4. **Test Reference**: Files containing tests for this task
5. **Spec Reference**: Links to relevant specification documents

### Cross-Referencing
- **By Task**: Use the Traceability Matrix to find code files for a specific task
- **By File**: Check the file-to-task table in Traceability Matrix
- **By Feature**: Refer to spec.md to find all related tasks

### Traceability Usage Examples

#### Example 1: I'm implementing T035 (Food Items endpoints)
1. Look up T035 in this document
2. Check "Code Reference": [internal/handlers/fooditem.go](../../internal/handlers/fooditem.go)
3. Review "Test Reference": [internal/handlers/fooditem_test.go](../../internal/handlers/fooditem_test.go)
4. Review "Spec Reference": [api_specification.md](api_specification.md#food-items-endpoints)
5. Dependencies: T014, T021, T023, T033, T034 must be done first

#### Example 2: I'm reviewing fooditem.go
1. Open [TRACEABILITY.md](TRACEABILITY.md)
2. Look at "File-to-Task Cross-Reference" table
3. Find row with `fooditem.go`
4. See which tasks it implements: T035 (endpoints), T014 (repository)
5. Check implementation status in related tasks

#### Example 3: I found a bug in fooditem_repository.go
1. Task: T014 (Implement repository pattern)
2. Test file: [fooditem_repository_test.go](../../internal/repository/fooditem_repository_test.go)
3. Check failing test and fix accordingly
4. Update test file to reflect new behavior

### Implementation Status Symbols
- ✅ **COMPLETE**: Fully implemented and tested
- 🔄 **PARTIAL**: Partially implemented or some components missing
- ❌ **NOT STARTED**: No implementation yet

### Dependency Chain
When working on a task:
1. Check "Blocking" dependencies - must be completed before starting this task
2. Check "Dependency" - check if required tasks are done
3. Use Traceability Matrix to identify critical path

### Quality Metrics
- **Overall Progress**: 8/75 tasks complete (11%)
- **With Good Test Coverage**: 4 tasks (config, middleware, database, repository)
- **Partial Implementation**: 5 tasks (fooditem model, handlers, error handling, server setup)
- **Not Started**: 58 tasks (77%)

### Best Practices for Maintaining Traceability

1. **When Starting New Work**:
   - Create task ID (next available T###)
   - Add to tasks.md with description
   - Create corresponding implementation file
   - Create corresponding test file
   - Link both in this document

2. **When Finishing a Task**:
   - Update status (✅ COMPLETE or 🔄 PARTIAL)
   - Link implementation files
   - Link test files
   - Update TRACEABILITY.md

3. **When Reviewing Code**:
   - Verify task ID is referenced
   - Check corresponding test file exists
   - Verify test coverage is adequate
   - Update status if needed

4. **Code Comments**:
   - Add `// T###: [description]` comment at top of implementation files
   - Add `// T###_TEST: [description]` comment at top of test files
   - Example: `// T006: Tenant context middleware` in middleware.go

---