package orders

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type seededMenu struct {
	menuID, itemID, componentID, optionID string
}

// seedPublishedDailyMenu creates a published daily menu with one simple item priced at 12.
func seedPublishedDailyMenu(t *testing.T, e *env, menuType string) seededMenu {
	t.Helper()
	food := e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/food-items", map[string]interface{}{
		"name": "Rice " + menuType, "description": "rice", "category_name": "vegetarian", "availability_status": "available",
	}).Data()["id"].(string)

	body := map[string]interface{}{
		"name": "Menu " + menuType, "menu_type": menuType,
		"categories": []map[string]interface{}{{"name": "Mains", "sequence": 1}},
	}
	if menuType == "catering" {
		body["event_date"] = "2026-12-01"
		body["event_location"] = "Community Hall"
	} else {
		body["menu_date"] = "2026-12-01"
	}
	menuID := e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/menus", body).Data()["id"].(string)
	tree := e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/menus/"+menuID, nil).Data()
	catID := tree["categories"].([]interface{})[0].(map[string]interface{})["id"].(string)

	e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/menus/"+menuID+"/items", map[string]interface{}{
		"category_id": catID, "kind": "simple", "name": "Rice Bowl",
		"components": []map[string]interface{}{{
			"food_item_id": food, "default_size_index": 0,
			"size_options": []map[string]interface{}{{"size_unit_id": "su_serving", "qty": 1, "price": 12}},
		}},
	})
	e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/menus/"+menuID+"/publish", nil)

	tree = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/menus/"+menuID, nil).Data()
	item := tree["categories"].([]interface{})[0].(map[string]interface{})["items"].([]interface{})[0].(map[string]interface{})
	comp := item["components"].([]interface{})[0].(map[string]interface{})
	opt := comp["size_options"].([]interface{})[0].(map[string]interface{})
	return seededMenu{menuID, item["id"].(string), comp["id"].(string), opt["id"].(string)}
}

func setOptionPrice(t *testing.T, e *env, m seededMenu, price float64) {
	t.Helper()
	e.mustCall(t, http.StatusOK, http.MethodPut,
		"/api/v1/menus/"+m.menuID+"/items/"+m.itemID+"/components/"+m.componentID+"/size-options/"+m.optionID,
		map[string]interface{}{"price": price})
}

