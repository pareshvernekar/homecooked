package models

// Menu status and type constants.
// REQMENU001, REQMENU006, REQMENU007
const (
	MenuStatusDraft     = "draft"
	MenuStatusPublished = "published"

	MenuTypeDaily    = "daily"
	MenuTypeWeekly   = "weekly"
	MenuTypeCatering = "catering"

	MenuItemKindSimple = "simple"
	MenuItemKindCombo  = "combo"
)

// Menu is a tenant-scoped sellable menu (daily/weekly/catering).
// REQMENU001
type Menu struct {
	ID            string  `db:"id" json:"id"`
	TenantID      string  `db:"tenant_id" json:"tenant_id"`
	Name          string  `db:"name" json:"name"`
	Description   string  `db:"description,omitempty" json:"description,omitempty"`
	MenuType      string  `db:"menu_type" json:"menu_type"`
	Status        string  `db:"status" json:"status"`
	IsActive      bool    `db:"is_active" json:"is_active"`
	MenuDate      *string `db:"menu_date" json:"menu_date,omitempty"`
	StartDate     *string `db:"start_date" json:"start_date,omitempty"`
	EndDate       *string `db:"end_date" json:"end_date,omitempty"`
	EventDate     *string `db:"event_date" json:"event_date,omitempty"`
	EventLocation *string `db:"event_location" json:"event_location,omitempty"`
	CreatedAt     int64   `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt     int64   `db:"updated_at" json:"updated_at,omitempty"`
}

// MenuCategory is a named section within a menu.
// REQMENU005
type MenuCategory struct {
	ID        string `db:"id" json:"id"`
	TenantID  string `db:"tenant_id" json:"tenant_id"`
	MenuID    string `db:"menu_id" json:"menu_id"`
	Name      string `db:"name" json:"name"`
	Sequence  int    `db:"sequence" json:"sequence"`
	IsActive  bool   `db:"is_active" json:"is_active"`
	CreatedAt int64  `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt int64  `db:"updated_at" json:"updated_at,omitempty"`
}

// MenuItem is a simple or combo offering under a category.
// REQITEM001 — no price / no size options on the item itself.
type MenuItem struct {
	ID          string  `db:"id" json:"id"`
	TenantID    string  `db:"tenant_id" json:"tenant_id"`
	MenuID      string  `db:"menu_id" json:"menu_id"`
	CategoryID  string  `db:"category_id" json:"category_id"`
	Kind        string  `db:"kind" json:"kind"`
	Name        string  `db:"name" json:"name"`
	Description string  `db:"description,omitempty" json:"description,omitempty"`
	Sequence    int     `db:"sequence" json:"sequence"`
	IsActive    bool    `db:"is_active" json:"is_active"`
	CreatedAt   int64   `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt   int64   `db:"updated_at" json:"updated_at,omitempty"`
	DefaultTotal *float64 `db:"-" json:"default_total,omitempty"` // REQITEM005 computed
	Components  []MenuItemComponent `db:"-" json:"components,omitempty"`
}

// MenuItemComponent links a food-item into a menu-item with a default size.
// REQITEM001, REQITEM003
type MenuItemComponent struct {
	ID                   string  `db:"id" json:"id"`
	TenantID             string  `db:"tenant_id" json:"tenant_id"`
	MenuItemID           string  `db:"menu_item_id" json:"menu_item_id"`
	FoodItemID           string  `db:"food_item_id" json:"food_item_id"`
	DefaultSizeOptionID  *string `db:"default_size_option_id" json:"default_size_option_id,omitempty"`
	IsActive             bool    `db:"is_active" json:"is_active"`
	CreatedAt            int64   `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt            int64   `db:"updated_at" json:"updated_at,omitempty"`
	FoodItem             *FoodItemDisplay `db:"-" json:"food_item,omitempty"` // live join
	SizeOptions          []MenuItemComponentSizeOption `db:"-" json:"size_options,omitempty"`
}

// FoodItemDisplay is the live-joined catalog slice for get-menu (no price).
// REQMENU003, REQFOOD001
type FoodItemDisplay struct {
	ID           string `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	IsVegetarian bool   `db:"is_vegetarian" json:"is_vegetarian"`
	Avoidance    string `db:"avoidance" json:"avoidance,omitempty"`
	ImageURL     string `db:"image_url" json:"image_url,omitempty"`
}

// MenuItemComponentSizeOption is how a component may be sized/priced.
// REQITEM003, REQITEM004
type MenuItemComponentSizeOption struct {
	ID         string  `db:"id" json:"id"`
	TenantID   string  `db:"tenant_id" json:"tenant_id"`
	ComponentID string `db:"component_id" json:"component_id"`
	SizeUnitID string  `db:"size_unit_id" json:"size_unit_id"`
	Qty        float64 `db:"qty" json:"qty"`
	Price      float64 `db:"price" json:"price"`
	IsActive   bool    `db:"is_active" json:"is_active"`
	IsDefault  bool    `db:"-" json:"is_default,omitempty"`
	CreatedAt  int64   `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt  int64   `db:"updated_at" json:"updated_at,omitempty"`
}

