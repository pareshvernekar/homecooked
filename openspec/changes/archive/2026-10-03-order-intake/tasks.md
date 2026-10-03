## 1. Schema and fixtures

- [x] 1.1 Add migration for `customer_order` (or `tenant_order`) with tenant composite key, menu_id FK, customer_name/phone, received_at, expected_at, pickedup_at, status, customization_text, total_override, frozen_total, is_active, timestamps (REQORDER001–REQORDER005)
- [x] 1.2 Add migration for `order_item`, `order_item_component_selection`, and freeze columns on lines (REQOLINE001–REQOLINE004)
- [x] 1.3 Add migration for `order_payment` with mode check constraint and positive amount (REQPAY001–REQPAY004)
- [x] 1.4 Sync `scripts/init/` and `tests/db/schema.sql`; add DB tests for FKs, mode/status checks, and tenant isolation

## 2. Models and menu price carve-out

- [x] 2.1 Replace stub `internal/models/order.go` and order DTOs with header/line/selection/payment models and request/response types
- [x] 2.2 Allow price-only updates on size options when parent menu is published; keep structural mutations draft-only (REQITEM006S03)
- [x] 2.3 Update menu-item service/handler tests for published price-only update vs structural reject (REQITEM006)

## 3. Order and line services

- [x] 3.1 Implement order repository create/list/get/update/deactivate with tenant scoping (REQORDER001–REQORDER003)
- [x] 3.2 Implement order service: published daily|catering gate, defaults, status update, freeze at PICKEDUP, total resolution (REQORDER001, REQORDER004, REQOLINE003)
- [x] 3.3 Implement line repository/service: selections validation against menu tree, overrides, customization, add/update/remove only when unfulfilled (REQOLINE001–REQOLINE004)
- [x] 3.4 Add service tests for create gates, live price change, overrides, freeze, unpaid PICKEDUP (REQORDER001, REQORDER004, REQOLINE003, REQPAY003)

## 4. Payment service

- [x] 4.1 Implement payment repository and service: record/list, derived aggregates, overpay allowed (REQPAY001–REQPAY004)
- [x] 4.2 Add payment service tests for partial, split modes, overpay, post-PICKEDUP payment, override-driven overage (REQPAY002, REQPAY003)

## 5. HTTP API

- [x] 5.1 Implement order handlers and routes under `/api/v1/orders` (create/list/get/patch) (REQORDER001–REQORDER005)
- [x] 5.2 Implement nested line routes (add/update/delete) (REQOLINE001–REQOLINE004)
- [x] 5.3 Implement nested payment routes (create/list) (REQPAY001, REQPAY004)
- [x] 5.4 Wire DI in `cmd/main.go` and `internal/server`; remove obsolete order stubs from docs/CLAUDE.md API section

## 6. Traceability and verification

- [x] 6.1 Annotate production code and tests with REQORDER*, REQOLINE*, REQPAY*, REQITEM006
- [x] 6.2 Add HTTP integration tests for order create → lines → pay → PICKEDUP freeze
- [x] 6.3 Run focused package tests and `go test ./...`; fix regressions
