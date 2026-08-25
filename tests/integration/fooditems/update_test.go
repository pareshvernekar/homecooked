package fooditems

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
	testhttp "github.com/pareshvernekar/homecooked/tests/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UpdateFoodItemTest tests the Food Item update API endpoint with partial data
func UpdateFoodItemTest(t *testing.T) {
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

	err = InsertFoodItem(t.Context(), db.DB, "Chicken Biryani (Original)", "Rich biryani with spices", 249.99, categoryID, uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to insert food item: %v", err)
	}

	itemUUID := "550e8400-e29b-41d4-a716-446655440000" // Sample UUID
	req := httptest.NewRequest(http.MethodPut, "/api/v1/food-items/"+itemUUID, strings.NewReader(""))

	body := map[string]interface{}{
		"price":               349.99,
		"availability_status": "available",
	}

	bodyBytes, _ := json.Marshal(body)
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	ctx, resp := testhttp.CreateTestContext(req)

	logger := logger.NewLogger()
	handler := handlers.NewFoodItemHandler(nil, logger)
	handler.UpdateFoodItem(ctx)

	if resp.Code != http.StatusAccepted {
		t.Errorf("Expected status 202 but got %d. Body: %s", resp.Code, resp.Body.String())
	}

	var response testhttp.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	paginatedData := response.Data.(map[string]interface{})
	totalItems := paginatedData["total"].(float64)
	countItems := int(paginatedData["count"].(int64))

	assert.Equal(t, float64(349.99), totalItems, "Price should be 349.99")
	t.Logf("Update test passed - Updated price to: %v", countItems)
}

// UpdateFoodItemPartialTest tests partial update functionality (update only some fields)
func UpdateFoodItemPartialTest(t *testing.T) {
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

	categoryID := uuid.New().String()
	err = InsertFoodItem(t.Context(), db.DB, "Paneer Tikka", "Grilled paneer", 199.00, categoryID, uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to insert item: %v", err)
	}

	body := map[string]interface{}{
		"availability_status": "low_stock",
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/food-items/:id", strings.NewReader(string(bodyBytes)))

	ctx, resp := testhttp.CreateTestContext(req)

	logger := logger.NewLogger()
	handler := handlers.NewFoodItemHandler(nil, logger)
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
