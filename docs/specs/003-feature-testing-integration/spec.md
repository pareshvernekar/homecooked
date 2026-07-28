# Feature Testing and Integration Testing Specification

**Feature Branch**: 003-feature-testing-integration  
**Created**: 2026-07-25  
**Status**: Draft - Awaiting Data Setup Clarifications  
**Input**: Implement comprehensive feature tests and integration tests for all REST API endpoints using TestContainers-go to manage database connections. Tests must exercise REST API endpoints, output Gherkin feature files, generate tests from Gherkin scenarios, verify API responses, and validate database state changes.

---

## Clarifications (Session 2026-07-26)

### Q1: Tenant Setup for Testing - Data Seed Strategy
**Answer**: Option A - Automatic default tenant creation

**Context**: The application uses multi-tenant architecture with PostgreSQL Row-Level Security (RLS). All tables require `tenant_id` column values. Tests need isolated tenant contexts to prevent data leakage between concurrent test scenarios.

**Decision Rationale**: For feature testing with BDD/Gherkin, the primary goal is rapid test iteration and clear behavioral validation. Creating a single default "test-tenant" eliminates boilerplate code in every test scenario while TestContainers ensures database isolation between parallel test runs. This approach aligns best with the stated goal of >90% code coverage without excessive setup complexity.

**Implementation Approach**: 
- Framework creates `test-tenant` container before first integration test run
- X-Tenant-ID header defaults to "test-tenant" for all scenarios in same test suite
- TestContainers ensures separate PostgreSQL instances for parallel test execution
- No manual tenant setup required in Gherkin feature files

**Affected Components**:
- `internal/testing/fixtures/tenants.go` - Tenant creation utility with single default instance
- Integration test scaffolding - Removes explicit tenant setup boilerplate
- Multi-tenant validation tests - Test cross-tenant isolation via header spoofing

---

### Q2: Food Category Data - Pre-seeded Categories vs Dynamic Creation
**Answer**: Option A - Pre-seed canonical categories

**Context**: The application normalizes category names to lowercase slug format (e.g., "Vegetarian" → "vegetarian"). All FoodItems reference categories via foreign key. Tests require valid category references before creating food items, or tests must skip category creation step and only test item operations.

**Decision Rationale**: Since the application validates against specific category enums (`IsValidCategory()`), tests must exercise all valid values to achieve >90% coverage. Pre-seeding eliminates setup complexity while ensuring normalization edge cases are tested (e.g., "VEGAN" → "vegan", "NON-VEGETARIAN" → "non-vegetarian"). This supports comprehensive validation of the business logic without requiring category-specific test scenarios.

**Implementation Approach**: 
- Create `internal/testing/fixtures/categories.go` with pre-populated categories during test suite initialization
- Framework sets up: vegetarian, non-vegetarian, vegan, dessert, beverage, main_course, sides, appetizer
- All FoodItem CRUD tests reference existing UUIDs, reducing code by ~40%
- Ensures category name normalization is tested at service layer (case conversion logic validated)

**Affected Components**:
- `internal/testing/fixtures/categories.go` - Pre-seeded test data with all valid category enums
- `internal/services/foodcategory/service.go` - Category normalization behavior validated across multiple cases
- Integration tests can focus on business logic, not setup boilerplate


**Question**: What is the recommended approach for food category data that all other entities depend on?

**Context**: The application normalizes category names to lowercase slug format (e.g., "Vegetarian" → "vegetarian"). All FoodItems reference categories via foreign key. Tests require valid category references before creating food items, or tests must skip category creation step and only test item operations.

**Suggested Answers**:

| Option | Answer | Implications |
|--------|---------|--------------|
| A | **Pre-seed canonical categories** - Framework creates standard categories (vegetarian, non-vegetarian, vegan, dessert, beverage, main_course, sides, appetizer) before all tests | ✅ All CRUD operations work out-of-the-box. Covers all valid category enum values defined in `IsValidCategory()`. Tests don't need category setup logic. Ensures complete test coverage of normalization behavior (e.g., "VEGAN" → "vegan").<br>**Recommended choice** |
| B | **Dynamic category creation per test** - Each CreateFoodItem test first creates required categories in same transaction | ✅ Tests remain independent, no shared state assumptions. If one category creation fails, entire test fails immediately (no cleanup). Requires more complex setup code.<br>**Tradeoff**: More robust isolation but higher implementation complexity for tests |
| C | **Skip category seeding** - Tests only validate FoodItem operations assuming categories exist | ✅ Minimal test setup overhead. Fast iteration time.<br>**Con**: Cannot test full CRUD lifecycle without additional setup. May skip important validation scenarios (invalid category references). Requires separate category test suite. |

---

### Q3: Test Isolation Strategy - Database Schema Versioning

**Question**: How should schema migrations be managed for TestContainers-based integration testing?

**Context**: Production database has migration scripts in `migrations/` directory. Tests using TestContainers need equivalent schema initialization to prevent migration failures or inconsistent test environments.

**Suggested Answers**:

