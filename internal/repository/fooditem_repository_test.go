package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pareshvernekar/homecooked/internal/models"
)

var testDB *sqlx.DB // Global variable shared across tests in this package

func TestMain(m *testing.M) {
	// 3. Run all tests in the package
	testDB = createDB() // Initialize the in-memory database once for all tests
	exitCode := m.Run()
	// 4. Teardown: Close connection or clean up
	err := testDB.Close()
	if err != nil {
		panic(err)
	}
	os.Exit(exitCode)
}
func createDB() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", "file=:memory:?_busy_timeout=2000&_cache_lru_size=100")
	if err != nil {
		panic(err)
	}

	// 2. Wipe the entire database schema completely
	// This drops all tables, views, triggers, and indexes instantly
	_, err = db.Exec("PRAGMA writable_schema = 1; DELETE FROM sqlite_schema; PRAGMA writable_schema = 0; VACUUM;")
	if err != nil {
		panic(err)
	}

	result, err := db.Exec(`CREATE TABLE IF NOT EXISTS food_item (id TEXT PRIMARY KEY, name TEXT, description TEXT, price DECIMAL(10,2), category TEXT, tenant_id VARCHAR(50), created_at TIMESTAMP, updated_at TIMESTAMP)`)
	if err != nil {
		panic(err)
	}
	if result == nil {
		panic("Failed to create food_item table")
	}

	return db
}

func TestCreate_FoodItemSuccess(t *testing.T) {
	desc := "A delicious burger"
	foodItem := models.FoodItem{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Name:        "Test Burger",
		Description: &desc,
		Price:       12.99,
		Category:    "main_course",
		TenantID:    "1",
	}

	repo := NewFoodItemRepository(testDB, "1")

	err := repo.Create(&foodItem)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUpdate_FoodItemSuccess(t *testing.T) {
	desc := "An original burger"
	existingFoodItem := models.FoodItem{
		ID:          "550e8400-e29b-41d4-a716-446655440001",
		Name:        "Original Burger",
		Description: &desc,
		Price:       12.99,
		Category:    "main_course",
		TenantID:    "1",
	}

	repo := NewFoodItemRepository(testDB, "1")

	err := repo.Update(&existingFoodItem)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestDelete_FoodItemSuccess(t *testing.T) {
	foodItemId := "550e8400-e29b-41d4-a716-446655440003"

	repo := NewFoodItemRepository(testDB, "1")

	err := repo.Delete(foodItemId)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestListByTenant_FoodItemsSuccess(t *testing.T) {

	repo := NewFoodItemRepository(testDB, "1")

	result, err := testDB.Exec(`INSERT INTO food_item (id, name, price) VALUES ('uuid-1', 'Burger 1', 10.99)`)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("Failed to insert food item")
	}

	_, _, _ = repo.ListByTenant("1", 20, 0)
}

func TestGetByID_FoodItemNotFound(t *testing.T) {

	repo := NewFoodItemRepository(testDB, "1")

	foodItem, err := repo.GetByID("non-existent-id")
	if err == nil {
		t.Error("Expected error for not found, got nil")
	}
	if foodItem != nil {
		t.Errorf("Expected nil food item, got: %v", foodItem)
	}
}
