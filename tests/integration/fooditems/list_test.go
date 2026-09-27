package fooditems

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestListFoodItems(t *testing.T) {
	env := setupIntegrationEnv(t)

	item1ID := createFoodItemViaAPI(t, env, "Chicken Biryani", 249.99)
	item2ID := createFoodItemViaAPI(t, env, "Paneer Tikka", 199.00)

	ctx, resp := doJSONRequest(t, http.MethodGet, "/api/v1/food-items?page=1&limit=10", nil, nil)
	env.Handler.GetFoodItems(ctx)

	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())
	response := decodeSuccess(t, resp)
	require.True(t, response.Success)

	items := asArray(t, response.Data)
	require.GreaterOrEqual(t, len(items), 2)

	found := map[string]bool{}
	for _, item := range items {
		obj := asObject(t, item)
		id, _ := obj["id"].(string)
		found[id] = true
		assert.Equal(t, testTenantID, obj["tenant_id"])
	}
	assert.True(t, found[item1ID], "expected item %s in list", item1ID)
	assert.True(t, found[item2ID], "expected item %s in list", item2ID)
}

func TestListFoodItems_Pagination(t *testing.T) {
	env := setupIntegrationEnv(t)

	_ = createFoodItemViaAPI(t, env, "Dish 1", 100.00)
	_ = createFoodItemViaAPI(t, env, "Dish 2", 150.00)
	_ = createFoodItemViaAPI(t, env, "Dish 3", 175.00)

	ctx, resp := doJSONRequest(t, http.MethodGet, "/api/v1/food-items?page=1&limit=2", nil, nil)
	env.Handler.GetFoodItems(ctx)

	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())
	response := decodeSuccess(t, resp)
	require.True(t, response.Success)

	items := asArray(t, response.Data)
	assert.LessOrEqual(t, len(items), 2, "page limit should cap returned items")
	assert.NotEmpty(t, items)
}
