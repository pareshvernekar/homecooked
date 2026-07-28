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

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

### Task T001: Initialize Go modules for test suite dependencies

- [ ] T001 Create go.mod entry for testcontainers-go dependency (github.com/testcontainers/testcontainers-go v0.37.x)
- [ ] T002 Add gherkin.go package for BDD scenario parsing (github.com/cucumber/gherkin-go/v13)
- [ ] T003 [P] Initialize Go workspace configuration (go.work) for multi-module test suite

---

## Phase 2: Foundational Infrastructure (Blocking Prerequisites)

**Purpose**: Core testing infrastructure that MUST be complete before ANY feature tests can run

### Task T004: Create BDD/Gherkin framework setup

- [ ] T004 [P] Create `internal/bdd/framework.go` with Gherkin scenario runner using go-ginkgo
- [ ] T005 [P] Define step definitions pattern for Given/When/Then keywords in `internal/bdd/steps/`
- [ ] T006 [P] Setup BDD test directory structure under `docs/specs/food_menu_system/features/testsuite-go/bdd/`

### Task T007: Create TestContainers PostgreSQL client infrastructure

- [ ] T007 [P] Create database container factory in `internal/testing/database/postgres_container.go` using testcontainers-go
- [ ] T008 [P] Implement PostgreSQL connection helper with migration support in `internal/testing/database/db.go`
- [ ] T009 [P] Create container lifecycle manager (start/migrate/stop) for test isolation
- [ ] T010 [P] Define database schema templates in `docs/specs/food_menu_system/schema/sql/` (food_category, food_item tables)

### Task T011: Setup shared test fixtures and utilities

- [ ] T011 [P] Create tenant fixture generator in `internal/testing/fixtures/tenants.go` (for multi-tenant isolation)
- [ ] T012 [P] Implement food category seeding utility in `internal/testing/fixtures/categories.go` (vegetarian, non-vegetarian, vegan categories)
- [ ] T013 [P] Create test utilities for HTTP request/response handling in `internal/testing/http/test_helper.go`

### Task T014: Configure CI/CD integration templates

- [ ] T014 [P] Create GitHub Actions workflow template in `.github/workflows/test-suite.yml`
- [ ] T015 [P] Define environment configuration for test containers (TEST_DB_HOST, TEST_DB_PORT)
- [ ] T016 [P] Add documentation for running tests locally in `docs/testing-guide.md`

---

## Phase 3: User Story 1 - Feature Testing with BDD (Priority: P1) 🎯 MVP

**Goal**: Implement Gherkin-based behavioral testing for all FoodItem CRUD endpoints  
**Independent Test**: Each .feature file can be parsed, compiled to Go tests, and executed independently without database dependency

### Tests for User Story 1 (BDD/Gherkin Scenarios)

- [ ] T017 [P] [US1] Create Gherkin feature file for FoodItem creation in `docs/specs/food_menu_system/features/food-items.feature` - Scenario: Valid food item creation
- [ ] T018 [P] [US1] Add Gherkin scenario for FoodItem listing in `docs/specs/food_menu_system/features/food-items.feature` - Scenario: List all food items with pagination
- [ ] T019 [P] [US1] Document FoodItem update validation rules in `docs/specs/food_menu_system/features/food-items.feature` - Scenario Outline: Update with partial data
- [ ] T020 [P] [US1] Create Gherkin scenarios for FoodItem deletion in `docs/specs/food_menu_system/features/food-items.feature` - Scenario: Delete non-existent food item

### Implementation Tasks for User Story 1

- [ ] T021 [P] [US1] Implement BDD step definition generator in `internal/bdd/generator.go` to convert Gherkin to Go test functions
- [ ] T022 [P] [US1] Create auto-converter utility in `internal/bdd/converter.go` using cucumber-msgpack for message parsing
- [ ] T023 [P] [US1] Setup BDD command-line interface in `cmd/test-generate/main.go` for running feature tests from CLI

---

## Phase 4: User Story 2 - Integration Testing with TestContainers (Priority: P2)

