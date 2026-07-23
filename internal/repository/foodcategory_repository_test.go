package repository

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Suite: PostgreSQLFoodCategoryRepository
// =============================================================================

func TestPostgreSQLFoodCategoryRepository_ListByTenant_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock successful query result - DB.Select types slice as []T (values, not pointers) based on T parameter
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "is_active", "created_at", "updated_at"}).
		AddRow("cat-uuid-1", "tenant-123", "Vegetarian", "Contains no meat or animal products", true, time.Now().UnixMilli(), time.Now().UnixMilli())
	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(is_active, true) as is_active, created_at, updated_at FROM food_category WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC`)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-123").WillReturnRows(rows)

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	categories, err := repo.ListByTenant(t.Context(), "tenant-123")

	require.NoError(t, err)
	assert.Equal(t, 1, len(categories))
	// SQL items are returned as value types []T due to sqlx generics behavior with DB.Select
	firstCategory := categories[0]
	assert.NotNil(t, &firstCategory, "Expected at least one category")
	assert.Equal(t, "cat-uuid-1", firstCategory.ID)
	assert.Equal(t, "Vegetarian", firstCategory.Name)
	assert.Equal(t, "Contains no meat or animal products", firstCategory.Description)
	assert.Equal(t, "tenant-123", firstCategory.TenantID)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_ListByTenant_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query returns no rows - empty slice
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "is_active", "created_at", "updated_at"})
	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(is_active, true) as is_active, created_at, updated_at FROM food_category WHERE current_setting('app.current_tenant_id')::TEXT = $1`)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-456").WillReturnRows(rows)

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-456",
		Logger:   logger.NewLogger(),
	}

	categories, err := repo.ListByTenant(t.Context(), "tenant-456")

	require.NoError(t, err)
	assert.Empty(t, categories)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_ListByTenant_ErrorMessageOnFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query fails due to connection error
	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(is_active, true) as is_active, created_at, updated_at FROM food_category WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC`)
	mock.ExpectQuery(queryRegex).WillReturnError(fmt.Errorf("connection failed"))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-789",
		Logger:   logger.NewLogger(),
	}
	categories, err := repo.ListByTenant(t.Context(), "tenant-789")

	require.Error(t, err)
	assert.Nil(t, categories)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_ListByTenant_TenantIsolationRLS(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Simulate RLS behavior - different tenant IDs should return different results
	tenantARows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "is_active", "created_at", "updated_at"}).
		AddRow("cat-a-1", "tenant-A", "Veg Option 1", "First vegetarian option", true, time.Now().UnixMilli(), time.Now().UnixMilli()).
		AddRow("cat-a-2", "tenant-A", "Veg Option 2", "Second vegetarian option", true, time.Now().Add(-1*time.Hour).UnixMilli(), time.Now().Add(-1*time.Hour).UnixMilli())

	// Tenant B's query - different tenant ID
	tenantBRows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "is_active", "created_at", "updated_at"}).
		AddRow("cat-b-1", "tenant-B", "Non-Veg Option 1", "First non-vegetarian option", true, time.Now().UnixMilli(), time.Now().UnixMilli()).
		AddRow("cat-b-2", "tenant-B", "Non-Veg Option 2", "Second non-vegetarian option", true, time.Now().Add(-1*time.Hour).UnixMilli(), time.Now().Add(-1*time.Hour).UnixMilli())

	// Expect two separate queries with different tenant IDs
	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(is_active, true) as is_active, created_at, updated_at FROM food_category WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC`)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-A").WillReturnRows(tenantARows)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-B").WillReturnRows(tenantBRows)

	repoATenant := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-A",
		Logger:   logger.NewLogger(),
	}

	repoBTenant := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-B",
		Logger:   logger.NewLogger(),
	}

	// Query for tenant A
	categoriesA, errA := repoATenant.ListByTenant(t.Context(), "tenant-A")
	require.NoError(t, errA)
	assert.Equal(t, 2, len(categoriesA))
	for _, cat := range categoriesA {
		assert.Equal(t, "tenant-A", cat.TenantID)
	}

	// Query for tenant B
	categoriesB, errB := repoBTenant.ListByTenant(t.Context(), "tenant-B")
	require.NoError(t, errB)
	assert.Equal(t, 2, len(categoriesB))
	for _, cat := range categoriesB {
		assert.Equal(t, "tenant-B", cat.TenantID)
	}

	// Verify mock expectations were all met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_ListByTenant_MultipleItemsWithNullDescription(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Include NULL description for some categories
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "is_active", "created_at", "updated_at"}).
		AddRow("cat-1", "tenant-xyz", "Pizza Category", "Dessert items", true, time.Now().UnixMilli(), time.Now().UnixMilli()). // Has description
		AddRow("cat-2", "tenant-xyz", "Hot & Spicy", "", true, time.Now().UnixMilli(), time.Now().UnixMilli())                  // NULL description

	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(is_active, true) as is_active, created_at, updated_at FROM food_category WHERE current_setting('app.current_tenant_id')::TEXT = $1`)
	mock.ExpectQuery(queryRegex).WillReturnRows(rows)

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-xyz",
		Logger:   logger.NewLogger(),
	}

	categories, err := repo.ListByTenant(t.Context(), "tenant-xyz")

	require.NoError(t, err)
	assert.Equal(t, 2, len(categories))

	// First category has description
	desc1 := categories[0].Description
	assert.Equal(t, "Dessert items", desc1)

	// Second category has NULL description - should be empty string due to sqlx behavior
	desc2 := categories[1].Description
	assert.Empty(t, desc2, "NULL values are returned as empty strings by sqlx")

	assert.Equal(t, "Pizza Category", categories[0].Name)
	assert.Equal(t, "Hot & Spicy", categories[1].Name)
	assert.Equal(t, "tenant-xyz", categories[1].TenantID)

	// Verify mock expectations
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock successful query result - DB.Select types slice as []T (values, not pointers) based on T parameter
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "is_active", "created_at", "updated_at"}).
		AddRow("cat-uuid-123", "tenant-123", "Vegetarian", "Contains no meat or animal products", true, time.Now().UnixMilli(), time.Now().UnixMilli())
	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(is_active, true) as is_active, created_at, updated_at FROM food_category WHERE current_setting('app.current_tenant_id')::TEXT = $1 AND id = $2`)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-123", "cat-uuid-123").WillReturnRows(rows)

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}
	category, err := repo.GetByID(t.Context(), "tenant-123", "cat-uuid-123")

	require.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, "cat-uuid-123", category.ID)
	assert.Equal(t, "Vegetarian", category.Name)
	assert.Equal(t, "Contains no meat or animal products", category.Description)
	assert.Equal(t, "tenant-123", category.TenantID)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query returns no rows because ID doesn't exist
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "is_active", "created_at", "updated_at"})

	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(is_active, true) as is_active, created_at, updated_at FROM food_category WHERE current_setting('app.current_tenant_id')::TEXT = $1 AND id = $2`)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-123", "cat-nonexistent").WillReturnRows(rows)

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	category, err := repo.GetByID(t.Context(), "tenant-123", "cat-nonexistent")

	require.Error(t, err, "Expected error for non-existent category")
	assert.Nil(t, category)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_GetByID_ErrorMessageOnFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query fails due to connection error
	mock.ExpectQuery(`SELECT id, tenant_id, name`).WillReturnError(fmt.Errorf("database connection error"))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-999",
		Logger:   logger.NewLogger(),
	}
	category, err := repo.GetByID(t.Context(), "tenant-999", "cat-uuid-xyz")

	require.Error(t, err)
	assert.Nil(t, category)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	currentTime := time.Now().UTC().UnixMilli()
	// Mock successful INSERT - Exec returns single value (not slice)
	mock.ExpectExec(`INSERT INTO food_category`).WithArgs("cat-uuid-123", "tenant-123", "Vegetarian", "Fresh vegetables and dairy", true, currentTime, currentTime).WillReturnResult(sqlmock.NewResult(123, 1))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}
	category := models.FoodCategory{
		ID:          "cat-uuid-123",
		TenantID:    "tenant-123",
		Name:        "Vegetarian",
		IsActive:    true,
		Description: "Fresh vegetables and dairy",
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}

	err = repo.Create(t.Context(), &category)

	require.NoError(t, err)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_Create_FailDuplicateID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	currentTime := time.Now().UTC().UnixMilli()
	// Mock INSERT fails due to unique constraint violation on id
	mock.ExpectExec(`INSERT INTO food_category`).WithArgs("cat-uuid-already-exists", "tenant-123", "Vegetarian", "Description", true, currentTime, currentTime).WillReturnError(fmt.Errorf("duplicate key value violates unique constraint \"food_category_pkey\""))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	category := models.FoodCategory{
		ID:          "cat-uuid-already-exists",
		TenantID:    "tenant-123",
		Name:        "Vegetarian",
		Description: "Description",
		IsActive:    true,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}

	err = repo.Create(t.Context(), &category)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_Create_FailDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock INSERT fails due to database error
	currentTime := time.Now().UTC().UnixMilli()
	mock.ExpectExec(`INSERT INTO food_category`).WithArgs("cat-uuid-new", "tenant-123", "Vegetarian", "Description", true, currentTime, currentTime).WillReturnError(fmt.Errorf("constraint violation"))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	category := models.FoodCategory{
		ID:          "cat-uuid-new",
		TenantID:    "tenant-123",
		Name:        "Vegetarian",
		Description: "Description",
		IsActive:    true,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}

	err = repo.Create(t.Context(), &category)

	require.Error(t, err)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")
	currentTime := time.Now().UTC().UnixMilli()
	// Mock successful UPDATE - Exec returns single value (not slice)
	queryRegex := regexp.QuoteMeta(`UPDATE food_category SET name = $1, description = $2, is_active = $3, updated_at = $4 WHERE id = $5 AND tenant_id = $6`)
	mock.ExpectExec(queryRegex).WithArgs("Updated Veg Name", "Updated description", true, currentTime, "cat-uuid-123", "tenant-123").WillReturnResult(sqlmock.NewResult(10, 1))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	category := models.FoodCategory{
		ID:          "cat-uuid-123",
		TenantID:    "tenant-123",
		Name:        "Updated Veg Name",
		Description: "Updated description",
		IsActive:    true,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}

	rowsAffected, err := repo.Update(t.Context(), &category)

	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected, "Expected 1 row to be affected by update")
	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_Update_FailNotExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")
	currentTime := time.Now().UTC().UnixMilli()
	// Mock UPDATE affects 0 rows because ID doesn't exist
	queryRegex := regexp.QuoteMeta(`UPDATE food_category SET name = $1, description = $2, is_active = $3, updated_at = $4 WHERE id = $5 AND tenant_id = $6`)
	mock.ExpectExec(queryRegex).WithArgs("New Name", "New Description", true, currentTime, "cat-nonexistent", "tenant-123").WillReturnResult(sqlmock.NewResult(0, 0))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	category := models.FoodCategory{
		ID:          "cat-nonexistent",
		TenantID:    "tenant-123",
		Name:        "New Name",
		Description: "New Description",
		IsActive:    true,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}

	rowsAffected, err := repo.Update(t.Context(), &category)

	require.NoError(t, err, "Expected error when updating non-existent category")
	assert.Equal(t, int64(0), rowsAffected, "Expected 0 rows to be affected when updating non-existent category")
	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_Update_FailDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")
	currentTime := time.Now().UTC().UnixMilli()

	// Mock UPDATE fails due to database error
	queryRegex := regexp.QuoteMeta(`UPDATE food_category SET name = $1, description = $2, is_active = $3, updated_at = $4 WHERE id = $5 AND tenant_id = $6`)
	mock.ExpectExec(queryRegex).WithArgs("New Name", "New Description", true, currentTime, "cat-uuid-xyz", "tenant-123").WillReturnError(fmt.Errorf("database error: constraint violation"))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	category := models.FoodCategory{
		ID:          "cat-uuid-xyz",
		TenantID:    "tenant-123",
		Name:        "New Name",
		Description: "New Description",
		IsActive:    true,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}

	rowsAffected, err := repo.Update(t.Context(), &category)

	require.Error(t, err)
	assert.Equal(t, int64(0), rowsAffected, "Expected 0 rows to be affected when update fails due to database error")
	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock successful UPDATE for soft delete (setting is_active = FALSE)
	updatedAt := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`UPDATE food_category SET is_active = FALSE, updated_at = $1 WHERE id = $2 AND tenant_id = $3`)
	mock.ExpectExec(queryRegex).WithArgs(updatedAt, "cat-uuid-123", "tenant-123").WillReturnResult(sqlmock.NewResult(10, 1))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	rowsAffected, err := repo.Delete(t.Context(), "tenant-123", "cat-uuid-123")

	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected, "Expected 1 row to be affected by delete")
	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_Delete_FailNotExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock UPDATE affects 0 rows because ID doesn't exist or soft delete already applied
	updatedAt := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`UPDATE food_category SET is_active = FALSE, updated_at = $1 WHERE id = $2 AND tenant_id = $3`)
	mock.ExpectExec(queryRegex).WithArgs(updatedAt, "cat-nonexistent", "tenant-123").WillReturnResult(sqlmock.NewResult(0, 0))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	ctx := t.Context()
	rowsAffected, err := repo.Delete(ctx, "tenant-123", "cat-nonexistent")

	require.NoError(t, err)
	assert.Equal(t, int64(0), rowsAffected, "Expected 0 rows to be affected when deleting non-existent category")
	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_Delete_FailDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock UPDATE fails due to database error
	updatedAt := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`UPDATE food_category SET is_active = FALSE, updated_at = $1 WHERE id = $2 AND tenant_id = $3`)
	mock.ExpectExec(queryRegex).WithArgs(updatedAt, "cat-uuid-xyz", "tenant-123").WillReturnError(fmt.Errorf("database error: constraint violation"))

	repo := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	rowsAffected, err := repo.Delete(t.Context(), "tenant-123", "cat-uuid-xyz")

	require.Error(t, err)
	assert.Equal(t, int64(0), rowsAffected, "Expected 0 rows to be affected when delete fails due to database error")

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodCategoryRepository_Delete_TenantIsolation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Create two categories with different IDs but same tenant - only one should be deleted
	category1ID := "cat-uuid-tenant-a"
	category2ID := "cat-uuid-tenant-b"

	updatedAt := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`UPDATE food_category SET is_active = FALSE, updated_at = $1 WHERE id = $2 AND tenant_id = $3`)
	mock.ExpectExec(queryRegex).WithArgs(updatedAt, category1ID, "tenant-A").WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectExec(queryRegex).WithArgs(updatedAt, category2ID, "tenant-B").WillReturnResult(sqlmock.NewResult(10, 1))

	repoATenant := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-A",
		Logger:   logger.NewLogger(),
	}

	repoBTenant := &PostgreSQLFoodCategoryRepository{
		DB:       dbx,
		TenantID: "tenant-B",
		Logger:   logger.NewLogger(),
	}

	// Delete from tenant A - should only affect category with ID cat-uuid-tenant-a
	rowsAffectedA, errA := repoATenant.Delete(t.Context(), "tenant-A", category1ID)
	require.NoError(t, errA)
	assert.Equal(t, int64(1), rowsAffectedA, "Expected 1 row to be affected when deleting category for tenant A")

	// Delete from tenant B - should only affect category with ID cat-uuid-tenant-b
	rowsAffectedB, errB := repoBTenant.Delete(t.Context(), "tenant-B", category2ID)
	require.NoError(t, errB)
	assert.Equal(t, int64(1), rowsAffectedB, "Expected 1 row to be affected when deleting category for tenant B")

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}
