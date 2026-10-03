## ADDED Requirements

### Requirement: REQOLINE001 Add line with per-component size selection
The system SHALL allow adding a line to an active unfulfilled order. Each line MUST reference a menu-item that belongs to the order's bound published menu. The client MAY supply a size_option_id per component; omitted components MUST use that component's default size option. Every supplied size_option_id MUST be one of the component's size options on the published menu. Simple and combo menu-items MUST use the same selection shape. Quantity MUST be a positive integer.

#### Scenario: REQOLINE001S01 Add line with defaults
- **GIVEN** an active unfulfilled order on a published menu that has a simple or combo menu-item with defaults
- **WHEN** the client adds a line with menu_item_id and quantity and omits size selections
- **THEN** the system stores selections equal to each component's default size option and returns success

#### Scenario: REQOLINE001S02 Add line with explicit selections
- **GIVEN** a combo menu-item with multiple allowed size options per component
- **WHEN** the client adds a line selecting an allowed size_option_id for each component
- **THEN** the system persists those selections and line unit price equals the sum of the selected options' current prices (absent line override)

#### Scenario: REQOLINE001S03 Reject invalid size option
- **GIVEN** an active unfulfilled order
- **WHEN** the client selects a size_option_id that is not on that component for the bound menu
- **THEN** the system returns HTTP 400 and does not add the line

#### Scenario: REQOLINE001S04 Reject line on picked up order
- **GIVEN** an order with status PICKEDUP
- **WHEN** the client attempts to add a line
- **THEN** the system returns HTTP 400 or 409

### Requirement: REQOLINE002 Line customization and unit price override
The system SHALL allow optional free-text `customization_text` on each line and an optional non-negative `unit_price_override`. When unit_price_override is set, that value MUST be used as the line unit price instead of the live sum of selected size-option prices. Overrides and customization updates MUST be rejected after PICKEDUP.

#### Scenario: REQOLINE002S01 Line spicy note and override
- **GIVEN** an unfulfilled order line
- **WHEN** the client sets customization_text to "spicy" and unit_price_override to 10
- **THEN** get order shows those fields and the line contributes 10 × quantity toward subtotal (before any order total_override)

#### Scenario: REQOLINE002S02 Clear line override restores live price
- **GIVEN** a line with unit_price_override set
- **WHEN** the client clears unit_price_override
- **THEN** the line unit price again equals the sum of current selected size-option prices

### Requirement: REQOLINE003 Live line pricing and contribution to charged total
While the order is unfulfilled, the system SHALL compute each line's unit price as unit_price_override if set, otherwise the sum of the current prices of the selected size options on the published menu. Line extended amount MUST be unit price × quantity. Order subtotal MUST be the sum of line extended amounts. If the order has total_override set, charged_total MUST equal total_override; otherwise charged_total MUST equal subtotal. After PICKEDUP, charged_total and line amounts MUST use the frozen values.

#### Scenario: REQOLINE003S01 Menu price change moves unfulfilled total
- **GIVEN** an unfulfilled order line without overrides whose selections reference size options priced at a sum of P
- **WHEN** those size-option prices are updated so the sum becomes P2
- **THEN** subsequent get order shows line unit P2 and charged_total reflecting P2 (absent order override)

#### Scenario: REQOLINE003S02 Order override wins over line math
- **GIVEN** an unfulfilled order whose line subtotal is S and total_override is O
- **WHEN** the client gets the order
- **THEN** charged_total equals O

#### Scenario: REQOLINE003S03 Frozen amounts ignore later menu price changes
- **GIVEN** an order frozen at PICKEDUP with charged_total T
- **WHEN** menu size-option prices change
- **THEN** charged_total remains T

### Requirement: REQOLINE004 Update and remove lines
The system SHALL allow updating selections, quantity, customization_text, and unit_price_override, and removing lines, only on active unfulfilled orders. Removal MUST exclude the line from subsequent get/list of that order.

#### Scenario: REQOLINE004S01 Update selection
- **GIVEN** an unfulfilled order line
- **WHEN** the client changes a component's size_option_id to another allowed option
- **THEN** subsequent get reflects the new selection and live price

#### Scenario: REQOLINE004S02 Remove line
- **GIVEN** an unfulfilled order with a line
- **WHEN** the client removes that line
- **THEN** subsequent get does not include the line
