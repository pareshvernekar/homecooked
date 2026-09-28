## ADDED Requirements

### Requirement: REQMENU001 Create tenant menu
The system SHALL allow an authenticated tenant context to create a menu with a non-empty name and a valid `menu_type` (`daily`, `weekly`, or `catering`). Type-specific metadata MAY be omitted on create. The created menu MUST have `status=draft`, `is_active=true`, and be scoped exclusively to that tenant.

#### Scenario: REQMENU001S01 Successful menu create as draft
- **GIVEN** a valid tenant ID in request context
- **WHEN** the client creates a menu with name and valid type (type metadata optional)
- **THEN** the system returns HTTP 201 with the menu including `id`, `tenant_id`, `menu_type`, `name`, `status=draft`, `is_active=true`, and timestamps

#### Scenario: REQMENU001S02 Reject create without name or invalid type
- **GIVEN** a valid tenant ID in request context
- **WHEN** the client creates a menu with an empty name or a `menu_type` outside `daily|weekly|catering`
- **THEN** the system returns HTTP 400 with a validation error

### Requirement: REQMENU002 List tenant menus
The system SHALL list active menus for the current tenant only. Menus belonging to other tenants MUST NOT appear. By default the list MUST return only `status=published` menus. Clients MUST be able to request drafts via a status filter (`draft` or `all`).

#### Scenario: REQMENU002S01 List returns only current tenant active published menus by default
- **GIVEN** active published menus exist for tenant A and tenant B, and tenant A also has an active draft
- **WHEN** tenant A lists menus without a status filter
- **THEN** the response includes only tenant A's active published menus (not drafts, not tenant B)

#### Scenario: REQMENU002S02 Inactive menus excluded from default list
- **GIVEN** tenant A has an active published menu and a deactivated menu
- **WHEN** tenant A lists menus
- **THEN** only the active published menu is returned

#### Scenario: REQMENU002S03 List drafts via status filter
- **GIVEN** tenant A has an active draft and an active published menu
- **WHEN** tenant A lists menus with `status=draft`
- **THEN** only the active draft is returned

### Requirement: REQMENU003 Get menu by ID
The system SHALL return a single active menu for the current tenant (draft or published), including its categories, menu-items (each with a computed `default_total`), components (with nested size options and default size indicators), and live-joined food-item display fields. Each component MUST include a live-joined display representation of the referenced food-item (`id`, `name`, and dietary/display fields such as `is_vegetarian`, `avoidance`, and `image_url` when present) and MUST NOT include a catalog price. Inactive or cross-tenant menus MUST NOT be returned as found.

#### Scenario: REQMENU003S01 Get active menu with structure and live-joined food items
- **GIVEN** an active menu for the tenant with categories, items, components (each with size options and a default), and food-items
- **WHEN** the client gets the menu by ID
- **THEN** the system returns HTTP 200 with nested categories, menu-items (each with computed `default_total`), components (each with live-joined food-item display fields, no catalog price, nested size options with prices, and default indicated)

#### Scenario: REQMENU003S02 Not found for other tenant or inactive
- **GIVEN** a menu that is inactive or belongs to another tenant
- **WHEN** the client gets that menu by ID under the current tenant
- **THEN** the system returns HTTP 404

#### Scenario: REQMENU003S03 Live join reflects catalog updates
- **GIVEN** a menu component referencing a food-item whose name later changes in the catalog
- **WHEN** the client gets the menu by ID
- **THEN** the component's embedded food-item name matches the current catalog name

### Requirement: REQMENU004 Update and deactivate menu
The system SHALL allow updating menu metadata for an active **draft** menu owned by the current tenant. Updates to published or inactive menus MUST be rejected. DELETE MUST deactivate the menu (`is_active=false`) rather than hard-deleting it, whether the menu is draft or published.

#### Scenario: REQMENU004S01 Update draft menu metadata
- **GIVEN** an active draft menu for the tenant
- **WHEN** the client updates allowed metadata (name, description, type-specific fields)
- **THEN** the system returns success and persists the new metadata

#### Scenario: REQMENU004S02 Reject update of published menu
- **GIVEN** an active published menu for the tenant
- **WHEN** the client attempts to update menu metadata
- **THEN** the system returns HTTP 409 or 400 indicating the menu is not editable while published

#### Scenario: REQMENU004S03 Delete deactivates menu
- **GIVEN** an active draft or published menu for the tenant
- **WHEN** the client deletes the menu
- **THEN** the menu is marked inactive and subsequent default list/get treat it as unavailable

#### Scenario: REQMENU004S04 Cannot update other tenant menu
- **GIVEN** a menu owned by tenant B
- **WHEN** tenant A attempts to update or delete that menu
- **THEN** the system returns HTTP 404

### Requirement: REQMENU005 Menu divided into categories
The system SHALL organize each menu into categories. Each category MUST belong to exactly one menu and MAY contain zero or more menu-items. Category order MUST be represented by a sequence value. Category mutations on a published menu MUST be rejected.

#### Scenario: REQMENU005S01 Create menu with categories
- **GIVEN** a valid tenant context
- **WHEN** the client creates or updates a **draft** menu including named categories with sequences
- **THEN** the menu response includes those categories in sequence order

