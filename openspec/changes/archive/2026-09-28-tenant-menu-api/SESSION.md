# Session notes — tenant-menu-api

**Date:** 2026-09-27 (updated)  
**Branch context:** `feature/menu-creation`  
**Change:** `openspec/changes/tenant-menu-api/`  
**Status:** OpenSpec artifacts complete (proposal, design, specs, tasks). Implementation not started. Open questions closed.

Resume with `/opsx:apply` starting at task **1.1** (schema/migrations).

---

## What we did this session (2026-09-27)

1. Resumed from prior session; reviewed outstanding design questions.
2. **Closed Q1:** Draft/publish lifecycle — `status` (`draft`|`published`) + `is_active`; heavy publish gate (structurally complete); published is read-only; unpublish → draft; deactivate via DELETE.
3. **Closed Q2:** Get-menu food-item embed = **live join** of display fields (no catalog price).
4. Added **REQMENU009**: at most one active published weekly menu per tenant `start_date` (enforced at publish; drafts may collide).
5. **Revisited sizing/pricing:** size options move under **components** (not menu-item). Cook sets allowed sizes + default at build time; combo default total = sum of defaults (no package discount). Order-time size changes among allowed options = future contract.
6. Rewrote **REQITEM003–005** (and aligned REQITEM001/004/006, REQMENU003/007, design, proposal, tasks). Old `REQITEM005S01` (package price P, no per-component prices) **invalidated**.

---

## Prior session (2026-09-26)

1. **Proposed** `tenant-menu-api` via `/opsx:propose`.
2. **Explored** combo specials, menu types, nested size options, hybrid size units.
3. **Updated** OpenSpec artifacts for that model (later superseded on size nesting).

---

## Decisions locked

| Topic | Decision |
|-------|----------|
| Special | Combo = multiple food-item components (fixed list) |
| Simple vs combo | Same shape; differs only by component count (1 vs ≥2) |
| Where size/price live | On **component** size options; required **default** per component |
| Combo pricing | Sum of component **default** size prices; required computed `default_total` on get; no package discount |
| Order-time (future) | Customer may pick among allowed component size options; price = sum of selections |
| Menu types | `daily`, `weekly`, `catering` |
| Size units | Hybrid: system standards + tenant customs |
| Standard seed | `serving`, `tray`, `piece`, `dozen`, `kg`, `g`, `liter` |
| Food-item | No catalog price (**BREAKING**) — unrelated to menu sellable prices |
| Lifecycle | `status`: draft \| published; `is_active` soft-delete |
| Create | Always `draft`; type metadata optional |
| Publish gate | Structurally complete incl. per-component sizes + defaults |
| Published | **Read-only**; must unpublish to edit |
| Unpublish | Allowed → back to `draft` |
| Delete menu | Soft deactivate (`is_active=false`) |
| Get-menu embed | Live join food-item display fields (no catalog price) |
| Weekly start_date uniqueness | One active **published** weekly per tenant `start_date` (REQMENU009) |
| Choice combos | Out of scope (“pick any 2 sides”) |
| Orders | Out of scope (future: per-component `size_option_id`s on the line) |

---

## Target data model (summary)

```
menu (type; status; is_active)
  └── menu_category
        └── menu_item (kind: simple | combo)   ← no price, no size_options
              └── menu_item_component[] → food_item + default_size_option
                    └── size_options[] → size_unit + qty + price

size_unit = system standards + tenant customs
```

- **Simple:** exactly 1 component  
- **Combo:** ≥ 2 components; default total = Σ default size prices  

---

## Artifacts

| File | Purpose |
|------|---------|
| `proposal.md` | Why / what / capabilities |
| `design.md` | How — schema, API shape, decisions |
| `specs/tenant-menu/spec.md` | REQMENU001–009 |
| `specs/menu-item/spec.md` | REQITEM001–006 |
| `specs/size-unit/spec.md` | REQSIZE001–003 |
| `specs/food-item-catalog/spec.md` | REQFOOD001 |
| `tasks.md` | Implementation checklist |

---

## Requirement ID map

- **REQMENU001–006** — create (draft)/list/get (live join)/update (draft only)/deactivate, categories, menu types  
- **REQMENU007–008** — publish (structural gate incl. component sizes/defaults), unpublish  
- **REQMENU009** — unique active published weekly `start_date` per tenant  
- **REQITEM001–006** — simple/combo, remove item, **component** size options + default, price on option, sum-of-defaults combo pricing, draft-only mutations  
- **REQSIZE001–003** — standards, tenant customs, size-option unit references  
- **REQFOOD001** — food item catalog has no price  

---

## Open questions

None.

---

## Suggested next steps

1. Run `/opsx:apply` starting at task **1.1** (schema/migrations).
2. Schema table: `menu_item_component_size_option` (not menu-item-level size options).

---

## Conversation arc (short)

Explore → combo + menu types + size units → draft/publish + live join → weekly start_date uniqueness → per-component sized/priced components (sum-of-parts) → REQITEM005 rewritten; ready to apply.
