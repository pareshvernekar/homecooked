## 0. Prerequisite

- [x] 0.1 Confirm `order-lifecycle` is implemented (or apply it first): ACCEPTED/DECLINED/READY actions and refuse_reason exist

## 1. Schema and tenant settings

- [x] 1.1 Migration: `tenant.cook_admin_phone`; `notification_outbox` (+ optional `sms_dev_sink`) (REQNOTIF001, REQNOTIF002, REQNOTIF003)
- [x] 1.2 Sync `scripts/init/` and `tests/db/schema.sql`; DB tests for outbox constraints
- [x] 1.3 Tenant settings GET/PUT for cook_admin_phone (REQNOTIF001)

## 2. Outbox and provider

- [x] 2.1 Replace obsolete notification model stubs with outbox/event models (REQNOTIF002)
- [x] 2.2 Implement SmsProvider interface + LocalSmsProvider (REQNOTIF004)
- [x] 2.3 Implement outbox repository: insert pending, claim, mark delivered/retry/dead (REQNOTIF003)
- [x] 2.4 Implement worker delivery cycle (cmd or gated goroutine) (REQNOTIF003, REQNOTIF004)

## 3. Wire enqueue to order hooks

- [x] 3.1 Enqueue `order.created` on create when cook phone set; skip when unset (REQNOTIF002S01–S02)
- [x] 3.2 Enqueue accept/decline/ready/pickup; never on start-preparing; same TX as status write (REQNOTIF002)
- [x] 3.3 Decline payload includes refuse_reason (REQNOTIF002S04)

## 4. API and tests

- [x] 4.1 GET notifications for order (REQNOTIF005); wire routes/DI; update CLAUDE.md
- [x] 4.2 Service/integration tests for enqueue matrix, skip cook phone, preparing no-op, worker deliver/dead (REQNOTIF002–REQNOTIF004)
- [x] 4.3 Annotate REQNOTIF*; run `go test ./...`
- [x] 4.4 Godog acceptance feature tests for cook phone settings, enqueue matrix, decline reason, audit list, worker delivery (REQNOTIF001–REQNOTIF005)
