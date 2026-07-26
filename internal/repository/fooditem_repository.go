package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// PostgreSQLFoodItemRepository implements FoodItemRepository using sqlx and RLS
type PostgreSQLFoodItemRepository struct {
	DB        *sqlx.DB
	TenantID  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Logger    *logger.Logger
}

// NewFoodItemRepository creates a new food item repository instance with dependency injection
func NewFoodItemRepository(db *sqlx.DB, logger *logger.Logger, tenantID string) *PostgreSQLFoodItemRepository {
	return &PostgreSQLFoodItemRepository{
		DB:       db,
		TenantID: tenantID,
		Logger:   logger,
	}
}

// GetByID retrieves a food item by its UUID
func (r *PostgreSQLFoodItemRepository) GetByID(ctx context.Context, tenantID string, id string) (*models.FoodItem, error) {
	var foodItem models.FoodItem
	r.Logger.Info(ctx, "GetByID: Fetching food item for tenant", "tenant_id", tenantID, "id", id)
	// Use explicit column selection to properly map pointer fields (time.Time pointers need explicit names)
	query := `SELECT id, tenant_id, name, COALESCE(description, ''), COALESCE(price, 0), category_id, created_at, updated_at FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1 AND id = $2`

	if err := r.DB.Get(&foodItem, query, tenantID, id); err != nil {
		r.Logger.Error(ctx, "GetByID: Failed to retrieve food item from database", "tenant_id", tenantID, "id", id, "error", err)
		return nil, err
	}
	r.Logger.Info(ctx, "GetByID: Successfully retrieved food item", "tenant_id", tenantID, "id", id)
	return &foodItem, nil
}

// Create inserts a new food item into the database
// Note: RLS (Row-Level Security) policies will automatically filter by tenant_id
func (r *PostgreSQLFoodItemRepository) Create(ctx context.Context, f *models.FoodItem) error {
	now := time.Now().UTC().UnixMilli()

	r.Logger.Info(ctx, "Create: Creating new food item", "tenant_id", r.TenantID, "id", f.ID, "name", f.Name)

	// Use Exec to insert the record (since we know the UUID)
	query := `INSERT INTO food_item (id, name, description, price, category_id, tenant_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	result, err := r.DB.Exec(query,
		f.ID, f.Name, f.Description, f.Price, f.CategoryID, r.TenantID, &now, &now)
	if err != nil {
		r.Logger.Error(ctx, "Create: Failed to create food item", "tenant_id", r.TenantID, "id", f.ID, "error", err)
		return err
	}

	// Check if rows affected > 0
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.Logger.Warn(ctx, "Create: Insert failed due to unique constraint", "tenant_id", r.TenantID, "id", f.ID)
		return nil // Insert failed due to unique constraint, which is expected
	}

	r.Logger.Info(ctx, "Create: Successfully created food item", "tenant_id", r.TenantID, "id", f.ID)
	return nil
}

// Update updates an existing food item
func (r *PostgreSQLFoodItemRepository) Update(ctx context.Context, foodItem *models.FoodItem) error {
	r.Logger.Info(ctx, "Update: Updating food item", "tenant_id", r.TenantID, "id", foodItem.ID)

	updatedAt := time.Now().UTC().UnixMilli()
	query := `UPDATE food_item SET name = $1, description = $2, price = $3, category_id = $4, updated_at = $5 WHERE id = $6 AND current_setting('app.current_tenant_id')::TEXT = $7`
	result, err := r.DB.Exec(query, foodItem.Name, foodItem.Description, foodItem.Price, foodItem.CategoryID, updatedAt, foodItem.ID, r.TenantID)
	if err != nil {
		r.Logger.Error(ctx, "Update: Failed to update food item", "tenant_id", r.TenantID, "id", foodItem.ID, "error", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.Logger.Warn(ctx, "Update: No rows affected when updating food item", "tenant_id", r.TenantID, "id", foodItem.ID)
		return err // Update failed due to unique constraint or no match
	}

	r.Logger.Info(ctx, "Update: Successfully updated food item", "tenant_id", r.TenantID, "id", foodItem.ID, "rows_affected", rowsAffected)
	return nil
}

// Delete removes a food item by its UUID
func (r *PostgreSQLFoodItemRepository) Delete(ctx context.Context, id string) (int64, error) {
	r.Logger.Info(ctx, "Delete: Deleting food item", "tenant_id", r.TenantID, "id", id)
	updatedAt := time.Now().UTC().UnixMilli()
	query := `UPDATE food_item SET is_active = FALSE, updated_at = $1 WHERE id = $2 AND current_setting('app.current_tenant_id')::TEXT = $3`
	result, err := r.DB.Exec(query, updatedAt, id, r.TenantID)
	if err != nil {
		r.Logger.Error(ctx, "Delete: Failed to delete food item", "tenant_id", r.TenantID, "id", id, "error", err)
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.Logger.Warn(ctx, "Delete: No rows affected when deleting food item", "tenant_id", r.TenantID, "id", id)
		return 0, err // Delete failed due to no match
	}

	r.Logger.Info(ctx, "Delete: Successfully deleted food item", "tenant_id", r.TenantID, "id", id)
	return rowsAffected, nil
}

// ListByTenant retrieves paginated food items for a specific tenant
// RLS policies ensure only tenant's data is returned
func (r *PostgreSQLFoodItemRepository) ListByTenant(ctx context.Context, tenantID string, offset, limit int) ([]*models.FoodItem, int64, error) {
	r.Logger.Info(ctx, "ListByTenant: Fetching paginated food items for tenant", "tenant_id", tenantID, "offset", offset, "limit", limit)

	// Count total items for pagination
	var total int64
	countQuery := `SELECT COUNT(*) FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1`
	if err := r.DB.Get(&total, countQuery, tenantID); err != nil {
		r.Logger.Error(ctx, "ListByTenant: Failed to count total items for pagination", "tenant_id", tenantID, "error", err)
		return nil, 0, err
	}

	// Select paginated items
	var foodItems []*models.FoodItem
	selectQuery := `SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(price, 0) as price, category_id as category_id, created_at, updated_at FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	if err := r.DB.Select(&foodItems, selectQuery, tenantID, limit, offset); err != nil {
		r.Logger.Error(ctx, "ListByTenant: Failed to retrieve paginated food items", "tenant_id", tenantID, "error", err)
		return nil, 0, err
	}

	r.Logger.Info(ctx, "ListByTenant: Successfully retrieved paginated food items", "tenant_id", tenantID, "offset", offset, "limit", limit, "total_count", total, "returned_count", len(foodItems))
	return foodItems, total, nil
}
