package models

import "time"

// CateringMenu represents a menu for catering events for a specific tenant
type CateringMenu struct {
	ID            string      `db:"id"`
	TenantID      string      `db:"tenant_id"`
	Name          string      `db:"name"`
	Description    string      `db:"description"`
	EventDate     time.Time   `db:"event_date"`
	EventLocation string      `db:"event_location"`
	CreatedAt     time.Time   `db:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at"`
}

// CreateCateringMenuRequest represents the request payload for creating a catering menu
type CreateCateringMenuRequest struct {
	TenantID    string    `json:"tenant_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Description  string     `json:"description,omitempty"`
	EventDate   time.Time  `json:"event_date" binding:"required"`
	EventLocation string `json:"event_location" binding:"required"`
}

// UpdateCateringMenuRequest represents the request payload for updating a catering menu
type UpdateCateringMenuRequest struct {
	TenantID    string      `json:"tenant_id,omitempty"`
	Name        string      `json:"name,omitempty"`
	Description  string     `json:"description,omitempty"`
	EventDate   time.Time   `json:"event_date"`
	EventLocation string    `json:"event_location"`
}
