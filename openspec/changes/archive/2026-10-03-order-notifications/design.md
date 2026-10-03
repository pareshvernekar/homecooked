## Context

`order-lifecycle` defines accept/refuse/prepare/ready/pickup and statuses including `ACCEPTED`, `DECLINED`, `READY`. This change adds async SMS-oriented notifications without choosing a commercial SMS vendor yet. Obsolete `internal/models/notification.go` assumes `user_id` and delivery statuses that do not match name+phone orders.

**Prerequisite:** Apply `order-lifecycle` before implementing this change.

## Goals / Non-Goals

**Goals:**

- Transactional outbox: status/order write and outbox insert in one DB transaction.
- Async worker delivers pending outbox rows via `SmsProvider`.
- Local provider for dev/tests (log or sink table).
- Tenant `cook_admin_phone` for cook alerts.
- Events: `order.created`→cook; `order.accepted`, `order.declined`, `order.ready`, `order.picked_up`→customer.
- At-least-once delivery with retries and dead-letter after N failures.
- Extensible event/channel/template keys for later channels.

**Non-Goals:**

- Real Twilio/SNS integration (interface only + local impl).
- Preparing SMS.
- Payment reminders, email, push, WhatsApp.
- Customer accounts / opt-in registry (use `customer_phone` on order).
- Guaranteeing exactly-once SMS at the carrier.

## Decisions

### 1. Transactional outbox in Postgres

```
BEGIN
  update customer_order / insert order
  insert notification_outbox (pending)
COMMIT
──▶ worker claims pending ──▶ SmsProvider.Send ──▶ mark delivered | retry | dead
```

**Rationale:** Survives process crash; order API stays fast.  
**Alternatives:** Sync SMS in request — rejected (provider latency/failures). Fire-and-forget goroutine — rejected (no guarantee).

### 2. Outbox row shape (conceptual)

- `id`, `tenant_id`, `order_id`, `event_type`, `channel` (`sms`), `recipient_phone`, `body` (or template_key + payload JSON)
- `status`: `pending` | `processing` | `delivered` | `dead`
- `attempts`, `next_attempt_at`, `last_error`, `provider_ref`, timestamps
- Unique or advisory idempotency: one successful enqueue per `(order_id, event_type)` for v1 (re-transitions shouldn’t exist for these events)

### 3. SmsProvider interface

```go
type SmsProvider interface {
  Send(ctx context.Context, msg SmsMessage) (providerRef string, err error)
}
```

- `LocalSmsProvider`: append to structured log and/or `sms_dev_sink` table for tests.
- Config selects provider via env (`SMS_PROVIDER=local`).

### 4. Tenant cook admin phone

Add `cook_admin_phone` on `tenant` (nullable VARCHAR).  
API: PATCH/GET tenant settings (minimal) or admin-only endpoint—keep thin: e.g. `PUT /api/v1/tenant/settings` with `{ "cook_admin_phone": "..." }` for current tenant.

If phone missing on `order.created`: still insert outbox row that fails to `dead` with clear error, **or** skip enqueue and log—prefer **enqueue + dead** so ops can see missed alerts. Design choice: **skip enqueue with structured warn** if no cook phone (avoid DLQ noise); document that cook phone must be configured for alerts.

**Decision:** If `cook_admin_phone` empty, do not enqueue `order.created`; log warning. Customer events always use order `customer_phone`.

### 5. Wire points (no lifecycle redesign)

Enqueue after successful:

| Hook | Event | Recipient |
|------|-------|-----------|
| Order create | `order.created` | cook_admin_phone |
| Accept | `order.accepted` | customer_phone |
| Refuse | `order.declined` (include refuse_reason in body) | customer_phone |
| Ready | `order.ready` | customer_phone |
| Pickup | `order.picked_up` | customer_phone |

No enqueue on start-preparing.

### 6. Worker

- Local: `cmd/notification-worker` or goroutine started from main behind `NOTIFICATION_WORKER=1`.
- Claim with `FOR UPDATE SKIP LOCKED`, backoff on failure, max attempts then `dead`.
- Acceptance/integration tests can run a single drain cycle against the local provider.

### 7. Replace stubs

Delete/replace obsolete Notification DTOs; do not expose a generic CRUD `POST /notifications` unless needed for admin list of outbox—optional `GET /api/v1/orders/:id/notifications` for delivery audit (nice-to-have; include if cheap).

## Risks / Trade-offs

- **[Lifecycle not applied first]** → Tasks state prerequisite; CI fails clearly.
- **[Missing cook phone]** → Silent skip of create alert; mitigate with config validation warning on settings GET.
- **[At-least-once duplicates]** → Idempotent enqueue per event; provider may still double-send on crash after send—accept for v1.
- **[Worker ops]** → Local-only runner initially; production deploy of worker is follow-up.

## Migration Plan

1. Migration for `cook_admin_phone`, `notification_outbox` (+ optional sink).
2. Sync init/schema.sql.
3. Deploy API with enqueue; run worker alongside `make local`.
4. Rollback: stop worker; migrate down drops outbox (safe if empty).

## Open Questions

None for v1 — SMS vendor choice deferred until a later change swaps the provider impl.
