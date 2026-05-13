package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// FoodItemRepository defines the interface for food item database operations
type FoodItemRepository interface {
	Create(foodItem *models.FoodItem) error
	GetByID(id string) (*models.FoodItem, error)
	Update(foodItem *models.FoodItem) error
	Delete(id string) error
	ListByTenant(tenantID string, offset, limit int) ([]models.FoodItem, int64, error)
}

// PostgreSQLFoodItemRepository implements FoodItemRepository using sqlx and RLS
type PostgreSQLFoodItemRepository struct {
	DB        *sqlx.DB
	TenantID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewFoodItemRepository creates a new food item repository instance with dependency injection
func NewFoodItemRepository(db *sqlx.DB, tenantID string) *PostgreSQLFoodItemRepository {
	return &PostgreSQLFoodItemRepository{
		DB:       db,
		TenantID: tenantID,
	}
}

// Create inserts a new food item into the database
// Note: RLS (Row-Level Security) policies will automatically filter by tenant_id
func (r *PostgreSQLFoodItemRepository) Create(f *models.FoodItem) error {
	now := time.Now().UTC()

	// Use Exec to insert the record (since we know the UUID)
	result, err := r.DB.Exec("INSERT INTO food_items (id, name, description, price, category, tenant_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		f.ID, f.Name, f.Description, f.Price, f.Category, r.TenantID, &now, &now)
	if err != nil {
		return err
	}

	// Check if rows affected > 0
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil // Insert failed due to unique constraint, which is expected
	}

	return nil
}

// GetByID retrieves a food item by its UUID
func (r *PostgreSQLFoodItemRepository) GetByID(id string) (*models.FoodItem, error) {
	var foodItem models.FoodItem
	query := `SELECT * FROM food_items WHERE id = $1`

	if err := r.DB.Get(&foodItem, query, id); err != nil {
		return nil, err
	}

	return &foodItem, nil
}

// Update updates an existing food item
func (r *PostgreSQLFoodItemRepository) Update(foodItem *models.FoodItem) error {
	result, err := r.DB.Exec("UPDATE food_items SET name = $2, description = $3, price = $4 WHERE id = $1",
		foodItem.ID, foodItem.Name, foodItem.Description, foodItem.Price)
	if err != nil {
		return err
	}

	// Check if rows affected > 0
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return err // Update failed due to unique constraint or no match
	}

	return err
}

// Delete removes a food item by its UUID
func (r *PostgreSQLFoodItemRepository) Delete(id string) error {
	result, err := r.DB.Exec("DELETE FROM food_items WHERE id = $1", id)
	if err != nil {
		return err
	}

	// Check if rows affected > 0
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return err // Delete failed due to no match
	}

	return err
}

// ListByTenant retrieves paginated food items for a specific tenant
// RLS policies ensure only tenant's data is returned
func (r *PostgreSQLFoodItemRepository) ListByTenant(tenantID string, offset, limit int) ([]models.FoodItem, int64, error) {

	// Count total items for pagination
	var total int64
	countQuery := `SELECT COUNT(*) FROM food_items WHERE current_setting('app.current_tenant_id')::TEXT = $1`
	if err := r.DB.Get(&total, countQuery, tenantID); err != nil {
		return nil, 0, err
	}

	// Select paginated items
	var foodItems []models.FoodItem
	selectQuery := `SELECT * FROM food_items 
                    WHERE current_setting('app.current_tenant_id')::TEXT = $1 
                    ORDER BY created_at DESC 
                    LIMIT $2 OFFSET $3`

	if err := r.DB.Select(&foodItems, selectQuery, tenantID, limit, offset); err != nil {
		return nil, 0, err
	}

	return foodItems, total, nil
}
