# order-lifecycle Specification

## Purpose
TBD - created by archiving change order-lifecycle. Update Purpose after archive.
## Requirements
### Requirement: REQLIFE001 Enforced order status transitions
The system SHALL allow order status changes only through the following transitions: `RECEIVED`→`ACCEPTED` (accept), `RECEIVED`→`DECLINED` (refuse), `ACCEPTED`→`IN_PROGRESS` (start preparing), `IN_PROGRESS`→`READY` (ready), `READY`→`PICKEDUP` (pickup). Any other status transition MUST be rejected. `DECLINED` and `PICKEDUP` MUST be terminal for status changes (except that after `PICKEDUP`, `pickedup_at` MAY still be corrected via header update without changing status).

#### Scenario: REQLIFE001S01 Accept then prepare then ready then pickup
- **GIVEN** an active order with status RECEIVED
- **WHEN** the client accepts, then starts preparing, then marks ready, then picks up
- **THEN** the order status is successively ACCEPTED, IN_PROGRESS, READY, and PICKEDUP

#### Scenario: REQLIFE001S02 Reject skip from RECEIVED to READY
- **GIVEN** an active order with status RECEIVED
- **WHEN** the client attempts to mark the order ready
- **THEN** the system returns HTTP 409 or 400 and status remains RECEIVED

#### Scenario: REQLIFE001S03 Reject transition out of DECLINED
- **GIVEN** an order with status DECLINED
- **WHEN** the client attempts accept, start preparing, ready, or pickup
- **THEN** the system returns HTTP 409 or 400 and status remains DECLINED

### Requirement: REQLIFE002 Accept order
The system SHALL provide an accept action that transitions an active order from `RECEIVED` to `ACCEPTED`. Accept from any other status MUST be rejected.

#### Scenario: REQLIFE002S01 Accept from RECEIVED
- **GIVEN** an active order with status RECEIVED
- **WHEN** the client accepts the order
- **THEN** the system returns success and status is ACCEPTED

### Requirement: REQLIFE003 Refuse order with reason
The system SHALL provide a refuse action that transitions an active order from `RECEIVED` to `DECLINED`. The client MAY supply a refuse reason string. When omitted or blank after trim, the system MUST store the default reason `No available slots`. The stored reason MUST appear on subsequent get of the order. Refuse from any status other than `RECEIVED` MUST be rejected.

#### Scenario: REQLIFE003S01 Refuse with default reason
- **GIVEN** an active order with status RECEIVED
- **WHEN** the client refuses the order without a reason
- **THEN** status is DECLINED and refuse_reason is "No available slots"

#### Scenario: REQLIFE003S02 Refuse with custom reason
- **GIVEN** an active order with status RECEIVED
- **WHEN** the client refuses with reason "Catering queue full"
- **THEN** status is DECLINED and refuse_reason is "Catering queue full"

### Requirement: REQLIFE004 Start preparing, ready, and pickup actions
The system SHALL provide start-preparing (`ACCEPTED`→`IN_PROGRESS`), ready (`IN_PROGRESS`→`READY`), and pickup (`READY`→`PICKEDUP`) actions. Pickup MUST apply the same freeze-at-pickup behavior required by order-intake (persist charged total and line frozen amounts; set `pickedup_at` if unset). Unpaid pickup MUST remain allowed.

#### Scenario: REQLIFE004S01 Start preparing from ACCEPTED
- **GIVEN** an active order with status ACCEPTED
- **WHEN** the client starts preparing
- **THEN** status is IN_PROGRESS

#### Scenario: REQLIFE004S02 Ready from IN_PROGRESS
- **GIVEN** an active order with status IN_PROGRESS
- **WHEN** the client marks the order ready
- **THEN** status is READY

#### Scenario: REQLIFE004S03 Pickup freezes totals
- **GIVEN** an active order with status READY and charged_total T
- **WHEN** the client marks the order picked up
- **THEN** status is PICKEDUP, frozen charged_total is T, and pickedup_at is set

### Requirement: REQLIFE005 Status changes only via lifecycle actions
The system SHALL reject attempts to change order status through the general order header update API. Status MUST change only via the accept, refuse, start-preparing, ready, and pickup actions defined by this capability.

#### Scenario: REQLIFE005S01 Reject status on header PATCH
- **GIVEN** an active order with status RECEIVED
- **WHEN** the client PATCHes the order header with status ACCEPTED or any other status
- **THEN** the system returns HTTP 400 and status remains RECEIVED

