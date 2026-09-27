package fooditems

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	"github.com/pareshvernekar/homecooked/internal/services/fooditem"
	"github.com/pareshvernekar/homecooked/internal/views"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
	testhttp "github.com/pareshvernekar/homecooked/tests/http"
	"github.com/stretchr/testify/require"
)

// ListFoodItemsTest tests the Food Item listing API endpoint with pagination
func ListFoodItemsTest(t *testing.T) {
	db, err := testdb.NewDatabaseHelper(t.Context())
	if err != nil {
		t.Fatalf("Failed to create database connection: %v", err)
	}

	defer func() {
		err := db.Terminate(t.Context())
		require.NoError(t, err, "Failed to terminate database connection")
	}()

	err = InitializeSchema(t.Context(), db.DB)
	require.NoError(t, err, "Failed to initialize schema")

	categoryID, err := SetupFoodCategory(t.Context(), db.DB)
	if err != nil {
		t.Fatalf("Failed to setup food category: %v", err)
	}

	item1Id, err := InsertFoodItem(t.Context(), db.DB, "Chicken Biryani", "Rich biryani with spices", 249.99, categoryID, uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to insert food item: %v", err)
	}

	item2Id, err := InsertFoodItem(t.Context(), db.DB, "Paneer Tikka", "Grilled cheese cubes", 199.00, categoryID, uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to insert food item: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/food-items?page=1&limit=10", strings.NewReader(""))

	ctx, resp := testhttp.CreateTestContext(req)

	logger := logger.NewLogger()
	foodCategoryRepository := repository.NewFoodCategoryRepository(db.DB, logger, "test-tenant")
	foodCategoryService := foodcategory.NewFoodCategoryService(foodCategoryRepository, logger)
	foodItemRepository := repository.NewFoodItemRepository(db.DB, logger, "test-tenant")
	foodItemService := fooditem.NewFoodItemService(foodItemRepository, logger, foodCategoryService)
	handler := handlers.NewFoodItemHandler(foodItemService, logger)
	handler.GetFoodItems(ctx)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200 but got %d. Body: %s", resp.Code, resp.Body.String())
	}

	var response views.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Errorf("Response should be successful")
	}

	// Loop through data to verify both items exist by their IDs
	item1Found := false
	item2Found := false

	// The response.Data contains []interface{} (each item is a FoodItem map)
	for _, item := range response.Data.([]interface{}) {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract ID and Name from the item
		idValue, idOk := itemMap["id"].(string)
		nameValue, _ := itemMap["name"].(string)

		// Check if this matches either item
		if idOk && (idValue == item1Id || idValue == item2Id) {
			if idValue == item1Id {
				item1Found = true
				t.Logf("✓ Found Item 1 (%s) with ID: %s", nameValue, idValue)
			}
			if idValue == item2Id {
				item2Found = true
				t.Logf("✓ Found Item 2 (%s) with ID: %s", nameValue, idValue)
			}
		}
	}

	// Assert both items were found
	if !item1Found {
		t.Errorf("Chicken Biryani (item1) was not found in the response. Expected ID: %s, Got %d items: ", item1Id, len(response.Data.([]interface{})))
		for _, item := range response.Data.([]interface{}) {
			itemMap := item.(map[string]interface{})
			t.Logf("   - Found: id=%v, name=%v", itemMap["id"], itemMap["name"])
		}
	}

	if !item2Found {
		t.Errorf("Paneer Tikka (item2) was not found in the response. Expected ID: %s, Got %d items: ", item2Id, len(response.Data.([]interface{})))
		for _, item := range response.Data.([]interface{}) {
			itemMap := item.(map[string]interface{})
			t.Logf("   - Found: id=%v, name=%v", itemMap["id"], itemMap["name"])
		}
	}
	t.Logf("List response - Success: %v, Data: %+v", response.Success, response.Data)

}

// ListFoodItemsWithPaginationTest tests pagination functionality
func TestListFoodItemsWithPagination(t *testing.T) {
	db, err := testdb.NewDatabaseHelper(t.Context())
	if err != nil {
		t.Fatalf("Failed to create database connection: %v", err)
	}

	defer func() {
		err := db.Terminate(t.Context())
		require.NoError(t, err, "Failed to terminate database connection")
	}()

	err = InitializeSchema(t.Context(), db.DB)
	require.NoError(t, err, "Failed to initialize schema")

	categoryID, _ := SetupFoodCategory(t.Context(), db.DB)

	item1Id, err := InsertFoodItem(t.Context(), db.DB, "Dish 1", "Description 1", 100.00, categoryID, "test-tenant")
	if err != nil {
		t.Fatalf("Failed to insert item: %v", err)
	}

	item2Id, err := InsertFoodItem(t.Context(), db.DB, "Dish 2", "Description 2", 150.00, categoryID, "test-tenant")
	if err != nil {
		t.Fatalf("Failed to insert item: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/food-items?page=1&limit=5", strings.NewReader(""))

	ctx, resp := testhttp.CreateTestContext(req)

	logger := logger.NewLogger()
	foodCategoryRepository := repository.NewFoodCategoryRepository(db.DB, logger, "test-tenant")
	foodCategoryService := foodcategory.NewFoodCategoryService(foodCategoryRepository, logger)
	foodItemRepository := repository.NewFoodItemRepository(db.DB, logger, "test-tenant")
	foodItemService := fooditem.NewFoodItemService(foodItemRepository, logger, foodCategoryService)
	handler := handlers.NewFoodItemHandler(foodItemService, logger)
	handler.GetFoodItems(ctx)

	var response views.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
		return
	}

	t.Logf("Received response with Success: %v, Data type: %T", response.Success, response.Data)

	// Store the validated food items for further assertion
	var validatedItems []interface{}
	// Loop through data and collect all items - they should match our inserted item IDs
	for _, item := range response.Data.([]interface{}) {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		idValue, _ := itemMap["id"].(string)
		nameValue, _ := itemMap["name"].(string)

		t.Logf("Found item: id=%v, name=%s", idValue, nameValue)

		// Collect the items
		validatedItems = append(validatedItems, item)

	}

	item1Set := false

	for _, item := range validatedItems {
		itemMap := item.(map[string]interface{})
		idValue, _ := itemMap["id"].(string)

		if idValue == item1Id {
			item1Set = true
		}
	}

	if !item1Set {
		t.Errorf("Dish 1 was not found in the response")
	} else {
		t.Logf("✓ Found Dish 1 in paginated list")
	}
	item2Set := false

	for _, item := range validatedItems {
		itemMap := item.(map[string]interface{})
		idValue, _ := itemMap["id"].(string)

		if idValue == item2Id {
			item2Set = true
		}
	}

	if !item2Set {
		t.Errorf("Dish 2 was not found in the response")
	} else {
		t.Logf("✓ Found Dish 2 in paginated list")
	}

	t.Log("✓ Verification passed - both items found in paginated response")
}