| Option | Answer | Implications |
|--------|---------|--------------|
| A | **Inline migrations in container startup** - Each TestContainer PostgreSQL instance has migration scripts baked into image, executed via `RUN` commands during container build | ✅ Consistent environment across all test runs. No separate migration orchestration needed. Tests start with known-good schema state.<br>**Implementation**: Custom Docker images per migration version (m1: food_category table, m2: add food_item table). |
| B | **Separate migration runner process** - Test framework spawns ephemeral migration service container that applies Flyway/Liquibase before tests connect | ✅ Production-grade migration orchestration. Supports rollback testing scenarios.<br>**Tradeoff**: Adds container startup overhead per test run (5-10 seconds). More complex infrastructure. Better for validating migration logic itself. |
| C | **DDL statements in initialization** - Simple CREATE TABLE/INDEX scripts as Go code, executed during TestContainer Start() sequence | ✅ No external dependencies. Easy to read/maintain inline migrations.<br>**Limitation**: Complex migrations (schema versioning, data transformations) harder to express. Migration execution must happen synchronously before test connects. |
**Decision**: **Option C selected** - DDL statements in initialization

**Rationale**: For feature testing focused on business logic validation, lightweight Go-based DDL execution provides zero-configuration setup without Flyway dependencies. Keeps test startup fast (~500ms) while avoiding complex migration orchestration. Sufficient for >90% coverage target since tests validate API behavior not migration logic itself.

**Implementation**:
- `internal/testing/database/schema.sql` with inline CREATE TABLE statements
- Executed during container initialization via RunAndReturnCommand()
- All integration tests get consistent table structure automatically

---

### Q4: Error Scenarios and Negative Testing - Data Validation Coverage

**Question**: What level of detail should error scenario tests cover for data validation failures?

**Context**: API endpoints validate request body fields (required fields, enum values like "available/unavailable/low_stock", price bounds >= 0). Invalid inputs must return appropriate HTTP status codes with error messages. Test coverage needs explicit error scenarios to be meaningful.

**Answer**: **Option A - Exhaustive validation matrix**

**Rationale**: For the >90% code coverage target, comprehensive test of all invalid states is essential. Testing "missing field", "empty value", "wrong type" (invalid UUID format), "wrong enum value", and boundary conditions like negative prices ensures complete validation logic verification. Each FoodItemCreateRequest field gets tested with:
- `name`: omitted, empty string "", too long (>100 chars)
- `price`: negative number, zero, extreme large value (999999.99)
- `availability_status`: omitted, wrong enum "not_available", all 3 valid values tested
- `category_name`: non-existent category UUID (404), invalid format

**Implementation**: Create `internal/testing/assertions/validation.go` with comprehensive validation assertion helpers that verify HTTP status codes and error messages match expected outcomes for each invalid input state.

---

### Q5: Transaction Rollback and Cleanup - Database State Management

**Question**: How should test database transactions be handled for cleanup between assertion blocks?

**Context**: Integration tests modify database state through INSERT/UPDATE/DELETE operations. After each test scenario, state should be clean to prevent cross-test pollution. Options include transaction per test (PostgreSQL atomicity), manual rollback on failure, or explicit teardown hooks.

**Answer**: **Option A - Transaction-per-test with automatic rollback**

**Rationale**: For TestContainers-based integration testing where >90% coverage is required, this approach provides guaranteed clean state between all tests without boilerplate overhead. Since each test runs in its own container, the slight transaction setup overhead (~10ms) is negligible compared to container startup cost. PostgreSQL native transaction rollback ensures no cross-test pollution while keeping test code simple and maintainable.

**Implementation**:
- `internal/testing/database/transaction.go` - Transaction wrapper utilities with `NewTx()` helper that returns cleanup function
- Test structure: `defer tx.Rollback()` at start of each integration test, `tx.Commit()` only after successful assertions
- Ensures 100% transaction rollback rate for all error scenarios (SC-004 requirement)


---

## Success Criteria Update Required

> **Note**: These success criteria need refinement based on answers to questions above.

### Measurable Outcomes (To be updated after Q1-Q5)

- **SC-001**: Feature tests achieve >90% API endpoint coverage - *Remains same*
- **SC-002**: Integration tests execute successfully against isolated PostgreSQL instances - *Depends on answers to Q2 and Q3 for test reliability rate*
- **SC-003**: Gherkin feature files are automatically converted to executable Go test cases with >90% code coverage - *Remains same*
- **SC-004**: All database operations (INSERT, SELECT, UPDATE, DELETE) performed during API testing are verified in TestContainer-managed PostgreSQL instances - *Add: Transaction rollback rate must be 100% (no cross-test pollution)*
- **SC-005**: Tenant isolation validation passes all cross-tenant data leakage tests - *Depends on answer to Q1 for implementation approach*

---

## Next Steps

After clarifications are answered, I will:
1. Update spec with confirmed decisions
2. Create detailed test architecture based on chosen data setup strategies
3. Generate appropriate fixtures and seeding utilities matching the selected approach
