package foodcategories

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/handlers"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	"github.com/pareshvernekar/homecooked/internal/views"
	"github.com/stretchr/testify/require"

	"github.com/gin-gonic/gin"

	testdb "github.com/pareshvernekar/homecooked/tests/db"
)

// TestCreateFoodCategory tests the Food Category creation API endpoint
func TestCreateFoodCategory(t *testing.T) {
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
	defer func() {
		err := testTenant.Cleanup()
		require.NoError(t, err, "Failed to cleanup test tenant")
	}()

	// Test case 1: Valid food category creation request
	body := map[string]interface{}{
		"name":        "test-vegetarian",
		"description": "Vegetarian meals",
		"tenant_id":   testTenant.ID,
	}

	var bodyBytes bytes.Buffer
	err = json.NewEncoder(&bodyBytes).Encode(body)
	if err != nil {
		t.Fatalf("Failed to encode request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/food-categories", strings.NewReader(bodyBytes.String()))
	w := httptest.NewRecorder()

	// Create handler with test dependencies
	logger := logger.NewLogger()
	foodCategoryRepository := repository.NewFoodCategoryRepository(db.DB, logger, "test-tenant")
	foodCategoryService := foodcategory.NewFoodCategoryService(foodCategoryRepository, logger)

	handler := handlers.NewFoodCategoryHandler(foodCategoryService, logger)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("tenant_id", testTenant.ID)

	handler.CreateCategory(c)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201 but got %d. Body: %s", w.Code, w.Body.String())
	}

	// Parse response
	var response views.SuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Errorf("Response should be successful")
	}

	categoryID := response.Data.(string)
	require.NotEmpty(t, categoryID, "Category ID should not be empty")
	// Verify database state
	err = VerifyDatabaseState(t, db.DB)
	require.NoError(t, err, "Failed to verify database state")
}

// VerifyDatabaseState verifies the database state after an operation
func VerifyDatabaseState(t *testing.T, db *sqlx.DB) error {
	count := 0
	query := `SELECT COUNT(*) FROM food_category`
	err := db.Get(&count, query)
	if err != nil {
		return err
	}

	t.Logf("Database state - Food categories count: %d", count)

	return nil
}
