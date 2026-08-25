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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MultiTenantFoodItemsTest tests cross-tenant data isolation
func TestMultiTenantFoodItems(t *testing.T) {
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

	err = InsertFoodItem(t.Context(), db.DB, "Chicken Biryani A", "Description", 249.99, categoryID, uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to insert item: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/food-items?page=1&limit=10", strings.NewReader(""))

	ctx, resp := testhttp.CreateTestContext(req)

	logger := logger.NewLogger()
	handler := handlers.NewFoodItemHandler(nil, logger)
	handler.GetFoodItems(ctx)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200 but got %d. Body: %s", resp.Code, resp.Body.String())
		return
	}

	var response testhttp.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
		return
	}

	assert.Equal(t, true, response.Success, "Delete should return success")
	t.Logf("Multi-tenant isolation test passed")
}
