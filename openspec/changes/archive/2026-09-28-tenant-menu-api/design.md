## Context

Food catalog APIs (`food_item`, `food_category`) exist with tenant composite keys and soft delete via `is_active`. Price currently lives on `food_item`. Stub models for weekly/catering menus and a string `size` on menu items exist but are not wired to schema or routes. This change introduces a unified tenant **Menu** with types, draft/publish lifecycle, combo menu-items, **per-component** nested size options, and hybrid size units.

## Goals / Non-Goals

**Goals:**

- Tenant-scoped menu CRUD where DELETE sets `is_active = false` (REQMENU001–REQMENU004).
- Menu lifecycle: `status` ∈ {`draft`, `published`} orthogonal to `is_active` (REQMENU007–REQMENU008).
- Menu types: `daily`, `weekly`, `catering` with type-specific metadata (REQMENU006).
- Menus structured as categories containing menu-items (REQMENU005).
- Menu-items are simple (1 component) or combo (N components); both use the same component + size-option shape (REQITEM001).
- Size options (1..N) and a required default live on each **component**; price on size option; multi-size allowed on all menu types (REQITEM003, REQITEM004).
- Combo default sellable price = sum of component default size prices; no package discount (REQITEM005).
- Size units: seeded standard enums + tenant-defined customs (REQSIZE001–REQSIZE003).
- Remove price from food-item catalog (REQFOOD001).
- Follow handler → service → repository layering and composite `(tenant_id, id)` keys.

**Non-Goals:**

- Order placement, notifications, or inventory deduction (order-time size selection is a future contract only).
- Choice-based combos (e.g. “pick any 2 sides”) — fixed component lists only; size choice is among predefined options per component.
- Package-level discounts or override totals on combos.
- Multi-currency or tax calculation.
- Public anonymous menu browsing without tenant context.
- Hard delete of menus or historical purge.
- In-place editing of published menus (must unpublish to draft first).

## Decisions

### 1. Unified menu with type discriminator

One `menu` table with `menu_type` ∈ {`daily`, `weekly`, `catering`}.

| Type | Metadata (nullable columns; required at publish) |
|------|--------------------------------------------------|
| `daily` | `menu_date` |
| `weekly` | `start_date`, `end_date` |
| `catering` | `event_date`, `event_location` |

**Rationale:** Shared categories/items/size-options engine; stubs’ separate WeeklyMenu/CateringMenu tables are superseded by one resource.

**Alternatives considered:** Separate APIs per type — more surface area for the same item engine.

### 2. Draft / publish lifecycle (`status` + `is_active`)

| Field | Values | Role |
|-------|--------|------|
| `status` | `draft` \| `published` | Commercial readiness |
| `is_active` | bool | Soft-delete / availability |

```
create ──► draft (is_active=true)     ← editable; type metadata optional
              │
              │  PUBLISH (structurally complete gate)
              ▼
           published                  ← read-only mutations
              │
     unpublish │    DELETE / deactivate
              ▼              │
            draft            ▼
         (editable)    is_active=false  (status retained; hidden from default list/get)
```

**Create:** Requires non-empty `name` and valid `menu_type`. Type-specific metadata is optional. Menu starts as `status=draft`, `is_active=true`.

**Publish gate (structurally complete):** All of the following MUST hold or publish returns HTTP 400 and status stays `draft`:

1. Type-required metadata present (daily: `menu_date`; weekly: `start_date`+`end_date`; catering: `event_date`+`event_location`).
2. At least one category.
3. At least one menu-item across categories.
4. Each menu-item: valid simple/combo component cardinality; every component references a same-tenant **active** food-item; every component has ≥1 size option and a valid **default** size option.

**Weekly published start_date uniqueness (REQMENU009):** Per tenant, at most one **active published** weekly menu may have a given `start_date`. Enforced at publish only. Drafts MAY share the same `start_date`. Daily and catering menus have no date uniqueness. Inactive menus do not hold the slot. Overlapping weekly date ranges with different start dates are allowed (out of scope).

**Published:** Metadata, categories, menu-items, components, and size options MUST NOT be mutated. Clients MUST unpublish (→ `draft`) to edit, or deactivate.

**Unpublish:** Active published menu → `status=draft` (structure retained; editable again).

**Deactivate:** `DELETE` sets `is_active=false` for draft or published; default list/get treat as unavailable (404).

**List defaults:** Active menus; default status filter is `published`. Clients MAY request drafts via `?status=draft` or both via `?status=all`.

**Alternatives considered:** Three-way enum (`draft|published|inactive`) — rejected in favor of orthogonal soft-delete matching food catalog. Light publish gate (metadata only) — rejected; published means sellable structure.

### 3. Data model — size options under components

```
menu
  └── menu_category
        └── menu_item (kind: simple | combo)
              └── menu_item_component (food_item_id, default_size_option_id)
                    └── menu_item_component_size_option (size_unit_id, qty, price)  -- 1..N

size_unit (system rows + tenant customs)
```

| Table | Role |
|-------|------|
| `menu` | Tenant menu; type + metadata; `status`; `is_active` |
| `menu_category` | Named section; `sequence`; `is_active` |
| `menu_item` | Sellable offering; `kind`; name/description; `sequence`; `is_active` — **no price, no size options** |
| `menu_item_component` | Links food-item into the offering; holds `default_size_option_id` |
| `menu_item_component_size_option` | Allowed size for that component: unit + qty + **price** |
| `size_unit` | Standard (`is_system`) or tenant custom |

**Constraints:**

