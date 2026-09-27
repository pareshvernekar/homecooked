package foodcategories

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestCreateFoodCategory(t *testing.T) {
	env := setupIntegrationEnv(t)

	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/categories", map[string]interface{}{
		"name":        "desserts",
		"description": "Sweet treats",
	}, nil)
	env.Handler.CreateCategory(ctx)

	require.Equal(t, http.StatusCreated, resp.Code, "body: %s", resp.Body.String())
	response := decodeSuccess(t, resp)
	require.True(t, response.Success)
	assert.Equal(t, "Food category created successfully", response.Message)

	obj := asObject(t, response.Data)
	assert.NotEmpty(t, obj["id"])
	assert.Equal(t, "desserts", obj["name"])
	assert.Equal(t, "Sweet treats", obj["description"])
	assert.Equal(t, testTenantID, obj["tenant_id"])
	assert.Equal(t, true, obj["is_active"])

	var count int
	err := env.DBConn.Get(&count, `SELECT COUNT(*) FROM food_category WHERE name = $1 AND tenant_id = $2`, "desserts", testTenantID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCreateFoodCategory_BlankName(t *testing.T) {
	env := setupIntegrationEnv(t)

	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/categories", map[string]interface{}{
		"name":        "",
		"description": "missing name",
	}, nil)
	env.Handler.CreateCategory(ctx)

	require.Equal(t, http.StatusBadRequest, resp.Code, "body: %s", resp.Body.String())
}
