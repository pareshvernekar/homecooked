package menus

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// REQMENU001, REQMENU003, REQMENU007, REQMENU008, REQITEM001, REQITEM005, REQITEM006
func TestMenu_CreatePublishUnpublishAndDefaultTotal(t *testing.T) {
	env := setupIntegrationEnv(t)
	foodA := createFoodItem(t, env, "Rice Bowl Base")
	foodB := createFoodItem(t, env, "Dal Bowl Base")
	menuID, categoryID := createDraftDailyMenu(t, env, "Tuesday Special")

	// Combo item with two components — default_total should be 5+7=12
	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/menus/"+menuID+"/items", map[string]interface{}{
		"category_id": categoryID,
		"kind":        "combo",
		"name":        "Thali",
		"components": []map[string]interface{}{
			{
				"food_item_id": foodA, "default_size_index": 0,
				"size_options": []map[string]interface{}{
					{"size_unit_id": "su_serving", "qty": 1, "price": 5.0},
				},
			},
			{
				"food_item_id": foodB, "default_size_index": 0,
				"size_options": []map[string]interface{}{
					{"size_unit_id": "su_serving", "qty": 1, "price": 7.0},
				},
			},
		},
	}, gin.Params{{Key: "id", Value: menuID}})
	env.MenuHandler.AddMenuItem(ctx)
	require.Equal(t, http.StatusCreated, resp.Code, "body: %s", resp.Body.String())

	ctx, resp = doJSONRequest(t, http.MethodPost, "/api/v1/menus/"+menuID+"/publish", nil, gin.Params{{Key: "id", Value: menuID}})
	env.MenuHandler.PublishMenu(ctx)
	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())

	// Published: mutations rejected
	addCtx, addResp := doJSONRequest(t, http.MethodPost, "/api/v1/menus/"+menuID+"/items", map[string]interface{}{
		"category_id": categoryID, "kind": "simple", "name": "Extra",
		"components": []map[string]interface{}{
			{"food_item_id": foodA, "default_size_index": 0,
				"size_options": []map[string]interface{}{{"size_unit_id": "su_tray", "qty": 1, "price": 20}}},
		},
	}, gin.Params{{Key: "id", Value: menuID}})
	env.MenuHandler.AddMenuItem(addCtx)
	require.Equal(t, http.StatusBadRequest, addResp.Code, "body: %s", addResp.Body.String())

	ctx, resp = doJSONRequest(t, http.MethodPost, "/api/v1/menus/"+menuID+"/unpublish", nil, gin.Params{{Key: "id", Value: menuID}})
	env.MenuHandler.UnpublishMenu(ctx)
	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())

	ctx, resp = doJSONRequest(t, http.MethodGet, "/api/v1/menus/"+menuID, nil, gin.Params{{Key: "id", Value: menuID}})
	env.MenuHandler.GetMenu(ctx)
	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())
	tree := asObject(t, decodeSuccess(t, resp).Data)
	assert.Equal(t, "draft", tree["status"])
	cats := tree["categories"].([]interface{})
	items := asObject(t, cats[0])["items"].([]interface{})
	require.NotEmpty(t, items)
	item := asObject(t, items[0])
	assert.Equal(t, 12.0, item["default_total"])
	comps := item["components"].([]interface{})
	require.Len(t, comps, 2)
	fi := asObject(t, asObject(t, comps[0])["food_item"])
	assert.Equal(t, "Rice Bowl Base", fi["name"])
	_, hasPrice := fi["price"]
	assert.False(t, hasPrice)
}

