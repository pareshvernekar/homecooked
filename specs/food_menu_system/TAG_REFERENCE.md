# Tag Reference Guide

This document provides a quick reference for using task tags (T###) throughout the Food Menu System codebase for traceability.

## Quick Reference

### Where to Find Information
| Need | Location | Reference |
|------|----------|-----------|
| Task definitions & status | [tasks.md](tasks.md) | All tasks T001-T999 with status |
| Task→Code mapping | [TRACEABILITY.md](TRACEABILITY.md#file-to-task-cross-reference) | By file and by task |
| Implementation details | [TRACEABILITY.md](TRACEABILITY.md) | Detailed status for each phase |
| Spec requirements | [spec.md](spec.md) | User requirements and features |
| API design | [api_specification.md](api_specification.md) | API endpoints and contracts |
| Implementation plan | [plan.md](plan.md) | Architecture and design decisions |

---

## Task IDs by Phase

### Phase 1: Setup (T001-T008) ✅ COMPLETE
- T001: Create project structure
- T002: Initialize Git repository
- T003: Set up Go environment
- T004: Install required packages
- T005: Create configuration files
- T006: Implement tenant context middleware
- T007: Create Makefile
- T008: Set up linting

### Phase 2: Database (T011-T016) ❌ PENDING
- T011: Create database tables with tenant_id
- T012: Set up row-level security policies
- T013: Database connection with tenant context
- T014: Implement repository pattern
- T015: Create migration scripts
- T016: Apply initial migrations

### Phase 3: Data Models (T021-T024) 🔄 PARTIAL
- T021: Define Go structs for entities
- T022: Create DTOs
- T023: Set up validation
- T024: Implement error handling

### Phase 4: API Endpoints (T031-T041) 🔄 PARTIAL
- T031: Define API endpoints ✅
- T032: Create API contracts ❌
- T033: Set up router ❌
- T034: Implement middleware ❌
- T035-T041: Endpoint implementations ❌

### Phase 5: Business Logic (T051-T054) ❌ PENDING
- T051: Create service interfaces
- T052: Implement business logic
- T053: JWT authentication
- T054: Role-based access control

### Phase 6: Testing (T061-T064) ❌ PENDING
- T061: Unit tests for services
- T062: Unit tests for handlers
- T063: Integration tests
- T064: Service integration tests

### Phase 7: Deployment (T071-T075) ❌ PENDING
- T071: Create Dockerfile
- T072: Create database Dockerfile
- T073: Set up docker-compose
- T074: CI pipeline
- T075: CD pipeline

### Phase 8: Operations (T081-T084) ❌ PENDING
- T081: Application monitoring
- T082: Monitoring alerts
- T083: Centralized logging
- T084: Security audits

### Utilities
- T997: Documentation
- T998: Test execution
- T999: Dependencies

---

## Code File Conventions

### Naming Patterns
```
internal/[module]/[feature].go          // Implementation
internal/[module]/[feature]_test.go     // Tests
```

### Comment Headers
Add task tags at the top of implementation and test files:

**Implementation file**:
```go
// T006: Implement tenant context middleware
// Extracts X-Tenant-ID from request headers and stores in context
package middleware

import (
    // ...
)
```

**Test file**:
```go
// T006_TEST: Tests for tenant context middleware
package middleware

import (
    "testing"
)
```

---

## Implementation Status Legend

| Symbol | Meaning | Action Required |
|--------|---------|-----------------|
| ✅ | Complete & tested | Review, use as reference |
| 🔄 | Partial/In-progress | Complete implementation, add tests |
| ❌ | Not started | Design needed, then implement |

---

## Finding Task Information

### By Task ID
```
1. Open tasks.md
2. Search for "T###"
3. Look at: Status, Implementation, Test Reference, Spec Reference
```

### By Code File
```
1. Open TRACEABILITY.md
2. Go to "File-to-Task Cross-Reference"
3. Find your file name
4. See which tasks it implements
```

### By Feature/Requirement
```
1. Open spec.md
2. Find the requirement/feature
3. Cross-reference to TRACEABILITY.md
4. Find associated task IDs
5. Check implementation status in tasks.md
```

---

## Critical Dependencies

### Must Complete Before Phase 2
- ✅ T001-T008 (Phase 1 Setup)

### Must Complete Before Phase 3
- T011 (Database schema)
- T013 (Database connection)

### Must Complete Before Phase 4
- T014 (Repository pattern)
- T021 (Data models)
- T023 (Validation)

### Must Complete Before Phase 6
- T035-T041 (API endpoints)
- T051-T052 (Services)

---

## Example: Implementing Task T021 (Define Go Structs)

### Step 1: Check Status
```
Look in tasks.md:
- Status: PENDING
- Dependency: T011 (must be done first) ❌
- Blocking: T022, T023, T035-T041
```

### Step 2: Review Specification
```
Open spec.md → Section "Data Models"
See all required entities:
- FoodItem ✅ (exists)
- Tenant ❌ (missing)
- User ❌ (missing)
- Category ❌ (missing)
- WeeklyMenu ❌ (missing)
- CateringMenu ❌ (missing)
- MenuItem ❌ (missing)
- Order ❌ (missing)
- OrderItem ❌ (missing)
- Notification ❌ (missing)
```

### Step 3: Create Implementation
```go
// T021: Define Go structs for all entities
// Defines data models for Tenant, FoodItem, Category, Menu, MenuItem, Order, OrderItem, Notification
package models

// Tenant represents a tenant organization
type Tenant struct {
    ID        string    `db:"id"`
    Name      string    `db:"name"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}

// ... more models
```

### Step 4: Create Tests
```go
// T021_TEST: Tests for entity models
// Validates struct definitions and database mappings
package models

import "testing"

func TestTenantModel(t *testing.T) {
    // Test struct tags
    // Test validation
    // Test marshalling
}
```

### Step 5: Update Documentation
```
1. Open tasks.md
2. Find T021
3. Update:
   - Status: 🔄 PARTIAL (was ❌ NOT STARTED)
   - Implementation: internal/models/tenant.go, etc.
   - Test Reference: internal/models/tenant_test.go, etc.

4. Open TRACEABILITY.md
5. Update Phase 3 section with new files
6. Update file-to-task table
```

---

## Quality Checklist

Before marking a task complete:

- [ ] Code implementation file created/updated
- [ ] Test file created with good coverage
- [ ] Comment headers added with task ID
- [ ] tasks.md updated with file references
- [ ] TRACEABILITY.md updated
- [ ] All tests pass: `make test`
- [ ] Code passes linting: `make lint`
- [ ] Code review completed
- [ ] Status updated to ✅ COMPLETE or 🔄 PARTIAL

---

## Reviewing Code by Task

### For Code Reviewer
```
1. Get task ID from code (look for "// T###:" comment)
2. Open tasks.md, find task details
3. Review code against spec reference
4. Check corresponding test file exists
5. Verify test coverage is adequate
6. Check that code doesn't break dependent tasks
```

### For Task Implementer
```
1. Get assigned task ID (e.g., T035)
2. Open tasks.md, find detailed requirements
3. Check dependencies in "Blocking" and "Dependency" fields
4. Review spec reference for requirements
5. Implement with test-first approach
6. Update files and documentation when done
```

---

## File Traceability Example

### File: `internal/handlers/fooditem.go`

**Header**:
```go
// T035: Implement Food Items endpoints
// Provides REST API handlers for CRUD operations on food items
// Handles: GET /api/food-items, POST /api/food-items, etc.
```

**In tasks.md**:
```
T035: Implement Food Items endpoints
  - Implementation: internal/handlers/fooditem.go
  - Test Reference: internal/handlers/fooditem_test.go
  - Spec Reference: api_specification.md#food-items-endpoints
```

**In TRACEABILITY.md**:
```
| internal/handlers/fooditem.go | T035 | Partial | ✅ fooditem_test.go |
```

---

## Troubleshooting

### "I can't find which task implemented this file"
1. Open TRACEABILITY.md
2. Search for file name in "File-to-Task Cross-Reference" table
3. Column 2 shows associated task IDs

### "I don't know where to start"
1. Open TRACEABILITY.md
2. Look for status "PENDING" in next phase
3. Check dependencies - pick task with all dependencies complete
4. Follow example workflow above

### "The tag system seems outdated"
1. Cross-check tasks.md against actual code files
2. Update any discrepancies
3. Update TRACEABILITY.md
4. Re-run this verification process monthly

---

## Updating Tags and References

### When Creating New Task
1. Assign next available T### in sequence
2. Add entry to tasks.md with:
   - Clear description
   - Dependencies (if any)
   - Test requirements
3. Link to spec section
4. Add to TRACEABILITY.md when code created

### When Implementation Changes
1. Update Implementation field in tasks.md
2. Update Code Reference with specific files/functions
3. Update TRACEABILITY.md file-to-task mapping
4. Update test reference if tests change

### When Moving to Next Phase
1. Verify all previous phase tasks complete
2. Check no blocking dependencies exist
3. Create new phase task IDs
4. Update critical path in TRACEABILITY.md
