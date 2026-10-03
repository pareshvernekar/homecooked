## ADDED Requirements

### Requirement: REQNOTIF001 Tenant cook admin phone setting
The system SHALL allow configuring a cook admin phone number per tenant. The value MUST be stored as tenant settings and MUST be readable and updatable by the current tenant API. An empty or unset cook admin phone MUST be allowed (notifications to cook are skipped until set).

#### Scenario: REQNOTIF001S01 Set cook admin phone
- **GIVEN** a valid tenant context
- **WHEN** the client sets cook_admin_phone to a non-empty phone string
- **THEN** subsequent get of tenant settings returns that cook_admin_phone

#### Scenario: REQNOTIF001S02 Clear cook admin phone
- **GIVEN** a tenant with cook_admin_phone set
- **WHEN** the client clears cook_admin_phone
- **THEN** subsequent get shows it empty or null

### Requirement: REQNOTIF002 Enqueue on order lifecycle events
The system SHALL enqueue an outbound SMS notification when each of the following occurs for an active order: order created (`order.created`), accepted (`order.accepted`), declined (`order.declined`), marked ready (`order.ready`), and picked up (`order.picked_up`). The system MUST NOT enqueue a notification solely because preparing started (`IN_PROGRESS`). Enqueue MUST occur in the same database transaction as the order write that caused the event (or fail the order write if enqueue insert fails).

#### Scenario: REQNOTIF002S01 Create enqueues cook notification when phone configured
- **GIVEN** the tenant has cook_admin_phone set
- **WHEN** a new order is created
- **THEN** a pending outbox row exists with event_type order.created and recipient equal to cook_admin_phone

#### Scenario: REQNOTIF002S02 Create skips cook notification when phone unset
- **GIVEN** the tenant has no cook_admin_phone
- **WHEN** a new order is created
- **THEN** no order.created outbox row is created and the order create still succeeds

#### Scenario: REQNOTIF002S03 Accept enqueues customer notification
- **GIVEN** an order that is accepted
- **WHEN** the accept action completes
- **THEN** a pending outbox row exists with event_type order.accepted and recipient equal to the order customer_phone

#### Scenario: REQNOTIF002S04 Decline enqueues customer notification including reason
- **GIVEN** an order that is refused with a refuse_reason
- **WHEN** the refuse action completes
- **THEN** a pending outbox row exists with event_type order.declined whose body or payload includes that refuse_reason

#### Scenario: REQNOTIF002S05 Ready and pickup enqueue customer notifications
- **GIVEN** an order that is marked ready, then picked up
- **WHEN** each action completes
- **THEN** pending or delivered outbox rows exist for order.ready and order.picked_up to the customer_phone

#### Scenario: REQNOTIF002S06 Preparing does not enqueue
- **GIVEN** an accepted order
- **WHEN** the client starts preparing
- **THEN** no new outbox row is created for a preparing event

### Requirement: REQNOTIF003 Async delivery with retry
The system SHALL deliver pending outbox notifications asynchronously through an SMS provider interface. Delivery MUST be attempted by a worker that claims pending rows, invokes the provider, and marks the row delivered on success. On failure the system MUST increment attempts and schedule a retry until a configured maximum, after which the row MUST be marked dead. Order HTTP handlers MUST NOT wait for provider send success to return success for the order action.

#### Scenario: REQNOTIF003S01 Local provider delivers pending row
- **GIVEN** a pending outbox row and the local SMS provider
- **WHEN** the worker runs a delivery cycle
- **THEN** the row status becomes delivered and the local provider records the message

#### Scenario: REQNOTIF003S02 Failed send retries then dead
- **GIVEN** a pending outbox row and a provider that always fails
- **WHEN** the worker exhausts the maximum attempts
- **THEN** the row status is dead and last_error is stored

### Requirement: REQNOTIF004 Pluggable SMS provider
The system SHALL send SMS through an injectable provider abstraction. A local provider MUST be available for development and automated tests without an external SMS vendor. The concrete provider MUST be selectable by configuration.

#### Scenario: REQNOTIF004S01 Local provider selected by config
- **GIVEN** configuration selecting the local SMS provider
- **WHEN** the worker delivers a message
- **THEN** the local provider handles the send and no external SMS API is required

### Requirement: REQNOTIF005 Notification audit per order
The system SHALL allow listing notification outbox records for an order belonging to the current tenant (event type, recipient, channel, status, timestamps, and refuse/reason content where applicable). Cross-tenant orders MUST NOT expose notification records.

#### Scenario: REQNOTIF005S01 List notifications for order
- **GIVEN** an order with one or more outbox rows for the tenant
- **WHEN** the client lists notifications for that order id
- **THEN** those rows are returned and no other tenant's rows appear
