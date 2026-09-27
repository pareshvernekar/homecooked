package fooditems

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/handlers"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	"github.com/pareshvernekar/homecooked/internal/services/fooditem"
	"github.com/pareshvernekar/homecooked/internal/views"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
	testhttp "github.com/pareshvernekar/homecooked/tests/http"
	"github.com/stretchr/testify/require"
)

const testTenantID = "test-tenant"

type integrationEnv struct {
	DB      *testdb.DatabaseHelper
	Handler *handlers.FoodItemHandler
	DBConn  *sqlx.DB
}

func setupIntegrationEnv(t *testing.T) *integrationEnv {
	t.Helper()

	db, err := testdb.NewDatabaseHelper(t.Context())
	require.NoError(t, err, "failed to create PostgreSQL connection")
	t.Cleanup(func() {
		require.NoError(t, db.Terminate(t.Context()))
	})

	require.NoError(t, testdb.InitializeSchema(t.Context(), db.DB))

	testTenant, err := testdb.CreateTestTenant(t.Context(), db.DB)
	require.NoError(t, err, "failed to create test tenant / seed vegetarian category")
	t.Cleanup(func() {
		_ = testTenant.Cleanup()
	})

	l := logger.NewLogger()
	categoryRepo := repository.NewFoodCategoryRepository(db.DB, l, testTenantID)
	categoryService := foodcategory.NewFoodCategoryService(categoryRepo, l)
	itemRepo := repository.NewFoodItemRepository(db.DB, l, testTenantID)
	itemService := fooditem.NewFoodItemService(itemRepo, l, categoryService)
	handler := handlers.NewFoodItemHandler(itemService, l)

	return &integrationEnv{DB: db, Handler: handler, DBConn: db.DB}
}

func doJSONRequest(t *testing.T, method, path string, body map[string]interface{}, params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	var bodyReader *strings.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = strings.NewReader(string(raw))
	} else {
		bodyReader = strings.NewReader("")
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", testTenantID)

	ctx, resp := testhttp.CreateTestContext(req)
	if params != nil {
		ctx.Params = params
	}
	return ctx, resp
}

func decodeSuccess(t *testing.T, resp *httptest.ResponseRecorder) views.SuccessResponse {
	t.Helper()
	var response views.SuccessResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &response))
	return response
}

func asObject(t *testing.T, data interface{}) map[string]interface{} {
	t.Helper()
	obj, ok := data.(map[string]interface{})
	require.True(t, ok, "expected response data object, got %T", data)
	return obj
}

func asArray(t *testing.T, data interface{}) []interface{} {
	t.Helper()
	arr, ok := data.([]interface{})
	require.True(t, ok, "expected response data array, got %T", data)
	return arr
}

func createFoodItemViaAPI(t *testing.T, env *integrationEnv, name string, price float64) string {
	t.Helper()

	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/food-items", map[string]interface{}{
		"name":                name,
		"description":         name + " description",
		"price":               price,
		"category_name":       "vegetarian",
		"availability_status": "available",
	}, nil)
	env.Handler.CreateFoodItem(ctx)
	require.Equal(t, http.StatusCreated, resp.Code, "body: %s", resp.Body.String())

	response := decodeSuccess(t, resp)
	obj := asObject(t, response.Data)
	id, ok := obj["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, id)
	return id
}
