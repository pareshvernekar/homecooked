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

	err = InsertFoodItem(t.Context(), db.DB, "Chicken Biryani", "Rich biryani with spices", 249.99, categoryID, uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to insert food item: %v", err)
	}

	err = InsertFoodItem(t.Context(), db.DB, "Paneer Tikka", "Grilled cheese cubes", 199.00, categoryID, uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to insert food item: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/food-items?page=1&limit=10", strings.NewReader(""))

	ctx, resp := testhttp.CreateTestContext(req)

	logger := logger.NewLogger()
	handler := handlers.NewFoodItemHandler(nil, logger)
	handler.GetFoodItems(ctx)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200 but got %d. Body: %s", resp.Code, resp.Body.String())
	}

	var response testhttp.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Errorf("Response should be successful")
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

	err = InsertFoodItem(t.Context(), db.DB, "Dish 1", "Description 1", 100.00, categoryID, uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to insert item: %v", err)
	}

	err = InsertFoodItem(t.Context(), db.DB, "Dish 2", "Description 2", 150.00, categoryID, uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to insert item: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/food-items?page=1&limit=5", strings.NewReader(""))

	ctx, resp := testhttp.CreateTestContext(req)

	logger := logger.NewLogger()
	handler := handlers.NewFoodItemHandler(nil, logger)
	handler.GetFoodItems(ctx)

	var response testhttp.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
		return
	}

	paginatedData := response.Data.(map[string]interface{})
	totalItems := paginatedData["total"].(float64)
	countItems := int(paginatedData["count"].(int64))

	if totalItems != 2 {
		t.Errorf("Total should be 2, got %.0f", totalItems)
		return
	}
	if countItems != 2 {
		t.Errorf("Count should be 2, got %d", countItems)
		return
	}

	t.Logf("Pagination test passed - Total: %d, Count: %d", int(totalItems), countItems)
}
