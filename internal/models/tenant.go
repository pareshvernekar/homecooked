package models

import "time"

// Tenant represents a client organization that uses the system
type Tenant struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description  string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// CreateTenantRequest represents the request payload for creating a tenant
type CreateTenantRequest struct {
	Name        string `json:"name" binding:"required"`
	Description  string    `json:"description,omitempty"`
}

// UpdateTenantRequest represents the request payload for updating a tenant
type UpdateTenantRequest struct {
	Name        string    `json:"name,omitempty"`
	Description  string    `json:"description,omitempty"`
}
