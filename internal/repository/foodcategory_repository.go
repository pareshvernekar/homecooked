package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// PostgreSQLFoodCategoryRepository implements FoodCategoryRepository using sqlx and RLS
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

// ListByTenant retrieves all food categories for a specific tenant using RLS
func (r *PostgreSQLFoodCategoryRepository) ListByTenant(ctx context.Context, tenantID string) ([]*models.FoodCategory, error) {
	var categories []*models.FoodCategory
	r.Logger.Info(ctx, "ListByTenant: Fetching food categories for tenant", "tenant_id", tenantID)
	// Query food categories for this tenant using parameterized SQL
	// RLS policies ensure only tenant's data is returned
	query := `SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(is_active, true) as is_active, created_at, updated_at FROM food_category WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC`

	err := r.DB.Select(&categories, query, tenantID)
	if err != nil {
		r.Logger.Error(ctx, "ListByTenant: Failed to retrieve food categories from database", "tenant_id", tenantID, "error", err)
		return nil, err
	}
	r.Logger.Info(ctx, "ListByTenant: Successfully retrieved food categories", "tenant_id", tenantID, "count", len(categories))
	return categories, nil
}

// ListByTenant retrieves all food categories for a specific tenant using RLS
func (r *PostgreSQLFoodCategoryRepository) GetByID(ctx context.Context, tenantID string, id string) (*models.FoodCategory, error) {
	var category models.FoodCategory
	r.Logger.Info(ctx, "GetByID: Fetching food category for tenant", "tenant_id", tenantID, "id", id)
	// Query food category for this tenant using parameterized SQL
	// RLS policies ensure only tenant's data is returned
	query := `SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(is_active, true) as is_active, created_at, updated_at FROM food_category WHERE current_setting('app.current_tenant_id')::TEXT = $1 AND id = $2`

	err := r.DB.Get(&category, query, tenantID, id)
	if err != nil {
		r.Logger.Error(ctx, "GetByID: Failed to retrieve food category from database", "tenant_id", tenantID, "id", id, "error", err)
		return nil, err
	}
	r.Logger.Info(ctx, "GetByID: Successfully retrieved food category", "tenant_id", tenantID, "id", id)
	return &category, nil
}

func (r *PostgreSQLFoodCategoryRepository) Create(ctx context.Context, category *models.FoodCategory) error {
	r.Logger.Info(ctx, "Create: Creating new food category", "tenant_id", r.TenantID, "name", category.Name)
	// Insert the new food category into the database
	query := `INSERT INTO food_category (id, tenant_id, name, description, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	currentTime := time.Now().UTC().UnixMilli()
	_, err := r.DB.Exec(query, category.ID, r.TenantID, category.Name, category.Description, category.IsActive, currentTime, currentTime)
	if err != nil {
		r.Logger.Error(ctx, "Create: Failed to create food category", "tenant_id", r.TenantID, "name", category.Name, "error", err)
		return err
	}
	r.Logger.Info(ctx, "Create: Successfully created food category", "tenant_id", r.TenantID, "name", category.Name)
	return nil
}

func (r *PostgreSQLFoodCategoryRepository) Update(ctx context.Context, category *models.FoodCategory) (int64, error) {

	r.Logger.Info(ctx, "Update: Updating food category", "tenant_id", r.TenantID, "id", category.ID)
	// Update the existing food category in the database
	query := `UPDATE food_category SET name = $1, description = $2, is_active = $3, updated_at = $4 WHERE id = $5 AND current_setting('app.current_tenant_id')::TEXT = $6`
	updatedAt := time.Now().UTC().UnixMilli()
	result, err := r.DB.Exec(query, category.Name, category.Description, category.IsActive, updatedAt, category.ID, r.TenantID)
	if err != nil {
		r.Logger.Error(ctx, "Update: Failed to update food category", "tenant_id", r.TenantID, "id", category.ID, "error", err)
		return 0, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.Logger.Info(ctx, "Update: No rows affected when updating food category", "tenant_id", r.TenantID, "id", category.ID)
	}
	r.Logger.Info(ctx, "Update: Successfully updated food category", "tenant_id", r.TenantID, "id", category.ID)
	return rowsAffected, nil
}

func (r *PostgreSQLFoodCategoryRepository) Delete(ctx context.Context, tenantID string, id string) (int64, error) {
	r.Logger.Info(ctx, "Delete: Deleting food category", "tenant_id", tenantID, "id", id)
	// Delete the food category from the database
	updatedAt := time.Now().UTC().UnixMilli()
	query := `UPDATE food_category SET is_active = FALSE, updated_at = $1 WHERE id = $2 AND current_setting('app.current_tenant_id')::TEXT = $3`
	result, err := r.DB.Exec(query, updatedAt, id, tenantID)
	if err != nil {
		r.Logger.Error(ctx, "Delete: Failed to delete food category", "tenant_id", tenantID, "id", id, "error", err)
		return 0, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.Logger.Info(ctx, "Delete: No rows affected when deleting food category", "tenant_id", tenantID, "id", id)
	}
	r.Logger.Info(ctx, "Delete: Successfully deleted food category", "tenant_id", tenantID, "id", id)
	return rowsAffected, nil
}
