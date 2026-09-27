package models

import "time"

// Order represents a user's order for a specific tenant
type Order struct {
	ID            string       `db:"id"`
	TenantID      string       `db:"tenant_id"`
	UserID        string       `db:"user_id"`
	MenuType      string        `db:"menu_type"` // "weekly" or "catering"
	MenuID        string        `db:"menu_id"`
	Status        string        `db:"status"` // "pending", "preparing", "ready", "delivered"
	OrderDate     *time.Time    `db:"order_date"`
	DeliveryDate  *time.Time    `db:"delivery_date"`
	TotalPrice    float64       `db:"total_price"`
	CreatedAt     time.Time     `db:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at"`
}

// CreateOrderRequest represents the request payload for creating an order
type CreateOrderRequest struct {
	TenantID  string      `json:"tenant_id" binding:"required"`
	UserID    string       `json:"user_id" binding:"required"`
	MenuType   string        `json:"menu_type" binding:"required,oneof=weekly catering"`
	MenuID     string       `json:"menu_id" binding:"required"`
	Status     string        `json:"status" binding:"required,oneof=pending preparing ready delivered"`
	OrderDate    *time.Time   `json:"order_date" binding:"required"`
	DeliveryDate *time.Time   `json:"delivery_date,omitempty"`
	Items       []CreateOrderItemRequest `json:"items" binding:"required,dive"`
}

// CreateOrderItemRequest represents the request payload for an order item
type CreateOrderItemRequest struct {
	MenuItemId string    `json:"menu_item_id" binding:"required"`
	Quantity    int       `json:"quantity" binding:"required,gte=1"`
	Price       float64   `json:"price"`
	Description *string    `json:"description,omitempty"`
}

// UpdateOrderRequest represents the request payload for updating an order
type UpdateOrderRequest struct {
	TenantID    string      `json:"tenant_id,omitempty"`
	UserID      string       `json:"user_id,omitempty"`
	MenuType     string        `json:"menu_type"`
	MenuID       string        `json:"menu_id"`
	Status       string        `json:"status" binding:"oneof=pending preparing ready delivered"`
	OrderDate    *time.Time   `json:"order_date"`
	DeliveryDate *time.Time   `json:"delivery_date,omitempty"`
	TotalPrice    float64       `json:"total_price"`
}
