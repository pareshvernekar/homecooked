## MODIFIED Requirements

### Requirement: REQITEM005 Combo price is sum of component defaults
For combo menu-items, the system SHALL treat the default sellable price as the sum of each component's default size-option price. That total MUST be exposed on get as a computed `default_total` field for convenience and MUST NOT be persisted as a separate combo package price. There MUST be no package-level discount or override total on the menu-item. Order-time selection (see order-line capability) MUST choose only from each component's allowed size options; the unfulfilled order line unit price is the sum of selected size-option prices unless a line or order override applies.

#### Scenario: REQITEM005S01 Combo default total equals sum of component default prices
- **GIVEN** a combo with components A and B whose default size options are priced at P_A and P_B
- **WHEN** the client gets the menu-item
- **THEN** each component includes its size options with prices, defaults are indicated, and the response includes `default_total` equal to P_A + P_B

#### Scenario: REQITEM005S02 No package-level price on combo
- **GIVEN** a combo menu-item
- **WHEN** the client gets the menu-item
- **THEN** the response does not include a menu-item-level size option or persisted package price field distinct from component size options and the computed `default_total`

#### Scenario: REQITEM005S03 Simple item default total equals its default size price
- **GIVEN** a simple menu-item whose component default size option is priced at P
- **WHEN** the client gets the menu-item
- **THEN** the response includes `default_total` equal to P

### Requirement: REQITEM006 Mutations require active draft menu
The system SHALL reject add, update, or remove operations on menu-items, components, and size options when the parent menu is inactive, not found for the tenant, or has `status=published`, **except** that updating the **price** of an existing size option on an active published menu MUST be allowed so unfulfilled orders can reflect live price changes. Structural changes (add/remove options, change qty/unit, change defaults, add/remove items or components) on published menus MUST still be rejected; the menu MUST be unpublished to draft for those mutations.

#### Scenario: REQITEM006S01 Reject ops on inactive menu
- **GIVEN** a deactivated menu
- **WHEN** the client attempts to add, update, or remove a menu-item, component, or size option on that menu
- **THEN** the system returns HTTP 404 or a conflict indicating the menu is not available

#### Scenario: REQITEM006S02 Reject structural ops on published menu
- **GIVEN** an active published menu
- **WHEN** the client attempts to add or remove a menu-item, component, or size option, or change size-option qty/unit/default
- **THEN** the system returns HTTP 409 or 400 indicating the menu is not editable while published

#### Scenario: REQITEM006S03 Allow price-only update on published size option
- **GIVEN** an active published menu with a size option priced at P
- **WHEN** the client updates only that size option's price to P2
- **THEN** the system accepts the update and subsequent menu get shows price P2
