# Session: order-intake

**Saved:** 2026-09-27  
**Change:** `openspec/changes/order-intake/`  
**Status:** Implementation complete (20/20 tasks). Ready to archive.  
**Plan:** Cursor plan “Order Intake System” (explore → propose → apply).

## How to resume

1. Read this file + `proposal.md` / `design.md` / `tasks.md`.
2. Implementation applied — archive with `/opsx:archive` when ready.

Do **not** re-litigate locked decisions below unless the user explicitly reopens them.

---

## Locked decisions (from explore)

### Menu binding and lines
- Orders only against **published** `daily` or `catering` menus (weekly out of scope).
- Line = menu-item + **per-component** `size_option_id` from that published menu only.
- Omitted selection → component **default** size option.
- No create-time price snapshot; live menu option prices while unfulfilled.

### Pricing
- Unfulfilled (`RECEIVED` | `IN_PROGRESS` | `COMPLETE`): live prices.
- Line `unit_price_override` and order `total_override`; **order override wins** when set.
- **Freeze at `PICKEDUP`**: persist charged total + line breakdown; later menu price changes ignored.
- Companion: **price-only** updates allowed on size options of **published** menus (structural edits still draft-only) — REQITEM006 modified.

### Payments (separate capability)
- 1:N payments: `cash` | `credit` | `paypal` | `zelle` | `venmo` + amount + optional text ref.
- Derived: `paid_amount`, `balance`, `payment_received`, `overpaid_amount`.
- Partial/split OK; **hybrid overpay (D)** — allow overage as credit; **no refunds in v1**.
- **PICKEDUP allowed while unpaid**; payment reminder tracker → **later spec**.
- Payments allowed anytime after order exists (including after pickup).

### Header
- `customer_name` + `customer_phone` required (no `user_id` this phase).
- `received_at` — when taken (default now; staff may set/backdate).
- `expected_at` — when ready; **required on create**.
- `pickedup_at` — pickup time (doubles as delivery until delivery exists); set on PICKEDUP.
- Customizations: **both** order-level and line-level free text.
- Status values only (no transition state machine this change); default `RECEIVED`.

### Explicit non-goals (this change)
- Weekly orders, delivery logistics, PSP integration, refunds, payment reminders, status lifecycle rules, customer accounts.

---

## Specs / REQ IDs

| Capability | IDs |
|------------|-----|
| `order-intake` | REQORDER001–005 |
| `order-line` | REQOLINE001–004 |
| `order-payment` | REQPAY001–004 |
| `menu-item` (delta) | MODIFIED REQITEM005, REQITEM006 |

## Tasks progress

All tasks in `tasks.md` are `- [x]` (20/20).

## Context links

- Prior menu work archived: `openspec/changes/archive/2026-09-28-tenant-menu-api/`
- Main specs already include tenant-menu, menu-item, size-unit, food-item-catalog
- Stub to replace: `internal/models/order.go`, order DTOs in `internal/models/dtos.go`
