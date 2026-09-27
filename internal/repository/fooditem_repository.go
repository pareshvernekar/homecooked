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

// PostgreSQLFoodItemRepository implements FoodItemRepository using sqlx.
// Rows are scoped by composite primary key (tenant_id, id).
// category_id must belong to the same tenant (composite FK).
type PostgreSQLFoodItemRepository struct {
	DB       *sqlx.DB
	TenantID string
	Logger   *logger.Logger
}

// NewFoodItemRepository creates a new food item repository instance with dependency injection
func NewFoodItemRepository(db *sqlx.DB, logger *logger.Logger, tenantID string) *PostgreSQLFoodItemRepository {
	return &PostgreSQLFoodItemRepository{
		DB:       db,
		TenantID: tenantID,
		Logger:   logger,
	}
}

// GetByID retrieves a food item by composite key (tenant_id, id).
func (r *PostgreSQLFoodItemRepository) GetByID(ctx context.Context, id string) (*models.FoodItem, error) {
	var foodItem models.FoodItem
	r.Logger.Info(ctx, "GetByID: Fetching food item for tenant", "tenant_id", r.TenantID, "id", id)
	query := `
		SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(price, 0) as price,
		       category_id, COALESCE(image_url, '') as image_url, COALESCE(avoidance, '') as avoidance,
		       COALESCE(is_vegetarian, false) as is_vegetarian,
		       COALESCE(availability_status, 'available') as availability_status, created_at, updated_at
		FROM food_item
		WHERE tenant_id = $1 AND id = $2`

	if err := r.DB.GetContext(ctx, &foodItem, query, r.TenantID, id); err != nil {
		r.Logger.Error(ctx, "GetByID: Failed to retrieve food item from database", "tenant_id", r.TenantID, "id", id, "error", err)
		return nil, err
	}
	r.Logger.Info(ctx, "GetByID: Successfully retrieved food item", "tenant_id", r.TenantID, "id", id)
	return &foodItem, nil
}