#### Scenario: REQMENU005S02 Empty category allowed on draft
- **GIVEN** an active draft menu for the tenant
- **WHEN** a category is added with no menu-items
- **THEN** the category is persisted and returned on get menu

### Requirement: REQMENU006 Menu type daily weekly or catering
The system SHALL require each menu to have exactly one type: `daily`, `weekly`, or `catering`. Type-specific metadata MUST be validated as required when publishing (`daily`: `menu_date`; `weekly`: `start_date` and `end_date`; `catering`: `event_date` and `event_location`). Type-specific metadata MAY be absent or incomplete while the menu is a draft.

#### Scenario: REQMENU006S01 Weekly menu can be created without dates as draft
- **GIVEN** a valid tenant context
- **WHEN** the client creates a weekly menu without start_date and end_date
- **THEN** the system persists the menu with `menu_type=weekly`, `status=draft`

#### Scenario: REQMENU006S02 Type metadata accepted on draft update
- **GIVEN** an active draft catering menu
- **WHEN** the client updates event_date and event_location
- **THEN** the system persists those fields while status remains draft

#### Scenario: REQMENU006S03 Publish rejects missing type-required metadata
- **GIVEN** an otherwise structurally complete draft weekly or catering menu missing type-required date/location fields
- **WHEN** the client publishes the menu
- **THEN** the system returns HTTP 400 with a validation error and status remains draft

### Requirement: REQMENU007 Publish menu
The system SHALL allow publishing an active draft menu when it is structurally complete. On success the menu MUST have `status=published`. Publish MUST NOT change `is_active`. Publishing an already-published, inactive, or cross-tenant menu MUST fail appropriately.

A menu is structurally complete when all of the following hold:

1. Type-required metadata is present for its `menu_type`.
2. At least one category exists.
3. At least one menu-item exists across categories.
4. Each menu-item has valid simple/combo component cardinality; every component references a same-tenant active food-item; every component has at least one size option and a valid default size option among them.

#### Scenario: REQMENU007S01 Publish structurally complete draft
- **GIVEN** an active draft menu with type-required metadata, at least one category, at least one complete menu-item (valid components each with size options and a default)
- **WHEN** the client publishes the menu
- **THEN** the system returns success and the menu has `status=published`

#### Scenario: REQMENU007S02 Reject publish when incomplete
- **GIVEN** an active draft menu missing a category, menu-item, component size option/default, valid component cardinality, or type-required metadata
- **WHEN** the client publishes the menu
- **THEN** the system returns HTTP 400 and status remains draft

#### Scenario: REQMENU007S03 Reject publish of inactive or other-tenant menu
- **GIVEN** a menu that is inactive or belongs to another tenant
- **WHEN** the client attempts to publish that menu under the current tenant
- **THEN** the system returns HTTP 404

### Requirement: REQMENU008 Unpublish menu
The system SHALL allow unpublishing an active published menu, setting `status=draft` while retaining structure and `is_active=true`. After unpublish, metadata and item mutations MUST be allowed again. Unpublish of a draft, inactive, or cross-tenant menu MUST fail appropriately.

#### Scenario: REQMENU008S01 Unpublish moves to draft
- **GIVEN** an active published menu for the tenant
- **WHEN** the client unpublishes the menu
- **THEN** the system returns success, `status=draft`, structure is retained, and subsequent updates are allowed

#### Scenario: REQMENU008S02 Reject unpublish of draft or inactive
- **GIVEN** a draft or inactive menu
- **WHEN** the client attempts to unpublish
- **THEN** the system returns HTTP 400/409 (draft) or HTTP 404 (inactive)

### Requirement: REQMENU009 Unique published weekly start date
The system SHALL enforce that, for a given tenant, at most one active published weekly menu may have a given `start_date`. The constraint MUST be evaluated when publishing a weekly menu. Multiple draft weekly menus MAY share the same `start_date`. Daily and catering menus MUST NOT be subject to this uniqueness rule. Inactive menus MUST NOT hold the `start_date` slot. Overlapping weekly date ranges with different `start_date` values are allowed.

#### Scenario: REQMENU009S01 Reject second published weekly with same start_date
- **GIVEN** tenant A has an active published weekly menu with `start_date=D` and an active draft weekly menu also with `start_date=D` that is otherwise structurally complete
- **WHEN** the client publishes the draft
- **THEN** the system returns HTTP 409 or 400 indicating the start date is already in use and the draft remains draft

#### Scenario: REQMENU009S02 Drafts may share start_date
- **GIVEN** tenant A has no published weekly menu for `start_date=D`
- **WHEN** the client creates or updates two draft weekly menus both with `start_date=D`
- **THEN** both drafts are persisted successfully

#### Scenario: REQMENU009S03 Slot freed by unpublish or deactivate
- **GIVEN** tenant A had an active published weekly menu with `start_date=D` that is then unpublished or deactivated, and another structurally complete draft weekly with `start_date=D`
- **WHEN** the client publishes the draft
- **THEN** the system returns success and the draft becomes published
