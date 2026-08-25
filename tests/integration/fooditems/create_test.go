package fooditems

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	models "github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	"github.com/pareshvernekar/homecooked/internal/services/fooditem"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
	testhttp "github.com/pareshvernekar/homecooked/tests/http"
	"github.com/stretchr/testify/require"
)

// CreateFoodItemTest tests the Food Item creation API endpoint
func TestCreateFoodItem(t *testing.T) {
	db, err := testdb.NewDatabaseHelper(t.Context())
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL connection: %v", err)
	}

	defer func() {
		err := db.Terminate(t.Context())
		require.NoError(t, err, "Failed to terminate database connection")
	}()

	// Initialize database schema
	err = testdb.InitializeSchema(t.Context(), db.DB)
	if err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	// Create test tenant
	testTenant, err := testdb.CreateTestTenant(t.Context(), db.DB)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	defer func() { _ = testTenant.Cleanup() }()

	// Test case 1: Valid food item creation request with valid category name
	body := map[string]interface{}{
		"name":                "Chicken Biryani",
		"description":         "Rich and flavorful biryani with aromatic spices",
		"price":               249.99,
		"category_name":       "vegetarian",
		"availability_status": "available",
	}

	var bodyBytes bytes.Buffer
	if err := json.NewEncoder(&bodyBytes).Encode(body); err != nil {
		t.Fatal("Failed to encode JSON body")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/food-items", strings.NewReader(bodyBytes.String()))

	ctx, resp := testhttp.CreateTestContext(req)

	logger := logger.NewLogger()
	foodCategoryRepository := repository.NewFoodCategoryRepository(db.DB, logger, "test-tenant")
	foodCategoryService := foodcategory.NewFoodCategoryService(foodCategoryRepository, logger)
	foodItemRepository := repository.NewFoodItemRepository(db.DB, logger, "test-tenant")
	foodItemService := fooditem.NewFoodItemService(foodItemRepository, logger, foodCategoryService)

	handler := handlers.NewFoodItemHandler(foodItemService, logger)
	handler.CreateFoodItem(ctx)

	if resp.Code != http.StatusCreated {
		t.Errorf("Expected status 201 but got %d. Body: %s", resp.Code, resp.Body.String())
	}

	// Parse response
	var response testhttp.SuccessResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response")
	}

	if !response.Success {
		t.Errorf("Response should be successful")
	}

	data := resp.Body.Bytes()

	t.Logf("Create test passed - Created food item successfully. Response body: %s", string(data))

	// Verify database state - insert into the category table
	var category models.FoodCategory
	query := `SELECT id, tenant_id, name, description, is_active, created_at, updated_at FROM food_category WHERE id = $1`
	err = db.DB.Get(&category, query, testTenant.ID)
	if err != nil {
		t.Fatalf("Failed to verify database state: %v", err)
	}

	t.Logf("Database state - Food category ID: %s, Name: %s", testTenant.ID, category.Name)
}
