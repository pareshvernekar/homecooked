## 1. Schema and migrations

- [x] 1.1 Add migration for `size_unit` (system + tenant customs) and seed standards: serving, tray, piece, dozen, kg, g, liter (REQSIZE001, REQSIZE002)
- [x] 1.2 Add migration for `menu`, `menu_category`, `menu_item`, `menu_item_component` (with `default_size_option_id`), `menu_item_component_size_option` with composite tenant keys, FKs, sequences, `status` (draft|published), `is_active`, menu_type metadata columns, and partial unique index on active published weekly `(tenant_id, start_date)` (REQMENU001, REQMENU005–REQMENU009, REQITEM001, REQITEM003–REQITEM005)
- [x] 1.3 Add migration dropping `food_item.price`; sync `scripts/init/`, `tests/db/schema.sql`, and fixtures (REQFOOD001)
- [x] 1.4 Add DB tests for constraints: simple/combo cardinality hooks where enforced in DB, component size_option price/qty checks, default FK integrity, soft-delete, size_unit uniqueness, weekly published start_date uniqueness (REQITEM001, REQITEM003–REQITEM004, REQSIZE002, REQMENU009)

## 2. Food-item catalog price removal

- [x] 2.1 Remove `Price` from food-item models, requests, responses, and validation (REQFOOD001)
- [x] 2.2 Update food-item repository/service/handler so price is never read or written (REQFOOD001)
- [x] 2.3 Update food-item unit and integration tests for create/list/get/update without price (REQFOOD001)

## 3. Size units

- [x] 3.1 Add SizeUnit models and repository (list system+tenant, create custom, soft-deactivate custom only) (REQSIZE001–REQSIZE003)
- [x] 3.2 Add size-unit service/handlers and routes under `/api/v1/size-units` (REQSIZE001–REQSIZE002)
- [x] 3.3 Add tests for standards present, custom isolation, duplicate code rejection, system unit immutability (REQSIZE001–REQSIZE003)

## 4. Menu models and repository

- [x] 4.1 Add Menu, MenuCategory, MenuItem, Component, ComponentSizeOption models and DTOs (including `status`, menu_type metadata, default size, live-join food display fields, required computed `default_total`) (REQMENU001, REQMENU003, REQMENU006–REQMENU009, REQITEM001, REQITEM003–REQITEM005)
- [x] 4.2 Implement menu repository: create/list (status filter)/get-with-tree (live join)/update/deactivate/publish/unpublish (including weekly start_date uniqueness check) (REQMENU001–REQMENU004, REQMENU007–REQMENU009)
- [x] 4.3 Implement menu-item repository: add/update/remove items; manage components and component size options/defaults; enforce tenant and active-draft-menu checks (REQITEM001–REQITEM006)
- [x] 4.4 Add repository tests for tenant isolation, combo vs simple rules, per-component multi size options, defaults, inactive/published mutation rejection (REQITEM001, REQITEM003, REQITEM005, REQITEM006)

## 5. Service and handlers

- [x] 5.1 Implement menu service: draft create, type-metadata optional on draft, category sequencing, structural publish gate (component sizes + defaults), weekly start_date uniqueness at publish, unpublish, published immutability (REQMENU001, REQMENU004–REQMENU009)
- [x] 5.2 Implement menu-item service: simple/combo rules, component size-option pricing, default validation, size-unit reference validation, draft-only mutations, sum-of-defaults for combos (REQITEM001–REQITEM006, REQSIZE003)
- [x] 5.3 Implement HTTP handlers for menus, publish/unpublish, nested items/components/size-options; wire tenant middleware (REQMENU001–REQMENU004, REQMENU007–REQMENU009, REQITEM001–REQITEM005)
- [x] 5.4 Register routes and DI in `internal/server` / bootstrap; align or replace weekly/catering stubs
- [x] 5.5 Add handler/service tests for draft/publish/unpublish, published weekly start_date conflict, published edit rejection, per-component multi-size, combo sum-of-defaults, type validation at publish, live-join embed, cross-tenant 404, negative price, and custom size units

## 6. Traceability and docs

- [x] 6.1 Annotate production code and tests with REQMENU*, REQITEM*, REQSIZE*, REQFOOD001
- [x] 6.2 Update API/docs that describe food-item price or menu placeholders to match Menu + size-unit APIs
- [x] 6.3 Run focused package tests, then `go test ./...`, and fix regressions
