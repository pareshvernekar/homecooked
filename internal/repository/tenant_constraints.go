package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// requireTenantExists enforces that tenant_id references an active tenant row.
func requireTenantExists(ctx context.Context, db *sqlx.DB, tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	var exists bool
	err := db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM tenant WHERE id = $1 AND is_active = TRUE
		)
	`, tenantID)
	if err != nil {
		return fmt.Errorf("verify tenant %q: %w", tenantID, err)
	}
	if !exists {
		return fmt.Errorf("tenant %q does not exist or is inactive", tenantID)
	}
	return nil
}

// requireCategoryInTenant enforces the composite FK (tenant_id, category_id).
func requireCategoryInTenant(ctx context.Context, db *sqlx.DB, tenantID, categoryID string) error {
	if categoryID == "" {
		return fmt.Errorf("category_id is required")
	}
	var exists bool
	err := db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM food_category
			WHERE tenant_id = $1 AND id = $2 AND is_active = TRUE
		)
	`, tenantID, categoryID)
	if err != nil {
		return fmt.Errorf("verify category %q for tenant %q: %w", categoryID, tenantID, err)
	}
	if !exists {
		return fmt.Errorf("category %q not found for tenant %q", categoryID, tenantID)
	}
	return nil
}
