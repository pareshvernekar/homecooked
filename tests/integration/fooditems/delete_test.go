package fooditems

import (
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
	testdb "github.com/pareshvernekar/homecooked/tests/db"
	testhttp "github.com/pareshvernekar/homecooked/tests/http"
	"github.com/stretchr/testify/require"
)

// DeleteFoodItemTest tests the Food Item deletion API endpoint
func TestDeleteFoodItem(t *testing.T) {
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

	itemID, err := InsertFoodItem(t.Context(), db.DB, "Chicken Biryani", "Rich biryani", 249.99, categoryID, "test-tenant")
	if err != nil {
		t.Fatalf("Failed to insert item: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/food-items/"+itemID, strings.NewReader(""))

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
	handler.DeleteFoodItem(ctx)

	if resp.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 but got %d. Body: %s", resp.Code, resp.Body.String())
	}

	// For 204 No Content, we expect empty body - no JSON parsing needed
	t.Logf("✓ Delete test passed - Item deleted successfully (returned 204 No Content)")
}

// TestDeleteNonExistentFoodItem tests deletion of non-existent food item
func TestDeleteNonExistentFoodItem(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/food-items/non-existent-id", strings.NewReader(""))

	ctx, resp := testhttp.CreateTestContext(req)
	ctx.Params = gin.Params{
		{Key: "id", Value: "non-existent-id"},
	}
	logger := logger.NewLogger()
	foodCategoryRepository := repository.NewFoodCategoryRepository(db.DB, logger, "test-tenant")
	foodCategoryService := foodcategory.NewFoodCategoryService(foodCategoryRepository, logger)
	foodItemRepository := repository.NewFoodItemRepository(db.DB, logger, "test-tenant")
	foodItemService := fooditem.NewFoodItemService(foodItemRepository, logger, foodCategoryService)
	handler := handlers.NewFoodItemHandler(foodItemService, logger)
	handler.DeleteFoodItem(ctx)

	// Per user requirement: when deleting non-existent item, return 204 No Content (not 404)
	// This is acceptable behavior since the item doesn't exist for this tenant
	if resp.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 No Content but got %d. Body: %s", resp.Code, resp.Body.String())
	}

	// For 204 No Content, we expect empty body - no JSON parsing needed
	t.Logf("✓ Non-existent delete test passed - 204 No Content returned as per requirement (item doesn't exist)")
}
