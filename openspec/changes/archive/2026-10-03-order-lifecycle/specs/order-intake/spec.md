## MODIFIED Requirements

### Requirement: REQORDER003 Update order header
The system SHALL allow updating customer fields, received_at, expected_at, pickedup_at, order-level customization_text, and total_override on an active order that is not `PICKEDUP` and not `DECLINED`, subject to money-override rules. Status MUST NOT be changed via this header update (see order-lifecycle). After `PICKEDUP`, header money overrides MUST be rejected; pickedup_at MAY still be corrected. After `DECLINED`, header money overrides and status-affecting updates MUST be rejected; get/list of the order remain allowed.

#### Scenario: REQORDER003S01 Update expected_at and customization
- **GIVEN** an active order with status RECEIVED or ACCEPTED
- **WHEN** the client updates expected_at and order customization_text
- **THEN** the system returns success and subsequent get reflects the changes

#### Scenario: REQORDER003S02 Set total_override
- **GIVEN** an unfulfilled active order (RECEIVED, ACCEPTED, IN_PROGRESS, or READY)
- **WHEN** the client sets total_override to a non-negative amount
- **THEN** charged_total equals total_override until cleared or the order is picked up

#### Scenario: REQORDER003S03 Reject total_override after PICKEDUP
- **GIVEN** an order with status PICKEDUP
- **WHEN** the client attempts to set or clear total_override
- **THEN** the system returns HTTP 400 or 409

#### Scenario: REQORDER003S04 Reject total_override after DECLINED
- **GIVEN** an order with status DECLINED
- **WHEN** the client attempts to set or clear total_override
- **THEN** the system returns HTTP 400 or 409

### Requirement: REQORDER004 Status values and freeze at PICKEDUP
The system SHALL support status values `RECEIVED`, `ACCEPTED`, `DECLINED`, `IN_PROGRESS`, `READY`, and `PICKEDUP`. New orders MUST start as `RECEIVED`. The legacy status value `COMPLETE` MUST NOT be accepted. Status transitions MUST follow the order-lifecycle capability. When status becomes `PICKEDUP`, the system MUST persist the charged total and per-line frozen amounts from the pricing resolution in effect at that moment, set `pickedup_at` to now if unset, and thereafter ignore live menu price changes for that order's charged_total. Unfulfilled (live pricing) statuses are `RECEIVED`, `ACCEPTED`, `IN_PROGRESS`, and `READY`.

#### Scenario: REQORDER004S01 Default status RECEIVED
- **GIVEN** a successful order create
- **WHEN** the client inspects the created order
- **THEN** status is RECEIVED

#### Scenario: REQORDER004S02 Freeze on PICKEDUP
- **GIVEN** an unfulfilled order whose charged_total is T from live or override pricing and status READY
- **WHEN** the client picks up the order via the lifecycle pickup action
- **THEN** the order stores frozen charged_total T, pickedup_at is set, and later menu size-option price changes do not change charged_total

#### Scenario: REQORDER004S03 Unpaid PICKEDUP allowed
- **GIVEN** an order with status READY and balance greater than zero
- **WHEN** the client picks up the order
- **THEN** the system accepts the update and payment_received remains false

#### Scenario: REQORDER004S04 Reject COMPLETE status value
- **GIVEN** any active order
- **WHEN** a client attempts to use status value COMPLETE
- **THEN** the system rejects the request and does not set status to COMPLETE
