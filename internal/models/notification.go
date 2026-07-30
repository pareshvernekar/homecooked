package models

import "time"

// Notification represents a notification sent to users for a specific tenant
type Notification struct {
	ID           string      `db:"id"`
	TenantID     string      `db:"tenant_id"`
	UserID       string      `db:"user_id"`
	OrderId      string      `db:"order_id"`
	Type         string      `db:"type"` // "order_placed", "order_preparing", "order_ready", "order_delivered"
	Message      string      `db:"message"`
	IsRead       bool        `db:"is_read"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
}

// CreateNotificationRequest represents the request payload for creating a notification
type CreateNotificationRequest struct {
	TenantID  string     `json:"tenant_id" binding:"required"`
	UserID    string      `json:"user_id" binding:"required"`
	OrderId    string       `json:"order_id,omitempty"`
	Type       string      `json:"type" binding:"required,oneof=order_placed order_preparing order_ready order_delivered"`
	Message     string      `json:"message" binding:"required"`
	IsRead      bool      `json:"is_read,omitempty"`
}

// UpdateNotificationRequest represents the request payload for updating a notification
type UpdateNotificationRequest struct {
	TenantID    string  `json:"tenant_id,omitempty"`
	UserID      string   `json:"user_id,omitempty"`
	OrderId      string   `json:"order_id,omitempty"`
	Type         string   `json:"type" binding:"oneof=order_placed order_preparing order_ready order_delivered"`
	Message       string    `json:"message"`
	IsRead        bool     `json:"is_read"`
}
