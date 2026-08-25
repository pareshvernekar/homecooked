---

description: "Test suite implementation tasks - Gherkin BDD tests and TestContainers integration with PostgreSQL database verification"

**Input**: Feature Testing & Integration Testing Specification from `/docs/specs/003-feature-testing-integration/spec.md`  
**Prerequisites**: spec.md (required), plan.md (optional)

## Context

This task list implements comprehensive feature testing and integration testing for the Food Menu System REST API endpoints using:
- **BDD/Gherkin** for behavioral specification (go-ginkgo + gherkin-go)
- **TestContainers-go** for isolated PostgreSQL database instances per test run
- **Database state verification** alongside HTTP response validation

The implementation covers all existing FoodItem CRUD endpoints with >90% code coverage target.

---

## Phase 1: Setup (Shared Infrastructure) ✅ COMPLETE

**Purpose**: Project initialization and basic structure

### Task T001: Initialize Go modules for test suite dependencies

- [X] T001 Create go.mod entry for testcontainers-go dependency (github.com/testcontainers/testcontainers-go v0.43.x)
- [X] T002 Add gherkin.go package for BDD scenario parsing (github.com/cucumber/gherkin-go/v13)
- [X] T003 Initialize Go workspace configuration (go.work)

---

## Phase 2: Foundational Infrastructure ✅ IN PROGRESS

**Purpose**: Core testing infrastructure that MUST be complete before ANY feature tests can run

### Task T004: Create BDD/Gherkin framework setup

- [ ] T004 Create `tests/features/bdd/framework.go` with Gherkin scenario runner using go-ginkgo
- [ ] T005 Define step definitions pattern for Given/When/Then keywords in `tests/features/bdd/steps/`
- [ ] T006 Setup BDD test directory structure under `tests/features/bdd/features/`

### Task T007: Create TestContainers PostgreSQL client infrastructure

- [X] T007 Create database container factory (tests/db/postgres_container.go) using testcontainers-go
- [X] T008 Implement PostgreSQL connection helper with migration support in tests/db/db.go
- [ ] T009 Create container lifecycle manager (start/migrate/stop) for test isolation

### Task T011: Setup shared test fixtures and utilities

- [X] T011 Create tenant fixture generator (tests/fixtures/tenants.go)
- [ ] T012 Implement food category seeding utility (tests/fixtures/categories.go)
- [X] T013 Create test utilities for HTTP request/response handling in tests/http/test_helper.go

### Task T014: Configure CI/CD integration templates

- [ ] T014 Create GitHub Actions workflow template in `.github/workflows/test-suite.yml`
- [ ] T015 Define environment configuration for test containers (TEST_DB_USER, TEST_DB_PASSWORD)
- [ ] T016 Add documentation for running tests locally in docs/testing-guide.md

---

## Phase 3: User Story 1 - Feature Testing with BDD (Priority: P1) 🎯 MVP ⏳ PENDING

### Tests for User Story 1 (BDD/Gherkin Scenarios)

- [X] T017 Create Gherkin feature file for FoodItem creation
- [ ] T018 Add Gherkin scenario for FoodItem listing
- [ ] T019 Document FoodItem update validation rules
- [ ] T020 Create Gherkin scenarios for FoodItem deletion

### Implementation Tasks for User Story 1

- [ ] T021 Implement BDD step definition generator (tests/features/generator.go)
- [ ] T022 Create auto-converter utility (tests/features/converter.go)
- [ ] T023 Setup BDD command-line interface in `cmd/test-generate/main.go`

---

## Phase 4: User Story 2 - Integration Testing with TestContainers (Priority: P2) ⏳ IN PROGRESS

### Tests for User Story 2 (TestContainers Integration)

#### Performance Tests

- [ ] T024 Create FoodItem creation performance test in tests/performance/fooditems/perf_create.go
- [ ] T025 Implement FoodItem listing performance test in tests/performance/fooditems/perf_list.go
- [ ] T026 Write FoodItem update performance tests in tests/performance/fooditems/perf_update.go
- [ ] T027 Create FoodItem deletion performance tests in tests/performance/fooditems/perf_delete.go

#### Integration Tests (✅ COMPLETE)

