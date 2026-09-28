package models

// SizeUnit represents a system standard or tenant-defined size unit.
// REQSIZE001, REQSIZE002
type SizeUnit struct {
	ID          string  `db:"id" json:"id"`
	TenantID    *string `db:"tenant_id" json:"tenant_id,omitempty"`
	Code        string  `db:"code" json:"code"`
	DisplayName string  `db:"display_name" json:"display_name"`
	IsSystem    bool    `db:"is_system" json:"is_system"`
	IsActive    bool    `db:"is_active" json:"is_active"`
	CreatedAt   int64   `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt   int64   `db:"updated_at" json:"updated_at,omitempty"`
}

// SizeUnitCreateRequest is the payload for creating a tenant custom size unit.
// REQSIZE002
type SizeUnitCreateRequest struct {
	Code        string `json:"code" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
}
