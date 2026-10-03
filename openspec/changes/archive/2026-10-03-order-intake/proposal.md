## Why

Home cooks need to take advance orders against published daily and catering menus, track readiness timing, capture text customizations, and record payments separately—without coupling kitchen fulfillment to payment settlement. The catalog/menu layer is in place; order stubs do not match the agreed pricing and payment model.

## What Changes

- Add tenant-scoped **order intake** for published `daily` and `catering` menus (weekly orders out of scope).
- Order header: customer name + phone, `received_at`, required `expected_at`, optional `pickedup_at`, status values `RECEIVED` | `IN_PROGRESS` | `COMPLETE` | `PICKEDUP`, order-level customization text, optional order `total_override`.
- Order lines: menu-item reference, quantity, per-component size selection from published options (defaults if omitted), line customization text, optional line `unit_price_override`.
- Pricing: live menu size-option prices while unfulfilled; freeze charged total (and line breakdown) at `PICKEDUP`; order override wins over line math when set.
- Add separate **order payment** recording (cash, credit, paypal, zelle, venmo) with derived paid/balance/overpay; unpaid pickup allowed.
- Replace obsolete order model stubs (`pending|preparing|ready|delivered`, `user_id`, weekly-oriented fields).
- Status **transition rules** and pending-payment reminders deferred to later changes.

## Capabilities

### New Capabilities
- `order-intake`: Create/list/get/update orders; header fields; status values; freeze at PICKEDUP; order-level customization and total override.
- `order-line`: Lines with per-component size selection from published menu; line overrides; line customization text; contribution to charged total.
- `order-payment`: Multi-mode payments; derived payment_received/balance/overpaid; record anytime after order exists; unpaid PICKEDUP allowed.

### Modified Capabilities
- `menu-item`: Clarify that order-time size selection (previously “future contract”) is now in scope via `order-line`—no change to menu-build rules; add a note that published option prices remain the live source for unfulfilled orders.

## Impact

- New migrations/tables for orders, lines, component selections, payments; sync `scripts/init/` and `tests/db/schema.sql`.
- New handlers/routes under `/api/v1/orders` (+ nested lines and payments); DI in `cmd/main.go` / `internal/server`.
- Depends on existing published menu tree and size options (`tenant-menu`, `menu-item`, `size-unit`).
- Replaces `internal/models/order.go` and order DTOs; removes/updates placeholder order stubs.
- Tests: DB constraints, service, HTTP integration; acceptance optional in this change.
