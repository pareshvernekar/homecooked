package menus

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/pareshvernekar/homecooked/internal/handlers"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	"github.com/pareshvernekar/homecooked/internal/services/fooditem"
	"github.com/pareshvernekar/homecooked/internal/services/menu"
	"github.com/pareshvernekar/homecooked/internal/services/menuitem"
	"github.com/pareshvernekar/homecooked/internal/services/sizeunit"
	"github.com/pareshvernekar/homecooked/internal/views"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
	testhttp "github.com/pareshvernekar/homecooked/tests/http"
)

const testTenantID = "test-tenant"

type integrationEnv struct {
	DB          *testdb.DatabaseHelper
	DBConn      *sqlx.DB
	MenuHandler *handlers.MenuHandler
	SizeHandler *handlers.SizeUnitHandler
	FoodHandler *handlers.FoodItemHandler
}

func setupIntegrationEnv(t *testing.T) *integrationEnv {
	t.Helper()
	db, err := testdb.NewDatabaseHelper(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Terminate(t.Context())) })

	require.NoError(t, testdb.InitializeSchema(t.Context(), db.DB))
	testTenant, err := testdb.CreateTestTenant(t.Context(), db.DB)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testTenant.Cleanup() })

	l := logger.NewLogger()
	categoryRepo := repository.NewFoodCategoryRepository(db.DB, l, testTenantID)
	categoryService := foodcategory.NewFoodCategoryService(categoryRepo, l)
	foodRepo := repository.NewFoodItemRepository(db.DB, l, testTenantID)
	foodService := fooditem.NewFoodItemService(foodRepo, l, categoryService)
	sizeRepo := repository.NewSizeUnitRepository(db.DB, l, testTenantID)
	sizeService := sizeunit.NewService(sizeRepo, l)
	menuRepo := repository.NewMenuRepository(db.DB, l, testTenantID)
	menuItemRepo := repository.NewMenuItemRepository(db.DB, l, testTenantID)
	menuSvc := menu.NewService(menuRepo, menuItemRepo, l)
	menuItemSvc := menuitem.NewService(menuItemRepo, menuSvc, sizeService, l)

	return &integrationEnv{
		DB:          db,
		DBConn:      db.DB,
		MenuHandler: handlers.NewMenuHandler(menuSvc, menuItemSvc, l),
		SizeHandler: handlers.NewSizeUnitHandler(sizeService, l),
		FoodHandler: handlers.NewFoodItemHandler(foodService, l),
	}
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
	require.True(t, ok, "expected object, got %T", data)
	return obj
}

func createFoodItem(t *testing.T, env *integrationEnv, name string) string {
	t.Helper()
	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/food-items", map[string]interface{}{
		"name": name, "description": name, "category_name": "vegetarian", "availability_status": "available",
	}, nil)
	env.FoodHandler.CreateFoodItem(ctx)
	require.Equal(t, http.StatusCreated, resp.Code, "body: %s", resp.Body.String())
	return asObject(t, decodeSuccess(t, resp).Data)["id"].(string)
}

func createDraftDailyMenu(t *testing.T, env *integrationEnv, name string) (menuID, categoryID string) {
	t.Helper()
	today := time.Now().UTC().Format("2006-01-02")
	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/menus", map[string]interface{}{
		"name": name, "menu_type": "daily", "menu_date": today,
		"categories": []map[string]interface{}{{"name": "Mains", "sequence": 1}},
	}, nil)
	env.MenuHandler.CreateMenu(ctx)
	require.Equal(t, http.StatusCreated, resp.Code, "body: %s", resp.Body.String())
	menuID = asObject(t, decodeSuccess(t, resp).Data)["id"].(string)

	ctx, resp = doJSONRequest(t, http.MethodGet, "/api/v1/menus/"+menuID, nil, gin.Params{{Key: "id", Value: menuID}})
	env.MenuHandler.GetMenu(ctx)
	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())
	tree := asObject(t, decodeSuccess(t, resp).Data)
	cats := tree["categories"].([]interface{})
	require.NotEmpty(t, cats)
	categoryID = asObject(t, cats[0])["id"].(string)
	return menuID, categoryID
}

func addSimpleItem(t *testing.T, env *integrationEnv, menuID, categoryID, foodItemID, name string, price float64) {
	t.Helper()
	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/menus/"+menuID+"/items", map[string]interface{}{
		"category_id": categoryID,
		"kind":        "simple",
		"name":        name,
		"components": []map[string]interface{}{
			{
				"food_item_id":       foodItemID,
				"default_size_index": 0,
				"size_options": []map[string]interface{}{
					{"size_unit_id": "su_serving", "qty": 1, "price": price},
				},
			},
		},
	}, gin.Params{{Key: "id", Value: menuID}})
	env.MenuHandler.AddMenuItem(ctx)
	require.Equal(t, http.StatusCreated, resp.Code, "body: %s", resp.Body.String())
}

func unusedUUID() string { return uuid.New().String() }
