package foodcategories

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestDeleteFoodCategory(t *testing.T) {
	env := setupIntegrationEnv(t)
	id := createCategoryViaAPI(t, env, "main_course", "Main dishes")

	ctx, resp := doJSONRequest(t, http.MethodDelete, "/api/v1/categories/"+id, nil, gin.Params{
		{Key: "id", Value: id},
	})
	env.Handler.DeleteCategory(ctx)

	require.Equal(t, http.StatusNoContent, resp.Code, "body: %s", resp.Body.String())

	var isActive bool
	err := env.DBConn.Get(&isActive, `SELECT is_active FROM food_category WHERE id = $1 AND tenant_id = $2`, id, testTenantID)
	require.NoError(t, err)
	assert.False(t, isActive, "delete should soft-delete by setting is_active=false")
}

func TestDeleteFoodCategory_NotFound(t *testing.T) {
	env := setupIntegrationEnv(t)

	ctx, resp := doJSONRequest(t, http.MethodDelete, "/api/v1/categories/non-existent-id", nil, gin.Params{
		{Key: "id", Value: "non-existent-id"},
	})
	env.Handler.DeleteCategory(ctx)

	require.Equal(t, http.StatusInternalServerError, resp.Code, "body: %s", resp.Body.String())
}
