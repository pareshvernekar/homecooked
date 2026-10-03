package orders

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/pareshvernekar/homecooked/internal/handlers"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/server"
	"github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	"github.com/pareshvernekar/homecooked/internal/services/fooditem"
	"github.com/pareshvernekar/homecooked/internal/services/menu"
	"github.com/pareshvernekar/homecooked/internal/services/menuitem"
	"github.com/pareshvernekar/homecooked/internal/services/notification"
	"github.com/pareshvernekar/homecooked/internal/services/order"
	"github.com/pareshvernekar/homecooked/internal/services/sizeunit"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
)

const testTenantID = "test-tenant"

// env serves the real route table (server.SetupRoutes) over a Postgres test container.
type env struct {
	router    *gin.Engine
	notifRepo *repository.PostgreSQLNotificationRepository
	logger    *logger.Logger
}

func setupEnv(t *testing.T) *env {
	t.Helper()
	db, err := testdb.NewDatabaseHelper(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Terminate(t.Context())) })
	require.NoError(t, testdb.InitializeSchema(t.Context(), db.DB))
	tenant, err := testdb.CreateTestTenant(t.Context(), db.DB)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tenant.Cleanup() })

	l := logger.NewLogger()
	categoryService := foodcategory.NewFoodCategoryService(repository.NewFoodCategoryRepository(db.DB, l, testTenantID), l)
	foodService := fooditem.NewFoodItemService(repository.NewFoodItemRepository(db.DB, l, testTenantID), l, categoryService)
	sizeService := sizeunit.NewService(repository.NewSizeUnitRepository(db.DB, l, testTenantID), l)
	menuItemRepo := repository.NewMenuItemRepository(db.DB, l, testTenantID)
	menuSvc := menu.NewService(repository.NewMenuRepository(db.DB, l, testTenantID), menuItemRepo, l)
	menuItemSvc := menuitem.NewService(menuItemRepo, menuSvc, sizeService, l)
	notifRepo := repository.NewNotificationRepository(db.DB, l)
	orderSvc := order.NewService(repository.NewOrderRepository(db.DB, l, testTenantID), menuSvc, l,
		order.WithNotifications(notification.NewBuilder(notifRepo, l)))
	notificationSvc := notification.NewService(notifRepo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.TenantMiddleware(l))
	server.SetupRoutes(router, db.DB, l,
		handlers.NewFoodCategoryHandler(categoryService, l),
		handlers.NewFoodItemHandler(foodService, l),
		handlers.NewSizeUnitHandler(sizeService, l),
		handlers.NewMenuHandler(menuSvc, menuItemSvc, l),
		handlers.NewOrderHandler(orderSvc, l),
		handlers.NewNotificationHandler(notificationSvc, l),
	)
	return &env{router: router, notifRepo: notifRepo, logger: l}
}

type apiResponse struct {
	Code int
	Body map[string]interface{}
}

func (a apiResponse) Data() map[string]interface{} {
	d, _ := a.Body["data"].(map[string]interface{})
	return d
}

func (a apiResponse) List() []interface{} {
	d, _ := a.Body["data"].([]interface{})
	return d
}

// call sends a JSON request through the router with the tenant header.
func (e *env) call(t *testing.T, method, path string, body interface{}) apiResponse {
	t.Helper()
	var reader *strings.Reader
	switch b := body.(type) {
	case nil:
		reader = strings.NewReader("")
	case string:
		reader = strings.NewReader(b) // raw JSON, e.g. to send explicit nulls
	default:
		raw, err := json.Marshal(b)
		require.NoError(t, err)
		reader = strings.NewReader(string(raw))
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", testTenantID)
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)

	out := apiResponse{Code: w.Code}
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out.Body), "body: %s", w.Body.String())
	}
	return out
}

func (e *env) mustCall(t *testing.T, want int, method, path string, body interface{}) apiResponse {
	t.Helper()
	r := e.call(t, method, path, body)
	require.Equal(t, want, r.Code, "%s %s -> %v", method, path, r.Body)
	return r
}
