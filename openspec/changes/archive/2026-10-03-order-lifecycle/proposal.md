## Why

Cooks need a real kitchen workflow: accept or refuse new orders (e.g. catering/weekly queue full), then prepare, mark ready, and complete pickup—without free-form status jumps. Order intake shipped statuses without a transition graph and used `COMPLETE` for “ready,” which does not match the kitchen vocabulary. Notifications will depend on these transitions; this change establishes the lifecycle first.

## What Changes

- **BREAKING**: Rename order status `COMPLETE` → `READY` (API, DB constraint, specs, tests).
- Add statuses `ACCEPTED` and `DECLINED`.
- Enforce a transition graph (no skips/reversals except existing post-`PICKEDUP` `pickedup_at` correction).
- Distinct cook actions: accept (ack), refuse (with optional reason; default `"No available slots"`), start preparing, mark ready, mark picked up.
- Refuse → `DECLINED` is terminal; reject payments and line mutations on declined orders.
- Lines remain editable after `ACCEPTED` until `PICKEDUP` (unchanged money freeze at pickup).
- Live pricing “unfulfilled” set becomes `RECEIVED | ACCEPTED | IN_PROGRESS | READY`.
- Explicit non-goal: SMS/notification delivery (separate change `order-notifications`).

## Capabilities

### New Capabilities
- `order-lifecycle`: Transition guards, cook actions (accept/refuse/prepare/ready/pickup), refuse reason, terminal `DECLINED` rules.

### Modified Capabilities
- `order-intake`: Status value set (`ACCEPTED`, `DECLINED`, `READY` replacing `COMPLETE`); update rules aligned with the graph; unfulfilled pricing statuses.
- `order-payment`: Reject recording payments when the order is `DECLINED`.
- `order-line`: Reject add/update/remove when the order is `DECLINED` (still allowed after `ACCEPTED` while not `PICKEDUP`).

## Impact

- DB: migrate `customer_order.status` check constraint; rename any `COMPLETE` rows to `READY`.
- Go models/constants, order service update path, payment/line guards, tests, acceptance Gherkin.
- API: prefer dedicated action routes under `/api/v1/orders/:id/...` (or guarded status updates—design decides); clients using `COMPLETE` must switch to `READY`.
- Depends on archived `order-intake` behavior; unlocks follow-on `order-notifications`.
