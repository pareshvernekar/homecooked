# order-payment Specification

## Purpose
TBD - created by archiving change order-intake. Update Purpose after archive.
## Requirements
### Requirement: REQPAY001 Record payment against an order
The system SHALL allow recording a payment on an existing active order for the current tenant. Each payment MUST include a positive amount and a mode of `cash`, `credit`, `paypal`, `zelle`, or `venmo`. An optional free-text reference MAY be stored. Payments MUST be allowed for any order status including `PICKEDUP`. Cross-tenant orders MUST NOT accept payments.

#### Scenario: REQPAY001S01 Record cash payment
- **GIVEN** an active order for the tenant
- **WHEN** the client records a payment with mode cash and amount 20
- **THEN** the system returns HTTP 201 and the payment appears on subsequent get of the order's payments

#### Scenario: REQPAY001S02 Reject invalid mode or non-positive amount
- **GIVEN** an active order
- **WHEN** the client records a payment with an unknown mode or amount <= 0
- **THEN** the system returns HTTP 400

#### Scenario: REQPAY001S03 Payment after PICKEDUP
- **GIVEN** an order with status PICKEDUP and outstanding balance
- **WHEN** the client records a payment covering the balance
- **THEN** the system accepts the payment and derived payment_received becomes true

### Requirement: REQPAY002 Derived payment aggregates
The system SHALL derive `paid_amount` as the sum of payment amounts for the order, `balance` as charged_total minus paid_amount, `payment_received` as true when balance <= 0, and `overpaid_amount` as max(0, paid_amount − charged_total). These fields MUST appear on order get responses and MUST NOT be independently writable by clients.

#### Scenario: REQPAY002S01 Partial payment
- **GIVEN** an order with charged_total 50 and one payment of 20
- **WHEN** the client gets the order
- **THEN** paid_amount is 20, balance is 30, payment_received is false, overpaid_amount is 0

#### Scenario: REQPAY002S02 Split modes
- **GIVEN** an order with charged_total 50
- **WHEN** the client records cash 20 and venmo 30
- **THEN** paid_amount is 50, balance is 0, and payment_received is true

### Requirement: REQPAY003 Overpay allowed without refund workflow
The system SHALL allow paid_amount to exceed charged_total. When that occurs, payment_received MUST be true, balance MUST be negative or zero as defined by charged_total − paid_amount, and overpaid_amount MUST equal the excess. The system MUST NOT require a refund entity in this capability. If charged_total later changes while unfulfilled (live price or override), aggregates MUST recompute against the new charged_total.

#### Scenario: REQPAY003S01 Cash overpay
- **GIVEN** an order with charged_total 47
- **WHEN** the client records a cash payment of 50
- **THEN** payment_received is true and overpaid_amount is 3

#### Scenario: REQPAY003S02 Override drop increases overage
- **GIVEN** an unfulfilled order with paid_amount 50 and charged_total 50
- **WHEN** the client sets total_override to 40
- **THEN** overpaid_amount is 10 and payment_received remains true

#### Scenario: REQPAY003S03 Price rise consumes overage
- **GIVEN** an unfulfilled order with paid_amount 50 and overpaid_amount 10 (charged_total 40)
- **WHEN** live or override charged_total becomes 55
- **THEN** balance is 5, payment_received is false, and overpaid_amount is 0

### Requirement: REQPAY004 List payments for an order
The system SHALL list payments for an order belonging to the current tenant in a stable order (for example by recorded time). Payments for other tenants' orders MUST NOT be listed.

#### Scenario: REQPAY004S01 List payments
- **GIVEN** an order with multiple payments
- **WHEN** the client lists payments for that order
- **THEN** all of that order's payments are returned with mode, amount, optional reference, and timestamps

