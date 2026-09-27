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
	id := createFoodItemViaAPI(t, env, "Chicken Biryani", 249.99)

	ctx, resp := doJSONRequest(t, http.MethodPut, "/api/v1/food-items/"+id, map[string]interface{}{
		"price":               349.99,
		"availability_status": "available",
	}, gin.Params{{Key: "id", Value: id}})
	env.Handler.UpdateFoodItem(ctx)

	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())
	response := decodeSuccess(t, resp)
	require.True(t, response.Success)
	assert.Equal(t, "Food item updated successfully", response.Message)

	var name string
	var price float64
	err := env.DBConn.QueryRow(
		`SELECT name, price FROM food_item WHERE id = $1 AND tenant_id = $2`, id, testTenantID,
	).Scan(&name, &price)
	require.NoError(t, err)
	assert.Equal(t, "Chicken Biryani", name, "partial update must preserve name")
	assert.Equal(t, 349.99, price)
}

func TestUpdateFoodItem_PartialAvailability(t *testing.T) {
	env := setupIntegrationEnv(t)
	id := createFoodItemViaAPI(t, env, "Paneer Tikka", 199.00)

	ctx, resp := doJSONRequest(t, http.MethodPut, "/api/v1/food-items/"+id, map[string]interface{}{
		"availability_status": "low_stock",
	}, gin.Params{{Key: "id", Value: id}})
	env.Handler.UpdateFoodItem(ctx)

	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())

	var status string
	var price float64
	err := env.DBConn.QueryRow(
		`SELECT availability_status, price FROM food_item WHERE id = $1 AND tenant_id = $2`, id, testTenantID,
	).Scan(&status, &price)
	require.NoError(t, err)
	assert.Equal(t, "low_stock", status)
	assert.Equal(t, 199.00, price, "partial update must preserve price")
}