// REQORDER001–REQORDER004, REQOLINE001–REQOLINE004, REQPAY001–REQPAY004, REQITEM006S03
func TestOrder_CreateLinesPayPickedUpFreeze(t *testing.T) {
	e := setupEnv(t)
	m := seedPublishedDailyMenu(t, e, "daily")
	expected := time.Now().Add(3 * time.Hour).UnixMilli()

	// REQORDER001S01
	created := e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/orders", map[string]interface{}{
		"menu_id": m.menuID, "customer_name": "Asha", "customer_phone": "555-0100",
		"expected_at": expected, "customization_text": "nut-free",
	}).Data()
	orderID := created["id"].(string)
	assert.Equal(t, "RECEIVED", created["status"])
	assert.Equal(t, testTenantID, created["tenant_id"])
	assert.Equal(t, m.menuID, created["menu_id"])
	assert.NotZero(t, created["received_at"])

	// REQOLINE001S01 / REQOLINE002S01: default selection, customization
	line := e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/orders/"+orderID+"/items", map[string]interface{}{
		"menu_item_id": m.itemID, "quantity": 2, "customization_text": "spicy",
	}).Data()
	lineID := line["id"].(string)
	assert.Equal(t, 12.0, line["unit_price"])
	assert.Equal(t, 24.0, line["extended_amount"])
	sels := line["selections"].([]interface{})
	require.Len(t, sels, 1)
	assert.Equal(t, m.optionID, sels[0].(map[string]interface{})["size_option_id"])

	got := e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID, nil).Data()
	assert.Equal(t, 24.0, got["charged_total"])
	assert.Equal(t, 0.0, got["paid_amount"])
	assert.Equal(t, 24.0, got["balance"])
	assert.Equal(t, false, got["payment_received"])
	assert.Equal(t, 0.0, got["overpaid_amount"])

	// REQITEM006S03 + REQOLINE003S01: price-only update on published menu moves live total
	setOptionPrice(t, e, m, 15)
	got = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID, nil).Data()
	assert.Equal(t, 30.0, got["charged_total"])

	// REQITEM006S02: structural update on published menu is rejected
	e.mustCall(t, http.StatusBadRequest, http.MethodPut,
		"/api/v1/menus/"+m.menuID+"/items/"+m.itemID+"/components/"+m.componentID+"/size-options/"+m.optionID,
		map[string]interface{}{"qty": 3})

	// REQOLINE002: override then clear with explicit JSON null (REQOLINE002S02)
	e.mustCall(t, http.StatusOK, http.MethodPatch, "/api/v1/orders/"+orderID+"/items/"+lineID,
		map[string]interface{}{"unit_price_override": 10})
	got = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID, nil).Data()
	assert.Equal(t, 20.0, got["charged_total"])
	e.mustCall(t, http.StatusOK, http.MethodPatch, "/api/v1/orders/"+orderID+"/items/"+lineID, `{"unit_price_override": null}`)
	got = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID, nil).Data()
	assert.Equal(t, 30.0, got["charged_total"])

	// REQORDER003S02: total_override wins, then clear with null
	patched := e.mustCall(t, http.StatusOK, http.MethodPatch, "/api/v1/orders/"+orderID,
		map[string]interface{}{"total_override": 25}).Data()
	assert.Equal(t, 25.0, patched["charged_total"])
	patched = e.mustCall(t, http.StatusOK, http.MethodPatch, "/api/v1/orders/"+orderID, `{"total_override": null}`).Data()
	assert.Equal(t, 30.0, patched["charged_total"])

	// REQPAY001/REQPAY002: split payments
	e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/orders/"+orderID+"/payments",
		map[string]interface{}{"mode": "cash", "amount": 10, "reference": "till"})
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders/"+orderID+"/payments",
		map[string]interface{}{"mode": "gold", "amount": 10})
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders/"+orderID+"/payments",
		map[string]interface{}{"mode": "cash", "amount": 0})
	got = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID, nil).Data()
	assert.Equal(t, 10.0, got["paid_amount"])
	assert.Equal(t, 20.0, got["balance"])
	assert.Equal(t, false, got["payment_received"])

	// REQORDER004S02/S03: PICKEDUP while unpaid freezes total
	// REQLIFE005S01: status via PATCH is rejected
	e.mustCall(t, http.StatusBadRequest, http.MethodPatch, "/api/v1/orders/"+orderID, map[string]interface{}{"status": "PICKEDUP"})
	// REQLIFE001S02: cannot skip ahead
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders/"+orderID+"/ready", nil)
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders/"+orderID+"/pickup", nil)
	// REQLIFE001S01
	assert.Equal(t, "ACCEPTED", e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+orderID+"/accept", nil).Data()["status"])
	assert.Equal(t, "IN_PROGRESS", e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+orderID+"/start-preparing", nil).Data()["status"])
	assert.Equal(t, "READY", e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+orderID+"/ready", nil).Data()["status"])
	picked := e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+orderID+"/pickup", nil).Data()
	assert.Equal(t, "PICKEDUP", picked["status"])
	assert.NotNil(t, picked["pickedup_at"])
	assert.Equal(t, 30.0, picked["frozen_total"])
	assert.Equal(t, false, picked["payment_received"])

	setOptionPrice(t, e, m, 99) // REQOLINE003S03
	got = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID, nil).Data()
	assert.Equal(t, 30.0, got["charged_total"])
	assert.Equal(t, 20.0, got["balance"])

	// REQORDER003S03 / REQOLINE001S04: rejected after PICKEDUP
	e.mustCall(t, http.StatusBadRequest, http.MethodPatch, "/api/v1/orders/"+orderID, map[string]interface{}{"total_override": 1})
	e.mustCall(t, http.StatusBadRequest, http.MethodPatch, "/api/v1/orders/"+orderID, `{"total_override": null}`)
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders/"+orderID+"/items",
		map[string]interface{}{"menu_item_id": m.itemID, "quantity": 1})
	e.mustCall(t, http.StatusBadRequest, http.MethodPatch, "/api/v1/orders/"+orderID+"/items/"+lineID,
		map[string]interface{}{"quantity": 5})
	e.mustCall(t, http.StatusBadRequest, http.MethodDelete, "/api/v1/orders/"+orderID+"/items/"+lineID, nil)

	// REQPAY001S03: payment after PICKEDUP
	e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/orders/"+orderID+"/payments",
		map[string]interface{}{"mode": "venmo", "amount": 20})
	got = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID, nil).Data()
	assert.Equal(t, 30.0, got["paid_amount"])
	assert.Equal(t, true, got["payment_received"])

	// REQPAY004S01
	payments := e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID+"/payments", nil).List()
	require.Len(t, payments, 2)
	assert.Equal(t, "cash", payments[0].(map[string]interface{})["mode"])
	assert.Equal(t, "venmo", payments[1].(map[string]interface{})["mode"])
}

