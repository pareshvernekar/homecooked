package models

// Tenant represents a client organization that uses the system
type Tenant struct {
	ID          string `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description,omitempty"`
	IsActive    bool   `db:"is_active" json:"is_active"`
	CreatedAt   int64  `db:"created_at" json:"created_at"`
	UpdatedAt   int64  `db:"updated_at" json:"updated_at"`
}

// CreateTenantRequest represents the request payload for creating a tenant
type CreateTenantRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
}

// UpdateTenantRequest represents the request payload for updating a tenant
type UpdateTenantRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