// MenuTree is get-menu payload with nested categories/items.
// REQMENU003
type MenuTree struct {
	Menu
	Categories []MenuCategoryTree `json:"categories"`
}

// MenuCategoryTree nests active menu-items under a category.
type MenuCategoryTree struct {
	MenuCategory
	Items []MenuItem `json:"items"`
}

// --- Request DTOs ---

// MenuCreateRequest creates a draft menu; type metadata optional.
// REQMENU001
type MenuCreateRequest struct {
	Name          string               `json:"name" binding:"required"`
	Description   string               `json:"description,omitempty"`
	MenuType      string               `json:"menu_type" binding:"required"`
	MenuDate      *string              `json:"menu_date,omitempty"`
	StartDate     *string              `json:"start_date,omitempty"`
	EndDate       *string              `json:"end_date,omitempty"`
	EventDate     *string              `json:"event_date,omitempty"`
	EventLocation *string              `json:"event_location,omitempty"`
	Categories    []MenuCategoryInput  `json:"categories,omitempty"`
}

// MenuUpdateRequest updates draft menu metadata / categories.
// REQMENU004
type MenuUpdateRequest struct {
	Name          *string             `json:"name,omitempty"`
	Description   *string             `json:"description,omitempty"`
	MenuDate      *string             `json:"menu_date,omitempty"`
	StartDate     *string             `json:"start_date,omitempty"`
	EndDate       *string             `json:"end_date,omitempty"`
	EventDate     *string             `json:"event_date,omitempty"`
	EventLocation *string             `json:"event_location,omitempty"`
	Categories    []MenuCategoryInput `json:"categories,omitempty"`
}

// MenuCategoryInput is used on create/update of a menu.
type MenuCategoryInput struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name" binding:"required"`
	Sequence int    `json:"sequence"`
}

// MenuItemCreateRequest adds a simple/combo item with components and size options.
// REQITEM001, REQITEM003
type MenuItemCreateRequest struct {
	CategoryID  string                     `json:"category_id" binding:"required"`
	Kind        string                     `json:"kind" binding:"required"`
	Name        string                     `json:"name" binding:"required"`
	Description string                     `json:"description,omitempty"`
	Sequence    int                        `json:"sequence"`
	Components  []MenuItemComponentInput   `json:"components" binding:"required"`
}

// MenuItemUpdateRequest updates a draft menu-item.
type MenuItemUpdateRequest struct {
	Name        *string                  `json:"name,omitempty"`
	Description *string                  `json:"description,omitempty"`
	Sequence    *int                     `json:"sequence,omitempty"`
	Components  []MenuItemComponentInput `json:"components,omitempty"`
}

// MenuItemComponentInput nests size options and designates a default.
type MenuItemComponentInput struct {
	FoodItemID          string                          `json:"food_item_id" binding:"required"`
	DefaultSizeOptionID *string                         `json:"default_size_option_id,omitempty"`
	DefaultSizeIndex    *int                            `json:"default_size_index,omitempty"` // index into SizeOptions on create
	SizeOptions         []MenuItemSizeOptionInput       `json:"size_options" binding:"required"`
}

// MenuItemSizeOptionInput is a size option payload.
type MenuItemSizeOptionInput struct {
	ID         string  `json:"id,omitempty"`
	SizeUnitID string  `json:"size_unit_id" binding:"required"`
	Qty        float64 `json:"qty" binding:"required"`
	Price      float64 `json:"price"`
}

// SizeOptionCreateRequest adds a size option to an existing component.
type SizeOptionCreateRequest struct {
	SizeUnitID string  `json:"size_unit_id" binding:"required"`
	Qty        float64 `json:"qty" binding:"required"`
	Price      float64 `json:"price"`
	IsDefault  bool    `json:"is_default,omitempty"`
}

// SizeOptionUpdateRequest updates a size option.
type SizeOptionUpdateRequest struct {
	SizeUnitID *string  `json:"size_unit_id,omitempty"`
	Qty        *float64 `json:"qty,omitempty"`
	Price      *float64 `json:"price,omitempty"`
	IsDefault  *bool    `json:"is_default,omitempty"`
}