// REQOLINE001S03, REQOLINE004S01, REQOLINE004S02
func TestOrder_LineSelectionAndRemoval(t *testing.T) {
	e := setupEnv(t)
	m := seedPublishedDailyMenu(t, e, "daily")
	orderID := e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/orders", map[string]interface{}{
		"menu_id": m.menuID, "customer_name": "Asha", "customer_phone": "555", "expected_at": time.Now().Add(time.Hour).UnixMilli(),
	}).Data()["id"].(string)

	// invalid size option -> 400 and no line added
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders/"+orderID+"/items", map[string]interface{}{
		"menu_item_id": m.itemID, "quantity": 1,
		"selections": []map[string]interface{}{{"menu_item_component_id": m.componentID, "size_option_id": "nope"}},
	})
	got := e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID, nil).Data()
	assert.Empty(t, got["lines"])

	line := e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/orders/"+orderID+"/items", map[string]interface{}{
		"menu_item_id": m.itemID, "quantity": 1,
		"selections": []map[string]interface{}{{"menu_item_component_id": m.componentID, "size_option_id": m.optionID}},
	}).Data()
	lineID := line["id"].(string)

	e.mustCall(t, http.StatusOK, http.MethodPatch, "/api/v1/orders/"+orderID+"/items/"+lineID,
		map[string]interface{}{"quantity": 3,
			"selections": []map[string]interface{}{{"menu_item_component_id": m.componentID, "size_option_id": m.optionID}}})
	got = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID, nil).Data()
	assert.Equal(t, 36.0, got["charged_total"])

	e.mustCall(t, http.StatusNoContent, http.MethodDelete, "/api/v1/orders/"+orderID+"/items/"+lineID, nil)
	got = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID, nil).Data()
	assert.Empty(t, got["lines"])
	e.mustCall(t, http.StatusNotFound, http.MethodDelete, "/api/v1/orders/"+orderID+"/items/"+lineID, nil)
}

// REQORDER001S02, REQORDER001S03, REQORDER002S01, REQORDER002S03
func TestOrder_CreateGatesAndList(t *testing.T) {
	e := setupEnv(t)
	daily := seedPublishedDailyMenu(t, e, "daily")
	catering := seedPublishedDailyMenu(t, e, "catering")
	expected := time.Now().Add(time.Hour).UnixMilli()

	// missing required header fields
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders",
		map[string]interface{}{"menu_id": daily.menuID, "customer_phone": "1", "expected_at": expected})
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders",
		map[string]interface{}{"menu_id": daily.menuID, "customer_name": "n", "expected_at": expected})
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders",
		map[string]interface{}{"menu_id": daily.menuID, "customer_name": "n", "customer_phone": "1"})

	// draft menu
	draftID := e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/menus", map[string]interface{}{
		"name": "Draft", "menu_type": "daily", "menu_date": "2026-12-02",
	}).Data()["id"].(string)
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders",
		map[string]interface{}{"menu_id": draftID, "customer_name": "n", "customer_phone": "1", "expected_at": expected})

	// unknown menu
	e.mustCall(t, http.StatusNotFound, http.MethodPost, "/api/v1/orders",
		map[string]interface{}{"menu_id": "missing", "customer_name": "n", "customer_phone": "1", "expected_at": expected})

	// catering menu is allowed
	e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/orders",
		map[string]interface{}{"menu_id": catering.menuID, "customer_name": "n", "customer_phone": "1", "expected_at": expected})
	e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/orders",
		map[string]interface{}{"menu_id": daily.menuID, "customer_name": "n2", "customer_phone": "1", "expected_at": expected})

	list := e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders", nil).List()
	assert.Len(t, list, 2)

	e.mustCall(t, http.StatusNotFound, http.MethodGet, "/api/v1/orders/does-not-exist", nil)
	e.mustCall(t, http.StatusNotFound, http.MethodPost, "/api/v1/orders/does-not-exist/payments",
		map[string]interface{}{"mode": "cash", "amount": 5})
}

// REQLIFE003S01, REQLIFE003S02, REQLIFE001S03, REQPAY001S04, REQOLINE001S06, REQORDER003S04
func TestOrder_RefuseDeclined(t *testing.T) {
	e := setupEnv(t)
	m := seedPublishedDailyMenu(t, e, "daily")
	expected := time.Now().Add(time.Hour).UnixMilli()
	newOrder := func() string {
		return e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/orders", map[string]interface{}{
			"menu_id": m.menuID, "customer_name": "Asha", "customer_phone": "555", "expected_at": expected,
		}).Data()["id"].(string)
	}

	// default reason, no body
	id := newOrder()
	d := e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+id+"/refuse", nil).Data()
	assert.Equal(t, "DECLINED", d["status"])
	assert.Equal(t, "No available slots", d["refuse_reason"])
	got := e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+id, nil).Data()
	assert.Equal(t, "No available slots", got["refuse_reason"])

	// DECLINED is terminal and read-only
	for _, action := range []string{"accept", "start-preparing", "ready", "pickup", "refuse"} {
		e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders/"+id+"/"+action, nil)
	}
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders/"+id+"/payments",
		map[string]interface{}{"mode": "cash", "amount": 5})
	e.mustCall(t, http.StatusBadRequest, http.MethodPost, "/api/v1/orders/"+id+"/items",
		map[string]interface{}{"menu_item_id": m.itemID, "quantity": 1})
	e.mustCall(t, http.StatusBadRequest, http.MethodPatch, "/api/v1/orders/"+id, map[string]interface{}{"total_override": 1})

	// custom reason
	id2 := newOrder()
	d = e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+id2+"/refuse",
		map[string]interface{}{"reason": "Catering queue full"}).Data()
	assert.Equal(t, "Catering queue full", d["refuse_reason"])

	e.mustCall(t, http.StatusNotFound, http.MethodPost, "/api/v1/orders/does-not-exist/accept", nil)
}