**Goal**: Implement database-verified integration tests using PostgreSQL containers  
**Independent Test**: Each test can be executed independently with isolated database instance, verifying both HTTP responses and SQL statements

### Tests for User Story 2 (TestContainers Integration)

- [ ] T024 [P] [US2] Create FoodItem creation integration test in `internal/integration/fooditems/create_test.go` - Given tenant has categories, POST to /api/v1/food-items
- [ ] T025 [P] [US2] Implement FoodItem listing integration test in `internal/integration/fooditems/list_test.go` - Verify pagination response with database query validation
- [ ] T026 [P] [US2] Write FoodItem update integration tests in `internal/integration/fooditems/update_test.go` - Test partial and full updates with DB verification
- [ ] T027 [P] [US2] Create FoodItem deletion integration tests in `internal/integration/fooditems/delete_test.go` - Verify cascade/deletion from database
- [ ] T028 [P] [US2] Implement tenant isolation tests in `internal/integration/fooditems/multi_tenant_test.go` - Verify no cross-tenant data leakage

### Implementation Tasks for User Story 2

- [ ] T029 [P] [US2] Create PostgreSQL container runtime wrapper in `internal/testing/container/runtime.go` using testcontainers-go lifecycle methods
- [ ] T030 [P] [US2] Implement database connection pooler configuration in `internal/testing/containers/pooler.go` with proper resource limits
- [ ] T031 [P] [US2] Setup migration runner for PostgreSQL container in `internal/testing/database/migrator.go` using custom Flyway-like Go-based migration execution

### Database Verification Implementation

- [ ] T032 [P] [US2] Create SQL assertion utilities in `internal/testing/db/assertions.go` (CountRows, VerifyColumns, CheckConstraints helpers)
- [ ] T033 [P] [US2] Implement transaction rollback utilities in `internal/testing/containers/transaction_helper.go` for test cleanup
- [ ] T034 [P] [US2] Generate database schema verification tests in `docs/specs/food_menu_system/features/schema-verification.feature` and convert to Go assertions

---

## Phase 5: User Story 3 - End-to-End Verification Suite (Priority: P3)

**Goal**: Complete CI/CD integration for automated testing pipeline  
**Independent Test**: Full test suite can be executed as part of CI/CD without manual intervention

### Tests for User Story 3 (CI/CD Integration)

- [ ] T035 [P] [US3] Create GitHub Actions workflow for feature tests in `.github/workflows/feature-tests.yml` - Parse and run all .feature files
- [ ] T036 [P] [US3] Implement integration test CI pipeline in `.github/workflows/integration-tests.yml` - Spin up containers, verify DB state
- [ ] T037 [P] [US3] Add coverage reporting setup in `internal/testing/coverage/collector.go` using gotest.tools/gocov for >90% coverage validation
- [ ] T038 [P] [US3] Create test report generator in `internal/testing/reports/formatter.go` producing JUnit and HTML format outputs

### Implementation Tasks for User Story 3

- [ ] T039 [P] [US3] Implement parallel test execution orchestrator in `cmd/test-parallel/main.go` for concurrent database container management
- [ ] T040 [P] [US3] Create environment validation utility in `internal/testing/env/validation.go` checking required configuration before test runs

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and overall test quality

### Task T041: Code quality improvements

- [ ] T041 [P] Implement comprehensive code coverage report in `internal/testing/coverage/report.go` with 90% target validation
- [ ] T042 [P] Create benchmark tests in `internal/handlers/benchmarks.go` for performance regression detection
- [ ] T043 [P] Add static analysis configuration in `.golangci.yml` for linting rules

### Task T044: Documentation and maintainability

- [ ] T044 [P] Update comprehensive testing guide in `docs/testing-guide.md` with new BDD/TestContainers patterns
- [ ] T045 [P] Generate API documentation for test utilities in `internal/testing/api/docs.go`
- [ ] T046 Run quickstart validation tests against generated test suite structure

---

## Dependencies & Execution Order

### Phase Dependencies

**Phase 1 (Setup)**: No dependencies - can start immediately

