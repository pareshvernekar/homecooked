package foodcategories

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestListFoodCategories(t *testing.T) {
	env := setupIntegrationEnv(t)

	_ = createCategoryViaAPI(t, env, "appetizer", "Starters")
	_ = createCategoryViaAPI(t, env, "beverage", "Drinks")

	ctx, resp := doJSONRequest(t, http.MethodGet, "/api/v1/categories", nil, nil)
	env.Handler.ListCategories(ctx)

	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())
	response := decodeSuccess(t, resp)
	require.True(t, response.Success)

	items := asArray(t, response.Data)
	// CreateTestTenant seeds "vegetarian"; plus the two created above.
	require.GreaterOrEqual(t, len(items), 3)

	names := map[string]bool{}
	for _, item := range items {
		obj := asObject(t, item)
		name, _ := obj["name"].(string)
		names[name] = true
		assert.Equal(t, testTenantID, obj["tenant_id"])
	}
	assert.True(t, names["vegetarian"])
	assert.True(t, names["appetizer"])
	assert.True(t, names["beverage"])
}
