package fooditems

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	"github.com/pareshvernekar/homecooked/internal/services/fooditem"
	"github.com/pareshvernekar/homecooked/internal/views"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
	testhttp "github.com/pareshvernekar/homecooked/tests/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateFoodItem tests the Food Item update API endpoint with partial data
func TestUpdateFoodItem(t *testing.T) {
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

	itemID, err := InsertFoodItem(t.Context(), db.DB, "Chicken Biryani (Original)", "Rich biryani with spices", 249.99, categoryID, "test-tenant")
	if err != nil {
		t.Fatalf("Failed to insert food item: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/food-items/"+itemID, strings.NewReader(""))

	body := map[string]interface{}{
		"price":               349.99,
		"availability_status": "available",
	}

	bodyBytes, _ := json.Marshal(body)
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	ctx, resp := testhttp.CreateTestContext(req)
	ctx.Params = gin.Params{
		{Key: "id", Value: itemID},
	}
	logger := logger.NewLogger()
	foodCategoryRepository := repository.NewFoodCategoryRepository(db.DB, logger, "test-tenant")
	foodCategoryService := foodcategory.NewFoodCategoryService(foodCategoryRepository, logger)
	foodItemRepository := repository.NewFoodItemRepository(db.DB, logger, "test-tenant")
	foodItemService := fooditem.NewFoodItemService(foodItemRepository, logger, foodCategoryService)
	handler := handlers.NewFoodItemHandler(foodItemService, logger)
	handler.UpdateFoodItem(ctx)

	require.Equal(t, http.StatusOK, resp.Code, "Expected status 200. Body: %s", resp.Body.String())

	var response views.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err, "Failed to parse response")
	assert.True(t, response.Success, "Response should be successful")
	assert.Equal(t, "Food item updated successfully", response.Message)

	// Verify the price was persisted in the database
	var updatedPrice float64
	err = db.DB.Get(&updatedPrice, `SELECT price FROM food_item WHERE id = $1 AND tenant_id = $2`, itemID, "test-tenant")
	require.NoError(t, err, "Failed to query updated food item")
	assert.Equal(t, 349.99, updatedPrice, "Price should be updated to 349.99")
	t.Logf("Update test passed - Updated price to: %v", updatedPrice)
}

// TestUpdateFoodItemPartial tests partial update functionality (update only some fields)
func TestUpdateFoodItemPartial(t *testing.T) {
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
	require.NoError(t, err, "Failed to setup food category")

	itemID, err := InsertFoodItem(t.Context(), db.DB, "Paneer Tikka", "Grilled paneer", 199.00, categoryID, "test-tenant")
	if err != nil {
		t.Fatalf("Failed to insert item: %v", err)
	}

	body := map[string]interface{}{
		"availability_status": "low_stock",
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/food-items/"+itemID, strings.NewReader(string(bodyBytes)))

	ctx, resp := testhttp.CreateTestContext(req)
	ctx.Params = gin.Params{
		{Key: "id", Value: itemID},
	}
	logger := logger.NewLogger()
	foodCategoryRepository := repository.NewFoodCategoryRepository(db.DB, logger, "test-tenant")
	foodCategoryService := foodcategory.NewFoodCategoryService(foodCategoryRepository, logger)
	foodItemRepository := repository.NewFoodItemRepository(db.DB, logger, "test-tenant")
	foodItemService := fooditem.NewFoodItemService(foodItemRepository, logger, foodCategoryService)
	handler := handlers.NewFoodItemHandler(foodItemService, logger)
	handler.UpdateFoodItem(ctx)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200 for partial update but got %d", resp.Code)
	}

	var response testhttp.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	t.Logf("Partial update test passed")
}