// REQMENU002
func TestMenu_ListDefaultsToPublished(t *testing.T) {
	env := setupIntegrationEnv(t)
	foodID := createFoodItem(t, env, "List Dish")
	menuID, categoryID := createDraftDailyMenu(t, env, "Draft Only Menu")
	addSimpleItem(t, env, menuID, categoryID, foodID, "Simple Dish", 9.5)

	ctx, resp := doJSONRequest(t, http.MethodGet, "/api/v1/menus", nil, nil)
	env.MenuHandler.ListMenus(ctx)
	require.Equal(t, http.StatusOK, resp.Code)
	if data := decodeSuccess(t, resp).Data; data != nil {
		arr, ok := data.([]interface{})
		require.True(t, ok)
		for _, row := range arr {
			assert.NotEqual(t, menuID, asObject(t, row)["id"])
		}
	}

	ctx, resp = doJSONRequest(t, http.MethodGet, "/api/v1/menus", nil, nil)
	ctx.Request.URL.RawQuery = "status=draft"
	env.MenuHandler.ListMenus(ctx)
	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())
	drafts, ok := decodeSuccess(t, resp).Data.([]interface{})
	require.True(t, ok, "draft list should be an array")
	found := false
	for _, row := range drafts {
		if asObject(t, row)["id"] == menuID {
			found = true
		}
	}
	assert.True(t, found, "draft menu should appear with status=draft filter")
}

// REQMENU009
func TestMenu_WeeklyPublishedStartDateUnique(t *testing.T) {
	env := setupIntegrationEnv(t)
	foodID := createFoodItem(t, env, "Weekly Dish")
	start, end := "2026-12-07", "2026-12-13"

	createWeekly := func(name string) string {
		ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/menus", map[string]interface{}{
			"name": name, "menu_type": "weekly", "start_date": start, "end_date": end,
			"categories": []map[string]interface{}{{"name": "Mains", "sequence": 1}},
		}, nil)
		env.MenuHandler.CreateMenu(ctx)
		require.Equal(t, http.StatusCreated, resp.Code, "body: %s", resp.Body.String())
		id := asObject(t, decodeSuccess(t, resp).Data)["id"].(string)
		ctx, resp = doJSONRequest(t, http.MethodGet, "/api/v1/menus/"+id, nil, gin.Params{{Key: "id", Value: id}})
		env.MenuHandler.GetMenu(ctx)
		catID := asObject(t, asObject(t, decodeSuccess(t, resp).Data)["categories"].([]interface{})[0])["id"].(string)
		addSimpleItem(t, env, id, catID, foodID, name+" item", 10)
		return id
	}

	first := createWeekly("Week A")
	ctx, resp := doJSONRequest(t, http.MethodPost, "/api/v1/menus/"+first+"/publish", nil, gin.Params{{Key: "id", Value: first}})
	env.MenuHandler.PublishMenu(ctx)
	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())

	second := createWeekly("Week A Dup")
	ctx, resp = doJSONRequest(t, http.MethodPost, "/api/v1/menus/"+second+"/publish", nil, gin.Params{{Key: "id", Value: second}})
	env.MenuHandler.PublishMenu(ctx)
	require.Equal(t, http.StatusBadRequest, resp.Code, "body: %s", resp.Body.String())
}

// REQSIZE001, REQSIZE002
func TestSizeUnits_ListAndCreateCustom(t *testing.T) {
	env := setupIntegrationEnv(t)

	ctx, resp := doJSONRequest(t, http.MethodGet, "/api/v1/size-units", nil, nil)
	env.SizeHandler.ListSizeUnits(ctx)
	require.Equal(t, http.StatusOK, resp.Code, "body: %s", resp.Body.String())
	units := decodeSuccess(t, resp).Data.([]interface{})
	codes := map[string]bool{}
	for _, u := range units {
		codes[asObject(t, u)["code"].(string)] = true
	}
	assert.True(t, codes["serving"])
	assert.True(t, codes["tray"])

	ctx, resp = doJSONRequest(t, http.MethodPost, "/api/v1/size-units", map[string]interface{}{
		"code": "party-pan-" + unusedUUID()[:8], "display_name": "Party Pan",
	}, nil)
	env.SizeHandler.CreateSizeUnit(ctx)
	require.Equal(t, http.StatusCreated, resp.Code, "body: %s", resp.Body.String())
	created := asObject(t, decodeSuccess(t, resp).Data)
	assert.Equal(t, false, created["is_system"])
}
