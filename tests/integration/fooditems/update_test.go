package fooditems

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFoodItem(t *testing.T) {
	env := setupIntegrationEnv(t)
	id := createFoodItemViaAPI(t, env, "Chicken Biryani")

	ctx, resp := doJSONRequest(t, http.MethodPut, "/api/v1/food-items/"+id, map[string]interface{}{
		"name":                "Chicken Biryani Deluxe",
		"availability_status": "available",
	}, gin.Params{{Key: "id", Value: id}})
	env.Handler.UpdateFoodItem(ctx)

	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())
	response := decodeSuccess(t, resp)
	require.True(t, response.Success)
	assert.Equal(t, "Food item updated successfully", response.Message)

	var name string
	err := env.DBConn.QueryRow(
		`SELECT name FROM food_item WHERE id = $1 AND tenant_id = $2`, id, testTenantID,
	).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Chicken Biryani Deluxe", name)
}

func TestUpdateFoodItem_PartialAvailability(t *testing.T) {
	env := setupIntegrationEnv(t)
	id := createFoodItemViaAPI(t, env, "Paneer Tikka")

	ctx, resp := doJSONRequest(t, http.MethodPut, "/api/v1/food-items/"+id, map[string]interface{}{
		"availability_status": "low_stock",
	}, gin.Params{{Key: "id", Value: id}})
	env.Handler.UpdateFoodItem(ctx)

	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())

	var status, name string
	err := env.DBConn.QueryRow(
		`SELECT availability_status, name FROM food_item WHERE id = $1 AND tenant_id = $2`, id, testTenantID,
	).Scan(&status, &name)
	require.NoError(t, err)
	assert.Equal(t, "low_stock", status)
	assert.Equal(t, "Paneer Tikka", name, "partial update must preserve name")
}
