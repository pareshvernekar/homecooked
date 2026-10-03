## MODIFIED Requirements

### Requirement: REQPAY001 Record payment against an order
The system SHALL allow recording a payment on an existing active order for the current tenant when the order status is not `DECLINED`. Each payment MUST include a positive amount and a mode of `cash`, `credit`, `paypal`, `zelle`, or `venmo`. An optional free-text reference MAY be stored. Payments MUST be allowed for any non-declined order status including `PICKEDUP`. Recording a payment against a `DECLINED` order MUST be rejected. Cross-tenant orders MUST NOT accept payments.

#### Scenario: REQPAY001S01 Record cash payment
- **GIVEN** an active non-declined order for the tenant
- **WHEN** the client records a payment with mode cash and amount 20
- **THEN** the system returns HTTP 201 and the payment appears on subsequent get of the order's payments

#### Scenario: REQPAY001S02 Reject invalid mode or non-positive amount
- **GIVEN** an active non-declined order
- **WHEN** the client records a payment with an unknown mode or amount <= 0
- **THEN** the system returns HTTP 400

#### Scenario: REQPAY001S03 Payment after PICKEDUP
- **GIVEN** an order with status PICKEDUP and outstanding balance
- **WHEN** the client records a payment covering the balance
- **THEN** the system accepts the payment and derived payment_received becomes true

#### Scenario: REQPAY001S04 Reject payment on DECLINED order
- **GIVEN** an order with status DECLINED
- **WHEN** the client attempts to record a payment
- **THEN** the system returns HTTP 400 or 409 and no payment is stored
