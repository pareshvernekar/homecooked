## ADDED Requirements

### Requirement: REQSIZE001 Standard size units available
The system SHALL provide a seeded set of standard size units available to all tenants. Standard units MUST include at least: `serving`, `tray`, `piece`, `dozen`, `kg`, `g`, and `liter`. Standard units MUST be readable by tenants and MUST NOT be deletable or renamable by tenants.

#### Scenario: REQSIZE001S01 List includes standard units
- **GIVEN** a valid tenant context
- **WHEN** the client lists size units
- **THEN** the response includes the standard units with stable codes

#### Scenario: REQSIZE001S02 Tenant cannot delete standard unit
- **GIVEN** a standard system size unit
- **WHEN** the client attempts to delete or deactivate that unit as the tenant
- **THEN** the system returns HTTP 403 or 400 indicating system units cannot be modified

### Requirement: REQSIZE002 Tenant-defined size units
The system SHALL allow a tenant to create custom size units scoped to that tenant. Custom units MUST have a non-empty code and display name unique within the tenant (among that tenant's customs). Custom units MUST appear in the tenant's size-unit list alongside standards.

#### Scenario: REQSIZE002S01 Create custom size unit
- **GIVEN** a valid tenant context
- **WHEN** the client creates a custom size unit with code and display name
- **THEN** the system returns HTTP 201 and subsequent list includes the custom unit for that tenant only

#### Scenario: REQSIZE002S02 Custom unit not visible to other tenants
- **GIVEN** tenant A created a custom size unit
- **WHEN** tenant B lists size units
- **THEN** tenant A's custom unit does not appear

#### Scenario: REQSIZE002S03 Reject duplicate custom code for tenant
- **GIVEN** tenant A already has a custom unit with code `party-pan`
- **WHEN** tenant A creates another custom unit with code `party-pan`
- **THEN** the system returns a validation or conflict error

### Requirement: REQSIZE003 Size options reference valid units
The system SHALL allow a size option to reference either a standard size unit or a custom size unit owned by the current tenant. References to another tenant's custom units or unknown units MUST be rejected.

#### Scenario: REQSIZE003S01 Size option with standard unit
- **GIVEN** an active draft menu-item component and a standard unit `tray`
- **WHEN** the client creates a size option using `tray` with valid qty and price
- **THEN** the system persists the size option under that component

#### Scenario: REQSIZE003S02 Size option with tenant custom unit
- **GIVEN** an active draft menu-item component and a custom unit owned by the tenant
- **WHEN** the client creates a size option using that custom unit
- **THEN** the system persists the size option under that component

#### Scenario: REQSIZE003S03 Reject other tenant custom unit
- **GIVEN** a custom size unit owned by tenant B and an active draft menu-item component for tenant A
- **WHEN** tenant A creates a size option referencing that unit
- **THEN** the system returns a validation or not-found error
