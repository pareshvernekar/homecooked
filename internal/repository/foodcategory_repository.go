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

// PostgreSQLFoodCategoryRepository implements FoodCategoryRepository using sqlx.
// Rows are scoped by composite primary key (tenant_id, id).
type PostgreSQLFoodCategoryRepository struct {
	DB       *sqlx.DB
	TenantID string
	Logger   *logger.Logger
}

// NewFoodCategoryRepository creates a new food category repository instance with dependency injection
func NewFoodCategoryRepository(db *sqlx.DB, logger *logger.Logger, tenantID string) *PostgreSQLFoodCategoryRepository {
	return &PostgreSQLFoodCategoryRepository{
		DB:       db,
		TenantID: tenantID,
		Logger:   logger,
	}
}

// ListByTenant retrieves all food categories for the repository's tenant.
func (r *PostgreSQLFoodCategoryRepository) ListByTenant(ctx context.Context) ([]*models.FoodCategory, error) {
	var categories []*models.FoodCategory
	r.Logger.Info(ctx, "ListByTenant: Fetching food categories for tenant", "tenant_id", r.TenantID)
	query := `
		SELECT id, tenant_id, name, COALESCE(description, '') as description,
		       COALESCE(is_active, true) as is_active, created_at, updated_at
		FROM food_category
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	err := r.DB.SelectContext(ctx, &categories, query, r.TenantID)
	if err != nil {
		r.Logger.Error(ctx, "ListByTenant: Failed to retrieve food categories from database", "tenant_id", r.TenantID, "error", err)
		return nil, err
	}
	r.Logger.Info(ctx, "ListByTenant: Successfully retrieved food categories", "tenant_id", r.TenantID, "count", len(categories))
	return categories, nil
}

// GetByID retrieves a food category by composite key (tenant_id, id).
func (r *PostgreSQLFoodCategoryRepository) GetByID(ctx context.Context, id string) (*models.FoodCategory, error) {
	var category models.FoodCategory
	r.Logger.Info(ctx, "GetByID: Fetching food category for tenant", "tenant_id", r.TenantID, "id", id)
	query := `
		SELECT id, tenant_id, name, COALESCE(description, '') as description,
		       COALESCE(is_active, true) as is_active, created_at, updated_at
		FROM food_category
		WHERE tenant_id = $1 AND id = $2`

	err := r.DB.GetContext(ctx, &category, query, r.TenantID, id)
	if err != nil {
		r.Logger.Error(ctx, "GetByID: Failed to retrieve food category from database", "tenant_id", r.TenantID, "id", id, "error", err)
		return nil, err
	}
	r.Logger.Info(ctx, "GetByID: Successfully retrieved food category", "tenant_id", r.TenantID, "id", id)
	return &category, nil
}

// Create inserts a new food category under the repository tenant.
// Enforces tenant FK before insert; always persists r.TenantID (ignores caller tenant_id).
func (r *PostgreSQLFoodCategoryRepository) Create(ctx context.Context, category *models.FoodCategory) error {
	r.Logger.Info(ctx, "Create: Creating new food category", "tenant_id", r.TenantID, "name", category.Name)

	if err := requireTenantExists(ctx, r.DB, r.TenantID); err != nil {
		r.Logger.Error(ctx, "Create: Tenant constraint failed", "tenant_id", r.TenantID, "error", err)
		return err
	}

	id := uuid.New().String()
	category.ID = id
	category.TenantID = r.TenantID
	currentTime := time.Now().UTC().UnixMilli()
	category.CreatedAt = currentTime
	category.UpdatedAt = currentTime

	query := `
		INSERT INTO food_category (id, tenant_id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.DB.ExecContext(ctx, query,
		category.ID, r.TenantID, category.Name, category.Description, category.IsActive, currentTime, currentTime)
	if err != nil {
		r.Logger.Error(ctx, "Create: Failed to create food category", "tenant_id", r.TenantID, "name", category.Name, "error", err)
		return err
	}
	r.Logger.Info(ctx, "Create: Successfully created food category", "tenant_id", r.TenantID, "id", category.ID, "name", category.Name)
	return nil
}

// Update updates an existing food category identified by composite key (tenant_id, id).
func (r *PostgreSQLFoodCategoryRepository) Update(ctx context.Context, category *models.FoodCategory) (int64, error) {
	r.Logger.Info(ctx, "Update: Updating food category", "tenant_id", r.TenantID, "id", category.ID)
	if category.ID == "" {
		return 0, fmt.Errorf("category id is required")
	}

	updatedAt := time.Now().UTC().UnixMilli()
	query := `UPDATE food_category SET name = $1, description = $2, is_active = $3, updated_at = $4 WHERE tenant_id = $5 AND id = $6`
	result, err := r.DB.ExecContext(ctx, query,
		category.Name, category.Description, category.IsActive, updatedAt, r.TenantID, category.ID)
	if err != nil {
		r.Logger.Error(ctx, "Update: Failed to update food category", "tenant_id", r.TenantID, "id", category.ID, "error", err)
		return 0, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.Logger.Warn(ctx, "Update: No rows affected when updating food category", "tenant_id", r.TenantID, "id", category.ID)
	}
	category.TenantID = r.TenantID
	category.UpdatedAt = updatedAt
	r.Logger.Info(ctx, "Update: Successfully updated food category", "tenant_id", r.TenantID, "id", category.ID)
	return rowsAffected, nil
}

// Delete soft-deletes a food category by composite key (tenant_id, id).
func (r *PostgreSQLFoodCategoryRepository) Delete(ctx context.Context, id string) (int64, error) {
	r.Logger.Info(ctx, "Delete: Deleting food category", "tenant_id", r.TenantID, "id", id)
	updatedAt := time.Now().UTC().UnixMilli()
	query := `UPDATE food_category SET is_active = FALSE, updated_at = $1 WHERE tenant_id = $2 AND id = $3`
	result, err := r.DB.ExecContext(ctx, query, updatedAt, r.TenantID, id)
	if err != nil {
		r.Logger.Error(ctx, "Delete: Failed to delete food category", "tenant_id", r.TenantID, "id", id, "error", err)
		return 0, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.Logger.Warn(ctx, "Delete: No rows affected when deleting food category", "tenant_id", r.TenantID, "id", id)
	}
	r.Logger.Info(ctx, "Delete: Successfully deleted food category", "tenant_id", r.TenantID, "id", id)
	return rowsAffected, nil
}