**Phase 2 (Foundational)**: Depends on Setup completion - BLOCKS all user stories
- All database/container infrastructure must be ready before T024-T034 tasks can execute
- Without this, integration tests cannot connect to PostgreSQL containers

**Phase 3 (US1 - BDD/Gherkin)**: Can start after Phase 1 setup but doesn't require Phase 2 (optional enhancement)
- Task T017-T023 for Gherkin framework implementation
- Independent of database infrastructure (tests first, generates tests)

**Phase 4 (US2 - TestContainers)**: Can run in parallel with US1 or after Phase 2 completion
- Requires T007-T010 to be complete for PostgreSQL container support
- T024-T034 cannot execute without database connection infrastructure

**Phase 5 (US3 - CI/CD)**: Dependent on Phases 3 & 4 completion
- Needs BDD framework from US1 and TestContainers from US2 to orchestrate execution

**Phase 6 (Polish)**: After all phases complete
- Final improvements across entire test suite

### User Story Dependencies

- **User Story 1 (P1 - BDD Framework)**: Stands alone - creates Gherkin files that generate Go tests
- **User Story 2 (P2 - Integration Tests)**: Independent of US1 but provides executable implementation of US1 scenarios
- **User Story 3 (P3 - CI/CD)**: Orchestrates both US1 and US2 execution in pipeline

### Within Each User Story

**US1 (BDD Framework)**:
- T017-T020: Create Gherkin feature files (.feature)
- T021-T023: Build converter to generate executable Go tests from .feature files

**US2 (Integration Tests)**:
- T029-T031: Database infrastructure implementation
- T024-T034: Test scenarios with database verification (requires infrastructure first)

**US3 (CI/CD)**:
- T035-T038: Pipeline integration (depends on US1 and US2 being complete)
- T039-T040: Parallel execution orchestration (needs both frameworks ready)

### Parallel Opportunities

```bash
# Phase 1 - All parallel [P]:
T001, T002, T003  # Module initialization

# Phase 2 - Database infrastructure all parallel [P]:
T007, T008, T009, T010, T011, T014  # Container setup, fixtures, CI templates

# Phase 3 (US1) - Gherkin tests and converter parallel:
T017-T020, T021-T023  # Test definitions vs generator implementation

# Phase 4 (US2) - Integration tests with DB infrastructure:
T029-T031, T024-T034  # Infrastructure before test scenarios

# Parallel Team Strategy:
- Dev A: Phases 1 & 2 (Infrastructure)
- Dev B: Phase 3 (BDD/Gherkin framework)
- Dev C: Phase 4 (Integration tests)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. **Setup**: Complete Phase 1 tasks T001-T003
2. **Infrastructure**: Create BDD framework with Gherkin support (T017-T023)
3. **Test Generation**: Generate executable tests from .feature files
4. **VALIDATE**: Run generated tests against existing API endpoints
5. **DEPLOY**: Demo test coverage to stakeholders

### Incremental Delivery

**Iteration 1**: BDD Framework Only
- Complete Phases 1 & 3 (Setup + Gherkin)
- Generate all test scenarios from .feature files
- Validates behavior without database

**Iteration 2**: Add Integration Testing  
- Complete Phase 2 + 4 (Infrastructure + TestContainers)
- Execute same scenarios with database verification
- Validates persistence layer correctness

**Iteration 3**: CI/CD Pipeline
- Complete Phase 5 (CI/CD integration)
- Automated test execution on every commit
- Coverage reporting and quality gates

### Parallel Team Strategy

```
Day 1 Setup (All team):
- Team Member 1: T001-T003 (Go module setup)
- Team Member 2: T004-T016 (BDD/Container framework + docs)

Day 2 Split:
- Team A: T017-T023 (Gherkin features + generator)
- Team B: T029-T034 (TestContainers + SQL assertions)

Day 3 Integration:
- Combine both frameworks, run parallel test suites
```

---

## Notes

- [P] tasks can run in parallel when different files, no dependencies
- [US1], [US2], [US3] labels map task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (especially for US2 database tests)
- Generate code after feature files are written (BDD-driven development)
