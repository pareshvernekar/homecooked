package models

import "time"

// MenuItem represents an item in a menu (weekly or catering) with size and price configuration for a specific tenant
type MenuItem struct {
	ID             string     `db:"id"`
	TenantID       string     `db:"tenant_id"`
	MenuID         string     `db:"menu_id"`
	MenuType       string     `db:"menu_type"` // "weekly" or "catering"
	FoodItemID     string     `db:"food_item_id"`
	Description    string     `db:"description"`
	Size           string     `db:"size"` // e.g., "single", "double"
	Price          float64    `db:"price"`
	Sequence       int        `db:"sequence"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

// CreateMenuItemRequest represents the request payload for creating a menu item
type CreateMenuItemRequest struct {
	TenantID   string     `json:"tenant_id" binding:"required"`
	MenuID     string     `json:"menu_id" binding:"required"`
	FoodItemID string     `json:"food_item_id" binding:"required"`
	Size       string     `json:"size" binding:"required"`
	Price      float64    `json:"price" binding:"required,gte=0"`
	Sequence   int        `json:"sequence" binding:"required"`
	Description string     `json:"description,omitempty"`
	MenuType   string     `json:"menu_type" binding:"required,oneof=weekly catering"`
}

// UpdateMenuItemRequest represents the request payload for updating a menu item
type UpdateMenuItemRequest struct {
	TenantID   string     `json:"tenant_id,omitempty"`
	MenuID     string     `json:"menu_id,omitempty"`
	FoodItemID string     `json:"food_item_id"`
	Size       string     `json:"size,omitempty"`
	Price      float64    `json:"price"`
	Sequence   int        `json:"sequence"`
	Description string     `json:"description,omitempty"`
}
