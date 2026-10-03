## 1. Schema

- [x] 1.1 Add migration: rename COMPLETE→READY data, expand status CHECK for ACCEPTED/DECLINED/READY, add `refuse_reason` (REQLIFE003, REQORDER004)
- [x] 1.2 Sync `scripts/init/` and `tests/db/schema.sql`; update DB constraint tests for new statuses and reject COMPLETE

## 2. Models and service lifecycle

- [x] 2.1 Update order status constants (`ACCEPTED`, `DECLINED`, `READY`); remove `COMPLETE`; add refuse_reason field (REQORDER004)
- [x] 2.2 Implement transition graph and action methods: Accept, Refuse, StartPreparing, Ready, Pickup (REQLIFE001–REQLIFE004)
- [x] 2.3 Reject `status` on header PATCH; reject header money overrides on DECLINED (REQLIFE005, REQORDER003)
- [x] 2.4 Reject payments and line add/update/remove on DECLINED; allow lines after ACCEPTED (REQPAY001, REQOLINE001, REQOLINE004)

## 3. HTTP API

- [x] 3.1 Add routes POST accept/refuse/start-preparing/ready/pickup under `/api/v1/orders/:id` (REQLIFE002–REQLIFE004)
- [x] 3.2 Wire handlers/DI; update CLAUDE.md Orders API section for lifecycle actions and READY rename

## 4. Tests and docs

- [x] 4.1 Service tests for happy path graph, skip/illegal transitions, refuse default/custom reason, DECLINED payment/line rejects (REQLIFE001–REQLIFE005, REQPAY001S04)
- [x] 4.2 Update existing order tests/acceptance that used COMPLETE or free-form PATCH status to use READY and action routes
- [x] 4.3 Annotate with REQLIFE*/updated REQ*; run `go test ./...`
