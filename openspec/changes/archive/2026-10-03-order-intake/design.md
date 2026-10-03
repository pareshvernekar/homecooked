## Context

Published daily/weekly/catering menus with per-component size options and sellable prices exist (`tenant-menu`, `menu-item`, `size-unit`). Order models/DTOs are stubs with obsolete status values and `user_id`. This change adds real order intake and a separate payment ledger against published daily and catering menus only.

## Goals / Non-Goals

**Goals:**

- Tenant-scoped order CRUD with customer name + phone, timing fields, status values, and dual-level text customizations.
- Lines with per-component size selection constrained to the published menu; live pricing until PICKEDUP; line and order price overrides.
- Separate payment records with multi-mode support and derived paid/balance/overpay.
- Layering: handler → service → repository; composite `(tenant_id, id)` keys; REQ* traceability.

**Non-Goals:**

- Weekly-menu orders.
- Delivery logistics ( `pickedup_at` stands in for handoff time).
- Status transition state machine / role gates.
- Payment provider integrations or refunds.
- Pending-payment reminder tracker (later spec).
- Customer accounts / `user_id`.
- Tax, tips as first-class entities, multi-currency.

## Decisions

### 1. Order binds to published daily|catering menu

Create validates `menu_id` is active, `status=published`, and `menu_type` ∈ {`daily`, `catering`}. Lines MUST reference menu-items on that menu. Cross-tenant or unpublished menus → 404/400.

**Alternatives considered:** Allow draft menus for quoting — rejected; commercial prices only from published menus.

### 2. Line shape = menu-item + per-component size_option_id

```
order_item
  ├── menu_item_id, quantity, customization_text?, unit_price_override?
  └── selections[]: menu_item_component_id → size_option_id
```

Omitted selection → component’s default size option. Reject unknown option ids or options not belonging to that component on the bound menu.

**Rationale:** Matches archived menu “future contract”; same shape for simple (1 component) and combo (N).

### 3. Live price until PICKEDUP; freeze on pickup

**Unfulfilled** (`RECEIVED` | `IN_PROGRESS` | `COMPLETE`):

1. Line unit = `unit_price_override` if set, else Σ current prices of selected size options.
2. Subtotal = Σ (line unit × qty).
3. Charged total = `total_override` if set, else subtotal.

**On transition to `PICKEDUP`:** Persist `frozen_total` and per-line frozen unit/extended amounts from the resolution above; set `pickedup_at` (now if unset). Subsequent menu price changes MUST NOT alter charged total. Overrides after freeze rejected.

**Rationale:** Cook may adjust published prices for open tickets; history stable after handoff.

**Alternatives considered:** Snapshot at create — rejected; freeze only at PICKEDUP — chosen.

### 4. Overrides for customer satisfaction

Both line `unit_price_override` and order `total_override` (nullable). Order override wins when set. Clear by setting null. Only while not PICKEDUP.

### 5. Payments as separate 1:N ledger

Table `order_payment`: mode ∈ {`cash`,`credit`,`paypal`,`zelle`,`venmo`}, amount > 0, optional reference text. Aggregate on read:

- `paid_amount` = Σ amounts
- `balance` = charged_total − paid_amount
- `payment_received` = balance ≤ 0
- `overpaid_amount` = max(0, paid − charged)

Allow overpay (hybrid D). No refund entity. PICKEDUP allowed when unpaid. Payments allowed anytime after order exists.

### 6. Header timing and customer

| Field | Rule |
|-------|------|
| `customer_name`, `customer_phone` | Required |
| `received_at` | Default now; staff may set |
| `expected_at` | Required on create |
| `pickedup_at` | Set/correctable; auto-now when status → PICKEDUP if null |
| `customization_text` | Optional order-level note |
| `status` | Default `RECEIVED`; free update among the four values (no transition graph this change) |

### 7. API surface

- `POST/GET/PATCH /api/v1/orders`, `GET /api/v1/orders/:id`
- Nested lines: create/update/delete under order while not PICKEDUP (except read)
- `POST/GET /api/v1/orders/:id/payments` (and get-by-id as needed)

Replace stub DTOs; wire DI like menu handlers.

## Risks / Trade-offs

- **[Live price vs open tickets]** → Cooks changing published prices move unpaid/unpicked totals; mitigate with overrides and freeze at pickup.
- **[No transition guards]** → Invalid status jumps possible; deferred lifecycle change will tighten.
- **[Unpaid PICKEDUP]** → Operational debt; reminder tracker later.
- **[Published menu price edits]** → Today published menus are structurally immutable; price-only updates may require allowing size-option price PATCH on published menus or unpublish→edit→republish. Prefer **allowing price-only updates on published size options** so live order totals work without unpublishing. Document as companion behavior in design/tasks; if existing immutability blocks it, add a narrow exception in menu-item service for price fields only.

## Migration Plan

1. Add migration `000005_order_intake` (orders, items, selections, payments, freeze columns).
2. Sync `scripts/init/` and `tests/db/schema.sql`.
3. Deploy API; no backfill (no production orders yet).
4. Rollback: migrate down drops new tables (safe if empty).

## Open Questions

None for v1 — deferred: status lifecycle, payment reminders, delivery, refunds, weekly orders.
