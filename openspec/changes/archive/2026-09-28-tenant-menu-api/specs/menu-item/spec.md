## ADDED Requirements

### Requirement: REQITEM001 Add simple or combo menu-item
The system SHALL allow adding a menu-item to a category within an active **draft** menu. A menu-item MUST be either `simple` (exactly one food-item component) or `combo` (two or more food-item components). Each component MUST reference an active food-item owned by the same tenant. Simple and combo menu-items MUST use the same component shape. A component MUST NOT have a price field of its own; sellable prices live on the component's size options.

#### Scenario: REQITEM001S01 Add simple menu-item
- **GIVEN** an active draft menu with a category and an active food-item for the tenant
- **WHEN** the client adds a simple menu-item with one component that has at least one size option and a default size option
- **THEN** the system returns success and the item appears under that category with size options nested under the component

#### Scenario: REQITEM001S02 Add combo menu-item
- **GIVEN** an active draft menu with a category and two or more active food-items for the tenant
- **WHEN** the client adds a combo menu-item with multiple components, each having at least one size option with a sellable price and a default size option
- **THEN** the system returns success and each component includes its size options with prices

#### Scenario: REQITEM001S03 Reject invalid component cardinality
- **GIVEN** an active draft menu and category
- **WHEN** the client adds a simple item with zero or more than one component, or a combo with fewer than two components
- **THEN** the system returns HTTP 400 with a validation error

#### Scenario: REQITEM001S04 Reject food-item from other tenant or inactive
- **GIVEN** a food-item that is inactive or belongs to another tenant
- **WHEN** the client includes it as a component on a draft menu
- **THEN** the system returns a validation or not-found error

### Requirement: REQITEM002 Remove menu-item from menu
The system SHALL allow removing a menu-item from an active **draft** menu. Removal MUST make the item unavailable on subsequent get/list of that menu (deactivate or equivalent soft remove).

#### Scenario: REQITEM002S01 Remove menu-item
- **GIVEN** an active draft menu containing a menu-item
- **WHEN** the client removes that menu-item
- **THEN** subsequent get menu does not include the removed item

#### Scenario: REQITEM002S02 Remove unknown item
- **GIVEN** an active draft menu
- **WHEN** the client removes a menu-item ID that does not exist on that menu
- **THEN** the system returns HTTP 404

### Requirement: REQITEM003 Size options nested under components
The system SHALL represent how each component is sized and priced via nested size options under that component (not under the menu-item). Each component MUST have one or more size options and MUST designate exactly one of them as the default size option. Each size option MUST include a size unit (standard or tenant custom), a positive quantity, and a non-negative price. Multiple size options per component MUST be allowed on daily, weekly, and catering menus. Size-option and default-size mutations MUST require an active draft menu.

#### Scenario: REQITEM003S01 Create component with multiple size options and default
- **GIVEN** an active draft daily (or weekly/catering) menu and valid size units
- **WHEN** the client adds a menu-item whose component(s) each include two or more size options and a default referencing one of those options
- **THEN** the system persists all size options nested under each component and records the default

#### Scenario: REQITEM003S02 Reject component without size options or default
- **GIVEN** an active draft menu and category
- **WHEN** the client adds a menu-item with a component that has no size options, or size options but no valid default
- **THEN** the system returns HTTP 400 with a validation error

#### Scenario: REQITEM003S03 Add or update size option on existing component
- **GIVEN** an active draft menu-item component
- **WHEN** the client adds or updates a size option with valid unit, qty, and price
- **THEN** the system persists the change and returns success

#### Scenario: REQITEM003S04 Reject default that is not one of the component's size options
- **GIVEN** an active draft menu-item component with size options
- **WHEN** the client sets the default size option to an ID that does not belong to that component
- **THEN** the system returns HTTP 400 with a validation error

### Requirement: REQITEM004 Price on component size option only
The system SHALL store the sellable price on the component size option only. The menu-item MUST NOT have a price column or menu-item-level size options. Food-item catalog records MUST NOT carry a catalog price. A component MUST NOT carry a price field of its own. Price MUST be greater than or equal to zero. Size option quantity MUST be greater than zero.

#### Scenario: REQITEM004S01 Reject negative price or non-positive qty
- **GIVEN** an active draft menu-item component
- **WHEN** the client creates or updates a size option with price < 0 or qty <= 0
- **THEN** the system returns HTTP 400 with a validation error

#### Scenario: REQITEM004S02 Get menu shows prices on component size options
- **GIVEN** a menu-item with components that have size options
- **WHEN** the client gets the menu
- **THEN** each component's size options include their prices, the component has no price field, and the food-item embed has no catalog price

### Requirement: REQITEM005 Combo price is sum of component defaults
For combo menu-items, the system SHALL treat the default sellable price as the sum of each component's default size-option price. That total MUST be exposed on get as a computed `default_total` field for convenience and MUST NOT be persisted as a separate combo package price. There MUST be no package-level discount or override total. Future order-time selection (out of scope) MUST choose only from each component's allowed size options; order price is the sum of selected size-option prices.

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
The system SHALL reject add, update, or remove operations on menu-items, components, and size options when the parent menu is inactive, not found for the tenant, or has `status=published`. Published menus MUST be unpublished to draft before such mutations are allowed.

#### Scenario: REQITEM006S01 Reject ops on inactive menu
- **GIVEN** a deactivated menu
- **WHEN** the client attempts to add, update, or remove a menu-item, component, or size option on that menu
- **THEN** the system returns HTTP 404 or a conflict indicating the menu is not available

#### Scenario: REQITEM006S02 Reject ops on published menu
- **GIVEN** an active published menu
- **WHEN** the client attempts to add, update, or remove a menu-item, component, or size option on that menu
- **THEN** the system returns HTTP 409 or 400 indicating the menu is not editable while published
