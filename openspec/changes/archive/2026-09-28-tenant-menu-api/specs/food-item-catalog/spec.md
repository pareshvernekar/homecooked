## ADDED Requirements

### Requirement: REQFOOD001 Food item has no price
The system SHALL treat food items as catalog definitions without price. Create and update food-item requests MUST NOT require or persist a price field. Food-item responses MUST NOT include a catalog price field. Sellable prices MUST live on menu-item size options instead.

#### Scenario: REQFOOD001S01 Create food item without price
- **GIVEN** a valid tenant context and category
- **WHEN** the client creates a food item without a price field
- **THEN** the system returns HTTP 201 and the response body does not include a price field

#### Scenario: REQFOOD001S02 Reject or ignore price on create/update
- **GIVEN** a valid tenant context
- **WHEN** the client sends a food-item create or update payload that includes a price field
- **THEN** the system does not persist price on the food-item (either by rejecting the unknown/disallowed field per API validation rules, or by ignoring it such that no food-item price is stored)

#### Scenario: REQFOOD001S03 List and get omit price
- **GIVEN** existing food items for the tenant
- **WHEN** the client lists or gets a food item
- **THEN** each food-item representation omits catalog price
