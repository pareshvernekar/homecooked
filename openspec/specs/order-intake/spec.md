# order-intake Specification

## Purpose
TBD - created by archiving change order-intake. Update Purpose after archive.
## Requirements
### Requirement: REQORDER001 Create order against published daily or catering menu
The system SHALL allow creating an order for the current tenant against an active menu with `status=published` and `menu_type` of `daily` or `catering`. Create MUST require non-empty `customer_name`, non-empty `customer_phone`, and `expected_at`. `received_at` MUST default to the server time when omitted and MAY be supplied by the client. The created order MUST have `status=RECEIVED`, `is_active=true`, and be scoped to that tenant. Orders against weekly menus, draft menus, inactive menus, or other tenants' menus MUST be rejected.

#### Scenario: REQORDER001S01 Successful create
- **GIVEN** a valid tenant context and an active published daily or catering menu for that tenant
- **WHEN** the client creates an order with customer name, phone, expected_at, and menu_id
- **THEN** the system returns HTTP 201 with the order including id, tenant_id, menu_id, status=RECEIVED, received_at, expected_at, customer fields, and timestamps

#### Scenario: REQORDER001S02 Reject missing required header fields
- **GIVEN** a valid tenant context and a published daily or catering menu
- **WHEN** the client creates an order without customer_name, customer_phone, or expected_at
- **THEN** the system returns HTTP 400 with a validation error

#### Scenario: REQORDER001S03 Reject unpublished weekly or cross-tenant menu
- **GIVEN** a draft menu, a weekly published menu, an inactive menu, or a menu belonging to another tenant
- **WHEN** the client creates an order referencing that menu
- **THEN** the system returns HTTP 400 or 404 and does not create an order

### Requirement: REQORDER002 List and get orders
The system SHALL list and get active orders for the current tenant only. Get MUST include header fields, lines with selections and resolved prices, derived payment aggregates (`paid_amount`, `balance`, `payment_received`, `overpaid_amount`), and charged total per the pricing rules. Cross-tenant or inactive orders MUST NOT be returned as found.

#### Scenario: REQORDER002S01 List tenant orders
- **GIVEN** active orders for tenant A and tenant B
- **WHEN** tenant A lists orders
- **THEN** only tenant A's active orders appear

#### Scenario: REQORDER002S02 Get order with derived money fields
- **GIVEN** an active order with lines and payments for the tenant
- **WHEN** the client gets the order by id
- **THEN** the response includes charged_total, paid_amount, balance, payment_received, and overpaid_amount consistent with current pricing and payment rules

#### Scenario: REQORDER002S03 Not found cross-tenant or inactive
- **GIVEN** an order belonging to another tenant or deactivated
- **WHEN** the client gets that order id
- **THEN** the system returns HTTP 404

### Requirement: REQORDER003 Update order header
The system SHALL allow updating customer fields, received_at, expected_at, pickedup_at, order-level customization_text, total_override, and status on an active order that is not `PICKEDUP`, except that setting status to `PICKEDUP` is allowed and triggers freeze behavior. After `PICKEDUP`, header money overrides MUST be rejected; pickedup_at MAY still be corrected.

#### Scenario: REQORDER003S01 Update expected_at and customization
- **GIVEN** an active order with status RECEIVED
- **WHEN** the client updates expected_at and order customization_text
- **THEN** the system returns success and subsequent get reflects the changes

#### Scenario: REQORDER003S02 Set total_override
- **GIVEN** an unfulfilled active order
- **WHEN** the client sets total_override to a non-negative amount
- **THEN** charged_total equals total_override until cleared or the order is picked up

#### Scenario: REQORDER003S03 Reject total_override after PICKEDUP
- **GIVEN** an order with status PICKEDUP
- **WHEN** the client attempts to set or clear total_override
- **THEN** the system returns HTTP 400 or 409

### Requirement: REQORDER004 Status values and freeze at PICKEDUP
The system SHALL support status values `RECEIVED`, `IN_PROGRESS`, `COMPLETE`, and `PICKEDUP`. New orders MUST start as `RECEIVED`. When status becomes `PICKEDUP`, the system MUST persist the charged total and per-line frozen amounts from the pricing resolution in effect at that moment, set `pickedup_at` to now if unset, and thereafter ignore live menu price changes for that order's charged_total. This change does NOT require a restricted transition graph among the four statuses.

#### Scenario: REQORDER004S01 Default status RECEIVED
- **GIVEN** a successful order create
- **WHEN** the client inspects the created order
- **THEN** status is RECEIVED

#### Scenario: REQORDER004S02 Freeze on PICKEDUP
- **GIVEN** an unfulfilled order whose charged_total is T from live or override pricing
- **WHEN** the client sets status to PICKEDUP
- **THEN** the order stores frozen charged_total T, pickedup_at is set, and later menu size-option price changes do not change charged_total

#### Scenario: REQORDER004S03 Unpaid PICKEDUP allowed
- **GIVEN** an order with balance greater than zero
- **WHEN** the client sets status to PICKEDUP
- **THEN** the system accepts the update and payment_received remains false

### Requirement: REQORDER005 Order-level customization text
The system SHALL allow an optional free-text customization on the order header (for example whole-order spice level or nut-free). The field MUST be text only with no structured enum validation beyond length/emptiness rules defined by the API.

#### Scenario: REQORDER005S01 Set order customization
- **GIVEN** an active unfulfilled order
- **WHEN** the client sets customization_text to "nut-free"
- **THEN** subsequent get returns that customization_text

