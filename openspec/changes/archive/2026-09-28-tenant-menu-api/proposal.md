## Why

Tenants need to publish sellable menus (daily, weekly, or catering) that group catalog food items into categories, support combo (multi-item) offerings, and price by per-component nested size options. Today price lives on the food-item catalog, which blocks menu-specific and size-specific pricing and conflates catalog identity with commercial presentation. Menus also need a draft → publish lifecycle so incomplete menus can be built safely without appearing as sellable.

## What Changes

- Add a tenant-scoped **Menu** API with full CRUD, where delete deactivates the menu (soft delete via `is_active`).
- Menus have `status` of `draft` or `published`. Create starts as draft; publish requires structural completeness; published menus are read-only until unpublished back to draft.
- For weekly menus, at most one active published menu may share a given `start_date` per tenant (enforced at publish; drafts may share).
- Menus have a **type**: `daily`, `weekly`, or `catering`, with type-appropriate metadata (required at publish, optional on draft create).
- Structure each menu into **categories**, each containing **menu-items**.
- Menu-items are either **simple** (one component) or **combo** (multiple components). Both use the same shape: components with nested **size_options** and a required **default** size.
- Sellable price lives on the component size option (`size_unit` + `qty` + `price`). Combo/simple get responses MUST include computed `default_total` (= sum of component default prices); no package discount.
- At menu-build time the cook sets allowed sizes and defaults; order-time size changes among those options are a future contract (price = sum of selected options).
- Get-menu **live-joins** food-item display fields on components (no catalog price; reflects current catalog).
- **Size units** are a hybrid: system **standard enums** plus **tenant-defined** custom units.
- Support updating a draft menu by adding/removing menu-items, components, and size options.
- **BREAKING**: Remove price from the food-item catalog model, APIs, and schema.

## Capabilities

### New Capabilities
- `tenant-menu`: Tenant-scoped menus of type daily/weekly/catering — create (draft), read, update (draft only), publish, unpublish, deactivate; organized into categories.
- `menu-item`: Menu-item lifecycle on draft menus — simple or combo, components with nested size options and defaults, add/remove/update.
- `size-unit`: Standard system size units plus tenant-defined custom units used by size options.
- `food-item-catalog`: Food items are catalog definitions without price; create/update/list responses no longer accept or return price.

### Modified Capabilities

<!-- No existing OpenSpec specs under openspec/specs/. -->

## Impact

- **APIs**: New `/api/v1/menus` (nested categories/items/components/size-options; publish/unpublish); size-unit list/create for tenant customs; **BREAKING** change to `/api/v1/food-items` (no `price`).
- **Schema**: New `menu` (with `status`), `menu_category`, `menu_item`, `menu_item_component`, `menu_item_component_size_option`, `size_unit` tables; migration to drop `food_item.price`.
- **Code**: New models, repositories, services, handlers; food-item price removal; route/DI wiring; align or replace weekly/catering stubs.
- **Tests**: Coverage for draft/publish/unpublish, menu types, combos, component size options, sum-of-defaults pricing, size units, tenant isolation, soft delete, published immutability, and food-item price removal.
