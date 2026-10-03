## Context

Order intake supports `RECEIVED | IN_PROGRESS | COMPLETE | PICKEDUP` with free status updates and money freeze at `PICKEDUP`. Cooks need accept/refuse before cooking, a clear “ready for pickup” state, and illegal jumps blocked. Notifications are a follow-on change and must not be implemented here—only stable transitions and fields they will observe.

## Goals / Non-Goals

**Goals:**

- Status set: `RECEIVED`, `ACCEPTED`, `DECLINED`, `IN_PROGRESS`, `READY`, `PICKEDUP` (**BREAKING** rename `COMPLETE` → `READY`).
- Enforced transition graph via cook action endpoints.
- Refuse with optional `reason` (default `"No available slots"`), stored on the order.
- Terminal `DECLINED`: no payments, no line mutations; get/list still work.
- Lines editable after `ACCEPTED` while status ∈ unfulfilled set.
- Unfulfilled (live pricing) = `RECEIVED | ACCEPTED | IN_PROGRESS | READY`.
- Preserve freeze-at-`PICKEDUP` and unpaid pickup allowed.

**Non-Goals:**

- SMS / notification outbox (change `order-notifications`).
- Role-based auth for cook vs customer (tenant header remains as today).
- Capacity/queue management beyond refuse reason text.
- Reversing transitions (un-decline, un-accept).
- Payment reminders.

## Decisions

### 1. Explicit action routes (not free-form PATCH status)

```
POST /api/v1/orders/:id/accept
POST /api/v1/orders/:id/refuse   body: { "reason"?: string }
POST /api/v1/orders/:id/start-preparing
POST /api/v1/orders/:id/ready
POST /api/v1/orders/:id/pickup
```

`PATCH /orders/:id` continues for header fields (customer, times, customization, overrides) but **MUST reject** client-supplied `status` changes (or ignore status with 400)—lifecycle only via actions.

**Rationale:** Matches cook verbs; prevents inventing illegal edges.  
**Alternatives:** Guarded `PATCH status` — rejected; weaker UX and easier to misuse.

### 2. Transition graph

| From | Action | To |
|------|--------|-----|
| RECEIVED | accept | ACCEPTED |
| RECEIVED | refuse | DECLINED |
| ACCEPTED | start-preparing | IN_PROGRESS |
| IN_PROGRESS | ready | READY |
| READY | pickup | PICKEDUP |

Illegal action → HTTP 409 (or 400). `DECLINED` and `PICKEDUP` are terminal for status (except `pickedup_at` correction after pickup as today).

### 3. Refuse reason column

Add `refuse_reason TEXT` on `customer_order` (nullable). On refuse: set to trimmed client reason or default `"No available slots"`. Exposed on get. Cleared never required (terminal).

### 4. BREAKING status rename + constraint migration

- Migrate existing `COMPLETE` → `READY`.
- Replace CHECK constraint to include `ACCEPTED`, `DECLINED`, `READY` and drop `COMPLETE`.
- Update Go constants, validation messages, DB tests, acceptance.

### 5. DECLINED vs payments/lines

- `RecordPayment` → reject when status is `DECLINED`.
- Add/update/remove lines → reject when `DECLINED` or `PICKEDUP` (ACCEPTED+ still allowed).
- Header money overrides → reject on `DECLINED` and `PICKEDUP`.

### 6. Extension seam for notifications

Actions that change status SHOULD be implemented so a later notifier can observe success (e.g. single service method per action). No outbox table in this change.

## Risks / Trade-offs

- **[BREAKING READY rename]** → Document in proposal; migrate rows; update all tests in same change.
- **[Clients using PATCH status]** → Return clear 400; acceptance/docs show action routes.
- **[No auth on actions]** → Same tenant middleware as today; mitigate later with roles.
- **[Split from notifications]** → Lifecycle can ship alone; notifications change must not redefine the graph.

## Migration Plan

1. Migration `000006_order_lifecycle`: add `refuse_reason`; update status CHECK; `UPDATE ... SET status='READY' WHERE status='COMPLETE'`.
2. Sync `scripts/init/` and `tests/db/schema.sql`.
3. Deploy API with action routes; reject `status` on PATCH.
4. Rollback: reverse migration only if no `ACCEPTED`/`DECLINED` rows (or map them carefully)—prefer forward-only in early env.

## Open Questions

None — locked in explore. Provider/SMS deferred to `order-notifications`.
