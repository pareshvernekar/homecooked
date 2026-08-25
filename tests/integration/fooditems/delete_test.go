package fooditems

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
	testhttp "github.com/pareshvernekar/homecooked/tests/http"
	"github.com/stretchr/testify/assert"
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

	err = InsertFoodItem(t.Context(), db.DB, "Chicken Biryani", "Rich biryani", 249.99, categoryID, "test-tenant")
	if err != nil {
		t.Fatalf("Failed to insert item: %v", err)
	}

	itemUUID := "550e8400-e29b-41d4-a716-446655440000" // Sample UUID
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/food-items/"+itemUUID, strings.NewReader(""))

	ctx, resp := testhttp.CreateTestContext(req)

	logger := logger.NewLogger()
	handler := handlers.NewFoodItemHandler(nil, logger)
	handler.DeleteFoodItem(ctx)

	if resp.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 but got %d. Body: %s", resp.Code, resp.Body.String())
	}

	var response testhttp.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	assert.Equal(t, true, response.Success, "Delete should return success")
	t.Logf("Delete test passed - Item deleted successfully")
}

// DeleteNonExistentFoodItemTest tests deletion of non-existent food item
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

	logger := logger.NewLogger()
	handler := handlers.NewFoodItemHandler(nil, logger)
	handler.DeleteFoodItem(ctx)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 but got %d. Body: %s", resp.Code, resp.Body.String())
	}

	var response testhttp.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	assert.Equal(t, true, response.Success, "Delete should return success")
	t.Logf("Non-existent delete test passed - 404 returned as expected")
}
