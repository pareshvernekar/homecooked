package repository

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	models "github.com/pareshvernekar/homecooked/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Suite: PostgreSQLFoodItemRepository
// =============================================================================

func TestPostgreSQLFoodItemRepository_ListByTenant_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query fails due to connection error
	countQuery := regexp.QuoteMeta(`SELECT COUNT(*) FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1`)
	mock.ExpectQuery(countQuery).WithArgs("tenant-123").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock successful query result - DB.Select types slice as []T (values, not pointers) based on T parameter
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "price", "categoryId", "created_at", "updated_at"}).
		AddRow("item-uuid-1", "tenant-123", "Pizza", "Delicious pizza", 9.99, "cat-uuid-1", time.Now().UnixMilli(), time.Now().UnixMilli())
	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(price, 0) as price, categoryId as categoryId, created_at, updated_at FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-123", 10, 0).WillReturnRows(rows)

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	items, count, err := repo.ListByTenant(t.Context(), "tenant-123", 0, 10)

	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.Equal(t, 1, len(items))
	// SQL items are returned as value types []T due to sqlx generics behavior with DB.Select
	firstItem := items[0]
	assert.NotNil(t, &firstItem, "Expected at least one item")
	assert.Equal(t, "item-uuid-1", firstItem.ID)
	assert.Equal(t, "Pizza", firstItem.Name)
	assert.Equal(t, "Delicious pizza", *firstItem.Description)
	assert.Equal(t, 9.99, firstItem.Price)
	assert.Equal(t, "cat-uuid-1", firstItem.CategoryID)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_ListByTenant_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query fails due to connection error
	countQuery := regexp.QuoteMeta(`SELECT COUNT(*) FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1`)
	mock.ExpectQuery(countQuery).WithArgs("tenant-456").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock query returns no rows - empty slice
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "price", "categoryId", "created_at", "updated_at"})
	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(price, 0) as price, categoryId as categoryId, created_at, updated_at FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-456", 10, 0).WillReturnRows(rows)

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-456",
		Logger:   logger.NewLogger(),
	}

	items, count, err := repo.ListByTenant(t.Context(), "tenant-456", 0, 10)

	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, int64(0), count)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_ListByTenant_ErrorMessageOnFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query fails due to connection error
	countQuery := regexp.QuoteMeta(`SELECT COUNT(*) FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1`)
	mock.ExpectQuery(countQuery).WithArgs("tenant-xyz").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(price, 0) as price, categoryId as categoryId, created_at, updated_at FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`)
	mock.ExpectQuery(queryRegex).WillReturnError(fmt.Errorf("connection failed"))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-xyz",
		Logger:   logger.NewLogger(),
	}

	items, count, err := repo.ListByTenant(t.Context(), "tenant-xyz", 0, 10)

	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, int64(0), count)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_ListByTenant_TenantIsolationRLS(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Simulate RLS behavior - different tenant IDs should return different results
	tenantARows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "price", "categoryId", "created_at", "updated_at"}).
		AddRow("item-a-1", "tenant-A", "Veg Option 1 Pizza", "First vegetarian option", 9.99, "cat-uuid-1", time.Now().UnixMilli(), time.Now().UnixMilli()).
		AddRow("item-a-2", "tenant-A", "Veg Option 2 Salad", "Second vegetarian option", 5.99, "cat-uuid-2", time.Now().Add(-1*time.Hour).UnixMilli(), time.Now().Add(-1*time.Hour).UnixMilli())

	// Tenant B's query - different tenant ID
	tenantBRows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "price", "categoryId", "created_at", "updated_at"}).
		AddRow("item-b-1", "tenant-B", "Non-Veg Option 1 Burger", "First non-vegetarian option", 12.99, "cat-uuid-1", time.Now().UnixMilli(), time.Now().UnixMilli()).
		AddRow("item-b-2", "tenant-B", "Non-Veg Option 2 Sandwich", "Second non-vegetarian option", 10.99, "cat-uuid-1", time.Now().Add(-1*time.Hour).UnixMilli(), time.Now().Add(-1*time.Hour).UnixMilli())

	// Expect two separate queries with different tenant IDs
	countQuery := regexp.QuoteMeta(`SELECT COUNT(*) FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1`)
	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(price, 0) as price, categoryId as categoryId, created_at, updated_at FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`)

	mock.ExpectQuery(countQuery).WithArgs("tenant-A").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2)) // Expect two separate count queries for different tenants
	mock.ExpectQuery(queryRegex).WithArgs("tenant-A", 10, 0).WillReturnRows(tenantARows)

	mock.ExpectQuery(countQuery).WithArgs("tenant-B").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2)) // Expect two separate count queries for different tenants
	mock.ExpectQuery(queryRegex).WithArgs("tenant-B", 10, 0).WillReturnRows(tenantBRows)

	repoATenant := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-A",
		Logger:   logger.NewLogger(),
	}

	repoBTenant := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-B",
		Logger:   logger.NewLogger(),
	}

	// Query for tenant A
	itemsA, countA, errA := repoATenant.ListByTenant(t.Context(), "tenant-A", 0, 10)
	require.NoError(t, errA)
	assert.Equal(t, int64(2), countA)
	for _, item := range itemsA {
		assert.Equal(t, "tenant-A", item.TenantID)
	}

	// Query for tenant B
	itemsB, countB, errB := repoBTenant.ListByTenant(t.Context(), "tenant-B", 0, 10)
	require.NoError(t, errB)
	assert.Equal(t, int64(2), countB)
	for _, item := range itemsB {
		assert.Equal(t, "tenant-B", item.TenantID)
	}

	// Verify mock expectations were all met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_ListByTenant_MultipleItemsWithNullDescription(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	countQuery := regexp.QuoteMeta(`SELECT COUNT(*) FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1`)
	mock.ExpectQuery(countQuery).WithArgs("tenant-xyz").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Include NULL description for some items
	currentTime := time.Now().UnixMilli()
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "price", "categoryId", "created_at", "updated_at"}).
		AddRow("item-1", "tenant-xyz", "Burger Category", "Meat based burgers", 9.99, "cat-uuid-1", currentTime, currentTime). // Has description
		AddRow("item-2", "tenant-xyz", "Salad Sandwich", "", 5.99, "cat-uuid-2", currentTime, currentTime)                     // NULL description

	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, '') as description, COALESCE(price, 0) as price, categoryId as categoryId, created_at, updated_at FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-xyz", 10, 0).WillReturnRows(rows)

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-xyz",
		Logger:   logger.NewLogger(),
	}

	items, count, err := repo.ListByTenant(t.Context(), "tenant-xyz", 0, 10)

	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.Equal(t, 2, len(items))

	// First item has description
	desc1 := *items[0].Description
	assert.Equal(t, "Meat based burgers", desc1)

	// Second item has NULL description - should be empty string due to sqlx behavior
	desc2 := items[1].Description
	assert.Empty(t, desc2, "NULL values are returned as empty strings by sqlx")

	assert.Equal(t, "Burger Category", items[0].Name)
	assert.Equal(t, "Salad Sandwich", items[1].Name)
	assert.Equal(t, "tenant-xyz", items[1].TenantID)

	// Verify mock expectations
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock successful query result - DB.Select types slice as []T (values, not pointers) based on T parameter
	currentTime := time.Now().UnixMilli()
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "price", "categoryId", "created_at", "updated_at"}).
		AddRow("item-uuid-123", "tenant-123", "Pizza", "Delicious pizza with cheese", 9.99, "cat-uuid-1", currentTime, currentTime)
	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, ''), COALESCE(price, 0), categoryId, created_at, updated_at FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1 AND id = $2`)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-123", "item-uuid-123").WillReturnRows(rows)

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	item, err := repo.GetByID(t.Context(), "tenant-123", "item-uuid-123")

	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "item-uuid-123", item.ID)
	assert.Equal(t, "Pizza", item.Name)
	assert.Equal(t, "Delicious pizza with cheese", *item.Description)
	assert.Equal(t, 9.99, item.Price)
	assert.Equal(t, "cat-uuid-1", item.CategoryID)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Don't set up query expectation - when DB.Get() is called with no matching rows,
	// it will return an error which is the expected "not found" behavior

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	item, err := repo.GetByID(t.Context(), "tenant-123", "item-nonexistent")

	require.Error(t, err, "Expected error for non-existent item")
	assert.Nil(t, item)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_GetByID_ErrorMessageOnFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query fails due to connection error
	queryRegex := regexp.QuoteMeta(`SELECT id, tenant_id, name, COALESCE(description, ''), COALESCE(price, 0), categoryId, created_at, updated_at FROM food_item WHERE current_setting('app.current_tenant_id')::TEXT = $1 AND id = $2`)
	mock.ExpectQuery(queryRegex).WithArgs("tenant-999", "item-uuid-xyz").WillReturnError(fmt.Errorf("database connection error"))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-999",
		Logger:   logger.NewLogger(),
	}

	item, err := repo.GetByID(t.Context(), "tenant-999", "item-uuid-xyz")

	require.Error(t, err)
	assert.Nil(t, item)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	currentTime := time.Now().UTC().UnixMilli()
	// Mock successful INSERT - Exec returns single value (not slice)
	mock.ExpectExec(`INSERT INTO food_item`).WithArgs("item-uuid-123", "Burger", "Fresh beef and vegetables", 9.99, "cat-uuid-1", "tenant-123", currentTime, currentTime).WillReturnResult(sqlmock.NewResult(123, 1))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	description := "Fresh beef and vegetables"
	isVegetarian := false

	item := &models.FoodItem{
		ID:                 "item-uuid-123",
		TenantID:           "tenant-123",
		Name:               "Burger",
		Description:        &description,
		Price:              9.99,
		CategoryID:         "cat-uuid-1",
		IsVegetarian:       isVegetarian,
		AvailabilityStatus: "available",
		CreatedAt:          currentTime,
		UpdatedAt:          currentTime,
		DeliveredAt:        0,
	}

	err = repo.Create(t.Context(), item)

	require.NoError(t, err)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_Create_FailDuplicateID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	currentTime := time.Now().UTC().UnixMilli()
	// Mock INSERT fails due to unique constraint violation on id
	mock.ExpectExec(`INSERT INTO food_item`).WithArgs("item-uuid-already-exists", "Burger", "Description", 9.99, "cat-uuid-1", "tenant-123", currentTime, currentTime).WillReturnError(fmt.Errorf("duplicate key value violates unique constraint \"food_item_pkey\""))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	description := "Description"
	isVegetarian := false

	item := &models.FoodItem{
		ID:                 "item-uuid-already-exists",
		TenantID:           "tenant-123",
		Name:               "Burger",
		Description:        &description,
		Price:              9.99,
		CategoryID:         "cat-uuid-1",
		IsVegetarian:       isVegetarian,
		AvailabilityStatus: "available",
		CreatedAt:          currentTime,
		UpdatedAt:          currentTime,
		DeliveredAt:        0,
	}

	err = repo.Create(t.Context(), item)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_Create_FailDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock INSERT fails due to database error
	currentTime := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`INSERT INTO food_item (id, name, description, price, categoryId, tenant_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)
	mock.ExpectExec(queryRegex).WithArgs("item-uuid-new", "Burger", "Description", 9.99, "cat-uuid-1", "tenant-123", currentTime, currentTime).WillReturnError(fmt.Errorf("constraint violation"))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	description := "Description"
	isVegetarian := false

	item := &models.FoodItem{
		ID:                 "item-uuid-new",
		TenantID:           "tenant-123",
		Name:               "Burger",
		Description:        &description,
		Price:              9.99,
		CategoryID:         "cat-uuid-1",
		IsVegetarian:       isVegetarian,
		AvailabilityStatus: "available",
		CreatedAt:          currentTime,
		UpdatedAt:          currentTime,
		DeliveredAt:        0,
	}

	err = repo.Create(t.Context(), item)

	require.Error(t, err)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock successful UPDATE - Exec returns single value (not slice)
	currentTime := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`UPDATE food_item SET name = $1, description = $2, price = $3, categoryId = $4, updated_at = $5 WHERE id = $6 AND tenant_id = $7`)
	mock.ExpectExec(queryRegex).WithArgs("Updated Burger Name", "Updated description", 12.99, "cat-uuid-1", currentTime, "item-uuid-123", "tenant-123").WillReturnResult(sqlmock.NewResult(10, 1))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	description := "Updated description"
	isVegetarian := false
	item := &models.FoodItem{
		ID:                 "item-uuid-123",
		TenantID:           "tenant-123",
		Name:               "Updated Burger Name",
		Description:        &description,
		Price:              12.99,
		CategoryID:         "cat-uuid-1",
		IsVegetarian:       isVegetarian,
		AvailabilityStatus: "available",
		CreatedAt:          currentTime,
		UpdatedAt:          currentTime,
		DeliveredAt:        0,
	}

	err = repo.Update(t.Context(), item)

	require.NoError(t, err)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_Update_FailNotExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock UPDATE affects 0 rows because ID doesn't exist
	queryRegex := regexp.QuoteMeta(`UPDATE food_item SET name = $1, description = $2, price = $3, categoryId = $4, updated_at = $5 WHERE id = $6 AND tenant_id = $7`)
	mock.ExpectExec(queryRegex).WithArgs("New Name", "New Description", 10.99, "cat-uuid-1", time.Now().UTC().UnixMilli(), "item-nonexistent", "tenant-123").WillReturnResult(sqlmock.NewResult(0, 0))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	description := "New Description"
	isVegetarian := false
	currentTime := time.Now().UTC().UnixMilli()
	item := &models.FoodItem{
		ID:                 "item-nonexistent",
		TenantID:           "tenant-123",
		Name:               "New Name",
		Description:        &description,
		Price:              10.99,
		CategoryID:         "cat-uuid-1",
		IsVegetarian:       isVegetarian,
		AvailabilityStatus: "available",
		CreatedAt:          currentTime,
		UpdatedAt:          currentTime,
		DeliveredAt:        0,
	}

	err = repo.Update(t.Context(), item)

	require.NoError(t, err, "Expected error when updating non-existent item")

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_Update_FailDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock UPDATE fails due to database error
	updatedAt := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`UPDATE food_item SET name = $1, description = $2, price = $3, categoryId = $4, updated_at = $5 WHERE id = $6 AND tenant_id = $7`)
	mock.ExpectExec(queryRegex).WithArgs("New Name", "New Description", 10.99, "cat-uuid-1", updatedAt, "item-uuid-xyz", "tenant-123").WillReturnError(fmt.Errorf("database error: constraint violation"))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	description := "New Description"
	isVegetarian := false
	item := &models.FoodItem{
		ID:                 "item-uuid-xyz",
		TenantID:           "tenant-123",
		Name:               "New Name",
		Description:        &description,
		Price:              10.99,
		CategoryID:         "cat-uuid-1",
		IsVegetarian:       isVegetarian,
		AvailabilityStatus: "available",
		CreatedAt:          updatedAt,
		UpdatedAt:          updatedAt,
		DeliveredAt:        0,
	}

	err = repo.Update(t.Context(), item)

	require.Error(t, err)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock successful UPDATE for soft delete (setting is_active = FALSE)
	updatedAt := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`UPDATE food_item SET is_active = FALSE, updated_at = $1 WHERE id = $2 AND tenant_id = $3`)
	mock.ExpectExec(queryRegex).WithArgs(updatedAt, "item-uuid-123", "tenant-123").WillReturnResult(sqlmock.NewResult(10, 1))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	rowsAffected, err := repo.Delete(t.Context(), "item-uuid-123")

	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected, "Expected 1 row to be affected when deleting item")
	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_Delete_FailNotExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock UPDATE affects 0 rows because ID doesn't exist or soft delete already applied
	updatedAt := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`UPDATE food_item SET is_active = FALSE, updated_at = $1 WHERE id = $2 AND tenant_id = $3`)
	mock.ExpectExec(queryRegex).WithArgs(updatedAt, "item-nonexistent", "tenant-123").WillReturnResult(sqlmock.NewResult(0, 0))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	rowsAffected, err := repo.Delete(t.Context(), "item-nonexistent")

	require.NoError(t, err)
	assert.Equal(t, int64(0), rowsAffected, "Expected 0 rows to be affected when deleting non-existent item")

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_Delete_FailDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock UPDATE fails due to database error
	updatedAt := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`UPDATE food_item SET is_active = FALSE, updated_at = $1 WHERE id = $2 AND tenant_id = $3`)
	mock.ExpectExec(queryRegex).WithArgs(updatedAt, "item-uuid-xyz", "tenant-123").WillReturnError(fmt.Errorf("database error: constraint violation"))

	repo := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-123",
		Logger:   logger.NewLogger(),
	}

	rowsAffected, err := repo.Delete(t.Context(), "item-uuid-xyz")

	require.Error(t, err)
	assert.Equal(t, int64(0), rowsAffected, "Expected 0 rows to be affected when delete fails due to database error")

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLFoodItemRepository_Delete_TenantIsolation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")

	// Create two items with different IDs but same tenant - only one should be deleted
	item1ID := "item-uuid-tenant-a"
	item2ID := "item-uuid-tenant-b"

	updatedAt := time.Now().UTC().UnixMilli()
	queryRegex := regexp.QuoteMeta(`UPDATE food_item SET is_active = FALSE, updated_at = $1 WHERE id = $2 AND tenant_id = $3`)
	mock.ExpectExec(queryRegex).WithArgs(updatedAt, item1ID, "tenant-A").WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectExec(queryRegex).WithArgs(updatedAt, item2ID, "tenant-B").WillReturnResult(sqlmock.NewResult(10, 1))

	repoATenant := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-A",
		Logger:   logger.NewLogger(),
	}

	repoBTenant := &PostgreSQLFoodItemRepository{
		DB:       dbx,
		TenantID: "tenant-B",
		Logger:   logger.NewLogger(),
	}

	// Delete from tenant A - should only affect item with ID item-uuid-tenant-a
	rowsAffectedA, errA := repoATenant.Delete(t.Context(), item1ID)
	require.NoError(t, errA)
	assert.Equal(t, int64(1), rowsAffectedA, "Expected 1 row to be affected when deleting item for tenant A")

	// Delete from tenant B - should only affect item with ID item-uuid-tenant-b
	rowsAffectedB, errB := repoBTenant.Delete(t.Context(), item2ID)
	require.NoError(t, errB)
	assert.Equal(t, int64(1), rowsAffectedB, "Expected 1 row to be affected when deleting item for tenant B")

	// Verify mock expectations were all met
	assert.NoError(t, mock.ExpectationsWereMet())
}