// Create inserts a new food item under the repository tenant.
// Enforces tenant FK and composite category FK (tenant_id, category_id).
func (r *PostgreSQLFoodItemRepository) Create(ctx context.Context, f *models.FoodItem) error {
	now := time.Now().UTC().UnixMilli()

	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	f.TenantID = r.TenantID

	r.Logger.Info(ctx, "Create: Creating new food item", "tenant_id", r.TenantID, "id", f.ID, "name", f.Name)

	if err := requireTenantExists(ctx, r.DB, r.TenantID); err != nil {
		r.Logger.Error(ctx, "Create: Tenant constraint failed", "tenant_id", r.TenantID, "error", err)
		return err
	}
	if err := requireCategoryInTenant(ctx, r.DB, r.TenantID, f.CategoryID); err != nil {
		r.Logger.Error(ctx, "Create: Category tenant constraint failed",
			"tenant_id", r.TenantID, "category_id", f.CategoryID, "error", err)
		return err
	}

	query := `INSERT INTO food_item (id, name, description, price, category_id, availability_status, tenant_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	result, err := r.DB.ExecContext(ctx, query,
		f.ID, f.Name, f.Description, f.Price, f.CategoryID, f.AvailabilityStatus, r.TenantID, now, now)
	if err != nil {
		r.Logger.Error(ctx, "Create: Failed to create food item", "tenant_id", r.TenantID, "id", f.ID, "error", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("create food item affected 0 rows for tenant %s id %s", r.TenantID, f.ID)
	}

	f.CreatedAt = now
	f.UpdatedAt = now
	r.Logger.Info(ctx, "Create: Successfully created food item", "tenant_id", r.TenantID, "id", f.ID)
	return nil
}

// Update updates an existing food item identified by composite key (tenant_id, id).
// Re-validates category belongs to the same tenant when category_id is set.
func (r *PostgreSQLFoodItemRepository) Update(ctx context.Context, id string, foodItem *models.FoodItem) error {
	r.Logger.Info(ctx, "Update: Updating food item", "tenant_id", r.TenantID, "id", id)
	if id == "" {
		return fmt.Errorf("food item id is required")
	}
	if foodItem.CategoryID != "" {
		if err := requireCategoryInTenant(ctx, r.DB, r.TenantID, foodItem.CategoryID); err != nil {
			r.Logger.Error(ctx, "Update: Category tenant constraint failed",
				"tenant_id", r.TenantID, "category_id", foodItem.CategoryID, "error", err)
			return err
		}
	}

	updatedAt := time.Now().UTC().UnixMilli()
	query := `UPDATE food_item SET name = $1, description = $2, price = $3, category_id = $4, image_url = $5, avoidance = $6, is_vegetarian = $7, availability_status = $8, updated_at = $9 WHERE tenant_id = $10 AND id = $11`
	result, err := r.DB.ExecContext(ctx, query,
		foodItem.Name, foodItem.Description, foodItem.Price, foodItem.CategoryID,
		foodItem.ImageURL, foodItem.Avoidance, foodItem.IsVegetarian, foodItem.AvailabilityStatus,
		updatedAt, r.TenantID, id)
	if err != nil {
		r.Logger.Error(ctx, "Update: Failed to update food item", "tenant_id", r.TenantID, "id", id, "error", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.Logger.Warn(ctx, "Update: No rows affected when updating food item", "tenant_id", r.TenantID, "id", id)
		return fmt.Errorf("food item %q not found for tenant %q", id, r.TenantID)
	}

	foodItem.TenantID = r.TenantID
	foodItem.UpdatedAt = updatedAt
	r.Logger.Info(ctx, "Update: Successfully updated food item", "tenant_id", r.TenantID, "id", id, "rows_affected", rowsAffected)
	return nil
}

// Delete soft-deletes a food item by composite key (tenant_id, id).
func (r *PostgreSQLFoodItemRepository) Delete(ctx context.Context, id string) (int64, error) {
	r.Logger.Info(ctx, "Delete: Deleting food item", "tenant_id", r.TenantID, "id", id)
	updatedAt := time.Now().UTC().UnixMilli()
	query := `UPDATE food_item SET is_active = FALSE, updated_at = $1 WHERE tenant_id = $2 AND id = $3`
	result, err := r.DB.ExecContext(ctx, query, updatedAt, r.TenantID, id)
	if err != nil {
		r.Logger.Error(ctx, "Delete: Failed to delete food item", "tenant_id", r.TenantID, "id", id, "error", err)
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.Logger.Warn(ctx, "Delete: No rows affected when deleting food item", "tenant_id", r.TenantID, "id", id)
		return 0, nil
	}

	r.Logger.Info(ctx, "Delete: Successfully deleted food item", "tenant_id", r.TenantID, "id", id)
	return rowsAffected, nil
}

// ListByTenant retrieves paginated food items for the repository's tenant.
func (r *PostgreSQLFoodItemRepository) ListByTenant(ctx context.Context, offset, limit int) ([]*models.FoodItem, int64, error) {
	r.Logger.Info(ctx, "ListByTenant: Fetching paginated food items for tenant", "tenant_id", r.TenantID, "offset", offset, "limit", limit)

	var total int64
	countQuery := `SELECT COUNT(*) FROM food_item WHERE tenant_id = $1`
	if err := r.DB.GetContext(ctx, &total, countQuery, r.TenantID); err != nil {
		r.Logger.Error(ctx, "ListByTenant: Failed to count total items for pagination", "tenant_id", r.TenantID, "error", err)
		return nil, 0, err
	}

	var foodItems []*models.FoodItem
	selectQuery := `SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(price, 0) as price, category_id as category_id, created_at, updated_at FROM food_item WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	if err := r.DB.SelectContext(ctx, &foodItems, selectQuery, r.TenantID, limit, offset); err != nil {
		r.Logger.Error(ctx, "ListByTenant: Failed to retrieve paginated food items", "tenant_id", r.TenantID, "error", err)
		return nil, 0, err
	}

	r.Logger.Info(ctx, "ListByTenant: Successfully retrieved paginated food items",
		"tenant_id", r.TenantID, "offset", offset, "limit", limit, "total_count", total, "returned_count", len(foodItems))
	return foodItems, total, nil
}
