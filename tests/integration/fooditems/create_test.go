package fooditems

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// REQFOOD001
func TestCreateFoodItem(t *testing.T) {
	env := setupIntegrationEnv(t)

	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/food-items", map[string]interface{}{
		"name":                "Chicken Biryani",
		"description":         "Rich and flavorful biryani with aromatic spices",
		"category_name":       "vegetarian",
		"availability_status": "available",
		"is_vegetarian":       true,
	}, nil)
	env.Handler.CreateFoodItem(ctx)

	require.Equal(t, http.StatusCreated, resp.Code, "body: %s", resp.Body.String())
	response := decodeSuccess(t, resp)
	require.True(t, response.Success)
	assert.Equal(t, "Food item created successfully", response.Message)

	obj := asObject(t, response.Data)
	id, ok := obj["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, id, "created food item must have a generated ID")
	assert.Equal(t, "Chicken Biryani", obj["name"])
	_, hasPrice := obj["price"]
	assert.False(t, hasPrice, "REQFOOD001: response must not include catalog price")
	assert.Equal(t, testTenantID, obj["tenant_id"])
	assert.Equal(t, "available", obj["availability_status"])
	assert.NotEmpty(t, obj["category_id"])

	var dbName string
	err := env.DBConn.QueryRow(
		`SELECT name FROM food_item WHERE id = $1 AND tenant_id = $2`, id, testTenantID,
	).Scan(&dbName)
	require.NoError(t, err)
	assert.Equal(t, "Chicken Biryani", dbName)
}

func TestCreateFoodItem_UnknownCategory(t *testing.T) {
	env := setupIntegrationEnv(t)

	// "vegan" is a valid category enum but is not seeded for this tenant.
	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/food-items", map[string]interface{}{
		"name":                "Mystery Dish",
		"description":         "Category missing from tenant catalog",
		"category_name":       "vegan",
		"availability_status": "available",
	}, nil)
	env.Handler.CreateFoodItem(ctx)

	require.Equal(t, http.StatusInternalServerError, resp.Code, "body: %s", resp.Body.String())
}

func TestCreateFoodItem_ValidationError(t *testing.T) {
	env := setupIntegrationEnv(t)

	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/food-items", map[string]interface{}{
		"description": "missing required fields",
	}, nil)
	env.Handler.CreateFoodItem(ctx)

	require.Equal(t, http.StatusBadRequest, resp.Code, "body: %s", resp.Body.String())
}

// REQFOOD001S02: price in payload is ignored (not persisted).
func TestCreateFoodItem_IgnoresPriceField(t *testing.T) {
	env := setupIntegrationEnv(t)

	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/food-items", map[string]interface{}{
		"name":                "Priced Payload",
		"description":         "client still sends price",
		"category_name":       "vegetarian",
		"availability_status": "available",
		"price":               99.99,
	}, nil)
	env.Handler.CreateFoodItem(ctx)

	require.Equal(t, http.StatusCreated, resp.Code, "body: %s", resp.Body.String())
	obj := asObject(t, decodeSuccess(t, resp).Data)
	_, hasPrice := obj["price"]
	assert.False(t, hasPrice)
}
