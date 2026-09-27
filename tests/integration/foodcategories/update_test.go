package foodcategories

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestUpdateFoodCategory(t *testing.T) {
	env := setupIntegrationEnv(t)
	id := createCategoryViaAPI(t, env, "sides", "Side dishes")

	ctx, resp := doJSONRequest(t, http.MethodPut, "/api/v1/categories/"+id, map[string]interface{}{
		"name":        "sides",
		"description": "Updated side dishes",
		"is_active":   true,
	}, gin.Params{{Key: "id", Value: id}})
	env.Handler.UpdateCategory(ctx)

	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())
	response := decodeSuccess(t, resp)
	require.True(t, response.Success)
	assert.Equal(t, "Food category updated successfully", response.Message)

	obj := asObject(t, response.Data)
	assert.Equal(t, id, obj["id"])
	assert.Equal(t, "Updated side dishes", obj["description"])

	var description string
	err := env.DBConn.Get(&description, `SELECT description FROM food_category WHERE id = $1 AND tenant_id = $2`, id, testTenantID)
	require.NoError(t, err)
	assert.Equal(t, "Updated side dishes", description)
}

func TestUpdateFoodCategory_BlankName(t *testing.T) {
	env := setupIntegrationEnv(t)
	id := createCategoryViaAPI(t, env, "vegan", "Plant based")

	ctx, resp := doJSONRequest(t, http.MethodPut, "/api/v1/categories/"+id, map[string]interface{}{
		"description": "no name provided",
	}, gin.Params{{Key: "id", Value: id}})
	env.Handler.UpdateCategory(ctx)

	require.Equal(t, http.StatusBadRequest, resp.Code, "body: %s", resp.Body.String())
}
