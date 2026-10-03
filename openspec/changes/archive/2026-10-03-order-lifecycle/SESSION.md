# Session: order-lifecycle

**Change:** `openspec/changes/order-lifecycle/`  
**Branch:** `feature/order-lifecycle`  
**Status:** Implementation complete (11/11 tasks). Ready to archive after review. Apply `order-notifications` next on its branch.

## Locked decisions

- Statuses: RECEIVED → ACCEPTED → IN_PROGRESS → READY → PICKEDUP; RECEIVED → DECLINED
- BREAKING: COMPLETE → READY
- Actions: accept / refuse / start-preparing / ready / pickup (not PATCH status)
- Refuse reason optional; default `"No available slots"`
- Lines editable after ACCEPTED; no payments/lines on DECLINED
- No SMS in this change

## Apply

`/opsx:apply order-lifecycle`
