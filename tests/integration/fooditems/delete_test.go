package fooditems

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestDeleteFoodItem(t *testing.T) {
	env := setupIntegrationEnv(t)
	id := createFoodItemViaAPI(t, env, "Chicken Biryani", 249.99)

	ctx, resp := doJSONRequest(t, http.MethodDelete, "/api/v1/food-items/"+id, nil, gin.Params{
		{Key: "id", Value: id},
	})
	env.Handler.DeleteFoodItem(ctx)

	require.Equal(t, http.StatusNoContent, resp.Code, "body: %s", resp.Body.String())

	var isActive bool
	err := env.DBConn.Get(&isActive, `SELECT is_active FROM food_item WHERE id = $1 AND tenant_id = $2`, id, testTenantID)
	require.NoError(t, err)
	assert.False(t, isActive, "delete should soft-delete by setting is_active=false")
}

func TestDeleteFoodItem_NotFound(t *testing.T) {
	env := setupIntegrationEnv(t)

	ctx, resp := doJSONRequest(t, http.MethodDelete, "/api/v1/food-items/non-existent-id", nil, gin.Params{
		{Key: "id", Value: "non-existent-id"},
	})
	env.Handler.DeleteFoodItem(ctx)

	// Idempotent delete: missing item still returns 204
	require.Equal(t, http.StatusNoContent, resp.Code, "body: %s", resp.Body.String())
}
