package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// FoodCategoryRepository defines the interface for food category database operations
type FoodCategoryRepository interface {
	ListByTenant(tenantID string) ([]models.FoodCategory, error)
}

// PostgreSQLFoodCategoryRepository implements FoodCategoryRepository using sqlx and RLS
type PostgreSQLFoodCategoryRepository struct {
	DB     *sqlx.DB
	TenantID string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewFoodCategoryRepository creates a new food category repository instance with dependency injection
func NewFoodCategoryRepository(db *sqlx.DB, tenantID string) *PostgreSQLFoodCategoryRepository {
	return &PostgreSQLFoodCategoryRepository{
		DB:         db,
		TenantID:   tenantID,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
}

// ListByTenant retrieves all food categories for a specific tenant using RLS
func (r *PostgreSQLFoodCategoryRepository) ListByTenant(tenantID string) ([]models.FoodCategory, error) {
	var categories []models.FoodCategory

	// Query food categories for this tenant using parameterized SQL
	// RLS policies ensure only tenant's data is returned
	query := `SELECT id, tenant_id, name, description, created_at, updated_at FROM food_category WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC`

	err := r.DB.Select(&categories, query, tenantID)
	if err != nil {
		return nil, err
	}

	return categories, nil
}
