package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// PostgreSQLSizeUnitRepository implements size unit data access.
// REQSIZE001–REQSIZE003
type PostgreSQLSizeUnitRepository struct {
	DB       *sqlx.DB
	TenantID string
	Logger   *logger.Logger
}

// NewSizeUnitRepository creates a size unit repository scoped to a tenant.
func NewSizeUnitRepository(db *sqlx.DB, l *logger.Logger, tenantID string) *PostgreSQLSizeUnitRepository {
	return &PostgreSQLSizeUnitRepository{DB: db, TenantID: tenantID, Logger: l}
}

// ListActive returns system standards plus the tenant's active custom units.
// REQSIZE001, REQSIZE002
func (r *PostgreSQLSizeUnitRepository) ListActive(ctx context.Context) ([]*models.SizeUnit, error) {
	var units []*models.SizeUnit
	query := `
		SELECT id, tenant_id, code, display_name, is_system, is_active, created_at, updated_at
		FROM size_unit
		WHERE is_active = TRUE
		  AND (is_system = TRUE OR tenant_id = $1)
		ORDER BY is_system DESC, code ASC`
	if err := r.DB.SelectContext(ctx, &units, query, r.TenantID); err != nil {
		r.Logger.Error(ctx, "ListActive: failed to list size units", "tenant_id", r.TenantID, "error", err)
		return nil, err
	}
	return units, nil
}

// GetByID returns a size unit usable by the current tenant (system or own custom).
// REQSIZE003
func (r *PostgreSQLSizeUnitRepository) GetByID(ctx context.Context, id string) (*models.SizeUnit, error) {
	var unit models.SizeUnit
	query := `
		SELECT id, tenant_id, code, display_name, is_system, is_active, created_at, updated_at
		FROM size_unit
		WHERE id = $1
		  AND is_active = TRUE
		  AND (is_system = TRUE OR tenant_id = $2)`
	if err := r.DB.GetContext(ctx, &unit, query, id, r.TenantID); err != nil {
		return nil, err
	}
	return &unit, nil
}

// CreateCustom inserts a tenant-scoped custom size unit.
// REQSIZE002
func (r *PostgreSQLSizeUnitRepository) CreateCustom(ctx context.Context, code, displayName string) (*models.SizeUnit, error) {
	now := time.Now().UTC().UnixMilli()
	id := uuid.New().String()
	tenantID := r.TenantID

	query := `
		INSERT INTO size_unit (id, tenant_id, code, display_name, is_system, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, FALSE, TRUE, $5, $5)`
	if _, err := r.DB.ExecContext(ctx, query, id, tenantID, code, displayName, now); err != nil {
		r.Logger.Error(ctx, "CreateCustom: failed", "tenant_id", r.TenantID, "code", code, "error", err)
		return nil, err
	}

	return &models.SizeUnit{
		ID:          id,
		TenantID:    &tenantID,
		Code:        code,
		DisplayName: displayName,
		IsSystem:    false,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// SoftDeactivateCustom soft-deletes a tenant custom unit. System units cannot be deactivated.
// REQSIZE001, REQSIZE002
func (r *PostgreSQLSizeUnitRepository) SoftDeactivateCustom(ctx context.Context, id string) error {
	now := time.Now().UTC().UnixMilli()
	query := `
		UPDATE size_unit
		SET is_active = FALSE, updated_at = $1
		WHERE id = $2
		  AND tenant_id = $3
		  AND is_system = FALSE
		  AND is_active = TRUE`
	result, err := r.DB.ExecContext(ctx, query, now, id, r.TenantID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("size unit %q not found or not deletable for tenant %q", id, r.TenantID)
	}
	return nil
}