- Simple: exactly one component.
- Combo: at least two components.
- Every component MUST have at least one active size option and exactly one default among them.
- Components reference same-tenant active food-items.
- Same food-item MAY appear alone and inside combos on the same menu (no unique `(menu, food_item)` on menu_item).
- Optional uniqueness: `(tenant_id, menu_item_id, food_item_id)` on components so a food-item appears once per offering.
- Partial unique index (or equivalent): one active published weekly menu per `(tenant_id, start_date)` (REQMENU009).

**Alternatives considered:** Menu-item-level size options with unpriced components — rejected; cook sets per-component sizes at build time and customers may change among allowed options at order time (future).

### 4. Price on component size_option; never on food_item or menu_item

Drop `food_item.price` (**BREAKING**). Menu-item has no price column and no size options. Components have no price column. Sellable price lives only on `menu_item_component_size_option`.

**Combo pricing:** Default display/sellable total = sum of each component's **default** size-option price. Exposed on get as computed `default_total` (not persisted). No package discount or override total. Simple items use the same field (= that component's default size price).

**Future orders (out of scope):** Order line references the menu-item and a selected `size_option_id` **per component** (defaulting to menu defaults). Order price = sum of selected size-option prices. Selections MUST be from that component's allowed options.

### 5. Nested size_options (all menu types)

Daily/weekly/catering may each expose multiple sizes per component. API nests size options under the component.

### 6. Size units: standard + tenant-defined

Seeded system units (read-only via API), at minimum:

`serving`, `tray`, `piece`, `dozen`, `kg`, `g`, `liter`

Tenants MAY create custom units (`is_system=false`, scoped by `tenant_id`). Size options MUST reference a system unit or a custom unit owned by the current tenant. Tenants MUST NOT delete or rename system units; customs MAY be soft-deactivated if unused.

**Alternatives considered:** Fixed enum only (too rigid); free string (weak validation/reporting).

### 7. Soft delete / deactivate

- `DELETE /menus/:id` → `menu.is_active = false` (draft or published).
- Remove menu-item / component / size option → soft-deactivate (draft menus only).
- Default list/get omit inactive; get inactive menu → 404.

### 8. Get-menu food-item embed (live join)

Get-menu returns components with a **live-joined** display slice of the current food-item (not a frozen snapshot):

- Include: `id`, `name`, dietary/display fields needed for UI (e.g. `is_vegetarian`, `avoidance`, `image_url` when present).
- Omit: catalog price (removed by REQFOOD001).
- Catalog edits (rename, dietary flags) appear on subsequent get-menu responses.
- Each component also includes its size options (with prices) and default size option indicator.
- Menu-item responses MUST include a computed `default_total` = sum of component default size-option prices (for simple items, that is the single component's default price). `default_total` is derived, not stored.

**Alternatives considered:** IDs only (forces N+1 client fetches); full frozen snapshot (orders/history concern; defer).

### 9. API shape

```
POST|GET          /api/v1/menus
GET|PUT|DELETE    /api/v1/menus/:id
POST              /api/v1/menus/:id/publish
POST              /api/v1/menus/:id/unpublish

POST|PUT|DELETE   /api/v1/menus/:id/items[/:itemId]
POST|PUT|DELETE   /api/v1/menus/:id/items/:itemId/components[/:componentId]
POST|PUT|DELETE   /api/v1/menus/:id/items/:itemId/components/:componentId/size-options[/:optionId]

GET               /api/v1/size-units          # system + tenant customs
POST              /api/v1/size-units          # create tenant custom
DELETE            /api/v1/size-units/:id      # soft-deactivate tenant custom only
```

Categories included in menu create/update and get-by-id payload. Components and size options MAY be nested in item create/update bodies. Tenant ID from middleware only. Item/component/size mutations and menu PUT require `status=draft` and `is_active=true`.

### 10. Packages

- `internal/models` — Menu*, SizeUnit; strip Price from FoodItem
- `internal/repository` — composite-key repos
- `internal/services/menu`, `sizeunit` (or equivalent)
- `internal/handlers`, `internal/server`
- Align `migrations/`, `scripts/init/`, `tests/db/schema.sql`
- Replace unused weekly/catering stub models carefully to avoid confusion

### 11. Validation & errors

- Size option price `>= 0`; qty `> 0`.
- At least one size option per component; default must reference one of that component's options.
- Simple vs combo component cardinality enforced.
- Inactive or published menu → no item/component/size/metadata mutations (404/conflict).
- Publish incomplete structure → 400.
- Publish weekly when another active published weekly already has the same `start_date` → 409/400 (REQMENU009).
- Standard `views.ErrorResponse` codes.

## Risks / Trade-offs

- **[BREAKING food-item API]** → Same-release client/test updates.
- **[Order line shape later]** → Orders need per-component `size_option_id`s (not a single menu-item size option); document as future contract.
- **[Stub collision]** → Migrate/replace `WeeklyMenu`/`CateringMenu`/`MenuItem` stubs in this change.
- **[Custom unit sprawl]** → Mitigate with standards in pickers and soft-deactivate for customs.
- **[Published immutability]** → Clients must unpublish to edit; clear API errors when mutating published menus.

## Migration Plan

1. Create `size_unit` + seed system units; create menu hierarchy tables including `status` and component-level size options.
2. Drop `food_item.price`; sync init/schema/test SQL.
3. Deploy app + API clients together.
4. Rollback: reverse migrations (restore nullable/`0` price if needed); only safe before heavy menu dependence.

## Open Questions

None — draft/publish, live-join embed, weekly start_date uniqueness, and per-component size/pricing settled (2026-09-27).
