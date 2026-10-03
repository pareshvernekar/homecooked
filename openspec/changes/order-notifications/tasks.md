## 0. Prerequisite

- [ ] 0.1 Confirm `order-lifecycle` is implemented (or apply it first): ACCEPTED/DECLINED/READY actions and refuse_reason exist

## 1. Schema and tenant settings

- [ ] 1.1 Migration: `tenant.cook_admin_phone`; `notification_outbox` (+ optional `sms_dev_sink`) (REQNOTIF001, REQNOTIF002, REQNOTIF003)
- [ ] 1.2 Sync `scripts/init/` and `tests/db/schema.sql`; DB tests for outbox constraints
- [ ] 1.3 Tenant settings GET/PUT for cook_admin_phone (REQNOTIF001)

## 2. Outbox and provider

- [ ] 2.1 Replace obsolete notification model stubs with outbox/event models (REQNOTIF002)
- [ ] 2.2 Implement SmsProvider interface + LocalSmsProvider (REQNOTIF004)
- [ ] 2.3 Implement outbox repository: insert pending, claim, mark delivered/retry/dead (REQNOTIF003)
- [ ] 2.4 Implement worker delivery cycle (cmd or gated goroutine) (REQNOTIF003, REQNOTIF004)

## 3. Wire enqueue to order hooks

- [ ] 3.1 Enqueue `order.created` on create when cook phone set; skip when unset (REQNOTIF002S01–S02)
- [ ] 3.2 Enqueue accept/decline/ready/pickup; never on start-preparing; same TX as status write (REQNOTIF002)
- [ ] 3.3 Decline payload includes refuse_reason (REQNOTIF002S04)

## 4. API and tests

- [ ] 4.1 GET notifications for order (REQNOTIF005); wire routes/DI; update CLAUDE.md
- [ ] 4.2 Service/integration tests for enqueue matrix, skip cook phone, preparing no-op, worker deliver/dead (REQNOTIF002–REQNOTIF004)
- [ ] 4.3 Annotate REQNOTIF*; run `go test ./...`
