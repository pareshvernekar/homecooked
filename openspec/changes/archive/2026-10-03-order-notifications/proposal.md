## Why

After order lifecycle exists, cooks and customers still learn about order progress only by polling the API. Home cooks need SMS (or a local stand-in) when a new order arrives, and customers need updates when the cook accepts, refuses, marks ready, or completes pickup—without blocking HTTP on the SMS provider.

## What Changes

- Add a durable **notification outbox** with async worker delivery (at-least-once, retries).
- Add a **SmsProvider** interface with a **local/dev provider** (no cloud SMS vendor locked in this change).
- Per-tenant **cook admin phone** setting used for new-order alerts.
- Enqueue notifications on: order created (→ cook), accepted / declined / ready / picked up (→ customer).
- Omit preparing SMS in v1 (extension point only).
- Replace obsolete notification model stubs (`user_id`, `order_delivered`).
- **Depends on** `order-lifecycle` being applied first (statuses/actions: ACCEPTED, DECLINED, READY, pickup, refuse_reason).

## Capabilities

### New Capabilities
- `order-notifications`: Outbox, provider abstraction, tenant cook phone, event→recipient mapping, delivery guarantees and observability.

### Modified Capabilities
- _(none at requirement level for existing order specs — this change observes lifecycle transitions; apply `order-lifecycle` before this change)_

## Impact

- New tables: notification outbox (+ optional delivery log); tenant `cook_admin_phone` (or tenant_settings).
- Worker entrypoint or in-process poller for local/dev; later swappable to Lambda/queue.
- Wire enqueue into order create + lifecycle actions (same TX as status write).
- Tests: outbox rows on transitions; local provider sink assertions; retry behavior.
- Explicit non-goals: Twilio/SNS selection, email/push, preparing SMS, payment reminders, SMS opt-out compliance beyond storing destinations.