- [X] T028 Create FoodItem creation integration test in tests/integration/fooditems/create_test.go
- [X] T029 Implement FoodItem listing integration test in tests/integration/fooditems/list_test.go
- [X] T030 Write FoodItem update integration tests in tests/integration/fooditems/update_test.go
- [X] T031 Create FoodItem deletion integration tests in tests/integration/fooditems/delete_test.go
- [X] T032 Implement tenant isolation tests in tests/integration/fooditems/multi_tenant_test.go

### Implementation Tasks for User Story 2

- [ ] T033 Create PostgreSQL container runtime wrapper (tests/db/runtime.go) using testcontainers-go lifecycle methods
- [ ] T034 Implement database connection pooler configuration (tests/db/connection.go) with proper resource limits
- [X] T035 Setup schema initialization in tests/db/schema.sql

### Database Verification Implementation

- [X] T036 Create SQL assertion utilities (tests/db/assertions.go)
- [X] T037 Implement transaction rollback utilities (tests/db/transaction_helper.go)
- [ ] T038 Generate database schema verification tests in docs/specs/food_menu_system/features/schema-verification.feature

---

## Phase 5: User Story 3 - End-to-End Verification Suite (Priority: P3) ⏳ PENDING

### Tests for User Story 3 (CI/CD Integration)

- [ ] T039 Create GitHub Actions workflow for feature tests in `.github/workflows/feature-tests.yml`
- [ ] T040 Implement integration test CI pipeline in `.github/workflows/integration-tests.yml`

### Implementation Tasks for User Story 3

- [ ] T041 Add coverage reporting setup (tests/coverage/collector.go) using gotest.tools/gocov for >90% coverage validation
- [ ] T042 Create test report generator (tests/reports/formatter.go) producing JUnit and HTML format outputs

---

## Phase 6: Polish & Cross-Cutting Concerns ⏳ PENDING

### Task T045: Code quality improvements

- [ ] T045 Implement comprehensive code coverage report (tests/coverage/report.go) with 90% target validation
- [ ] T046 Create benchmark tests in tests/performance/benchmarks.go for performance regression detection
- [ ] T047 Add static analysis configuration in `.golangci.yml` for linting rules

### Task T048: Documentation and maintainability

- [ ] T048 Update comprehensive testing guide in docs/testing-guide.md with new BDD/TestContainers patterns
- [ ] T049 Generate API documentation for test utilities (tests/db/docs.go)
- [ ] T050 Run quickstart validation tests against generated test suite structure

---

## Implementation Progress Summary

| Phase | Status | Tasks Complete | Description |
|-------|--------|---------------|-------------|
| Phase 1 | ✅ Complete | 3/3 | Setup dependencies (testcontainers-go, gherkin-go) |
| Phase 2 | ⏳ In Progress | ~70% | Database container factory, connection helpers, HTTP test utilities created |
| Phase 3 | ⏳ Pending | 0/4 | BDD/Gherkin framework setup |
| Phase 4 | ⏳ In Progress | ~60% | Integration tests for CRUD operations (create, list, update, delete) complete |
| Phase 5 | ⏳ Pending | 0/6 | CI/CD integration and coverage reporting |
| Phase 6 | ⏳ Pending | 0/8 | Code quality, benchmarks, documentation |

---

## Completed Components ✅

### Database Infrastructure (Phase 2)
- ✅ PostgreSQL container factory using testcontainers-go v0.43
- ✅ Database connection helper with sqlx and proper lifecycle management
- ✅ Connection utilities for test database setup
- ✅ Schema initialization SQL files
- ✅ Tenant fixture generator for multi-tenant testing

### Test Utilities Created
- ✅ FoodItem creation integration tests (create_test.go)
- ✅ FoodItem listing integration tests (list_test.go) with pagination
- ✅ FoodItem update integration tests (update_test.go) - partial/full update scenarios
- ✅ FoodItem deletion integration tests (delete_test.go)
- ✅ Multi-tenant isolation tests (multi_tenant_test.go)

---

## Next Steps 🔜

1. **T035**: Create schema migration runner for PostgreSQL container initialization
2. **T034**: Add connection pool configuration to ConnectionHelper with max connections settings
3. **T012**: Implement food category seeding utility (categories.fixture.go)
4. **T009**: Complete container lifecycle manager with proper start/stop functionality

### Priority Queue:
- High: T035, T034, T012 - Required for complete integration test execution
- Medium: T009 - Improves container lifecycle management
- Low: CI templates (T014), BDD framework (Phase 3) - Nice to have features
