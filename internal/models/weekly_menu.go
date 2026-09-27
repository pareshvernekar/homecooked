package models

import "time"

// WeeklyMenu represents a weekly menu for daily tiffin for a specific tenant
type WeeklyMenu struct {
	ID            string     `db:"id"`
	TenantID      string     `db:"tenant_id"`
	Name          string     `db:"name"`
	Description    string     `db:"description"`
	StartDate     time.Time  `db:"start_date"`
	EndDate       time.Time  `db:"end_date"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

// CreateWeeklyMenuRequest represents the request payload for creating a weekly menu
type CreateWeeklyMenuRequest struct {
	TenantID string `json:"tenant_id" binding:"required"`
	Name     string  `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	StartDate time.Time  `json:"start_date" binding:"required"`
	EndDate   time.Time  `json:"end_date" binding:"required"`
}

// UpdateWeeklyMenuRequest represents the request payload for updating a weekly menu
type UpdateWeeklyMenuRequest struct {
	TenantID  string     `json:"tenant_id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	StartDate time.Time  `json:"start_date"`
	EndDate   time.Time  `json:"end_date"`
}
