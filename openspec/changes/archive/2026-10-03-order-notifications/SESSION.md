# Session: order-notifications

**Change:** `openspec/changes/order-notifications/`  
**Branch:** `feature/order-notifications` (stacked on `feature/order-lifecycle`)  
**Status:** Artifacts complete. **Requires `order-lifecycle` applied first.**

## Locked decisions

- Async outbox + worker; local SmsProvider; vendor TBD later
- Cook admin phone per tenant; skip create SMS if unset
- Events: created→cook; accepted/declined/ready/picked_up→customer
- No preparing SMS
- Decline body includes refuse_reason

## Apply

`/opsx:apply order-notifications` (after lifecycle)
