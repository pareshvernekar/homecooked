package orders

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pareshvernekar/homecooked/internal/services/notification"
)

func (e *env) notifications(t *testing.T, orderID string) []map[string]interface{} {
	t.Helper()
	list := e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/orders/"+orderID+"/notifications", nil).List()
	out := make([]map[string]interface{}, 0, len(list))
	for _, r := range list {
		out = append(out, r.(map[string]interface{}))
	}
	return out
}

func byEvent(rows []map[string]interface{}, event string) map[string]interface{} {
	for _, r := range rows {
		if r["event_type"] == event {
			return r
		}
	}
	return nil
}

// REQNOTIF001–REQNOTIF005 over HTTP.
func TestOrder_Notifications(t *testing.T) {
	e := setupEnv(t)
	m := seedPublishedDailyMenu(t, e, "daily")
	newOrder := func() string {
		return e.mustCall(t, http.StatusCreated, http.MethodPost, "/api/v1/orders", map[string]interface{}{
			"menu_id": m.menuID, "customer_name": "Asha", "customer_phone": "555-0100",
			"expected_at": time.Now().Add(3 * time.Hour).UnixMilli(),
		}).Data()["id"].(string)
	}

	// REQNOTIF001S01 / REQNOTIF002S02: unset by default; create succeeds without an order.created row.
	settings := e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/tenant/settings", nil).Data()
	assert.Nil(t, settings["cook_admin_phone"])
	assert.NotEmpty(t, settings["warnings"])
	noCook := newOrder()
	assert.Empty(t, e.notifications(t, noCook))

	// REQNOTIF001S01
	put := e.mustCall(t, http.StatusOK, http.MethodPut, "/api/v1/tenant/settings", map[string]interface{}{"cook_admin_phone": "555-0199"}).Data()
	assert.Equal(t, "555-0199", put["cook_admin_phone"])
	settings = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/tenant/settings", nil).Data()
	assert.Equal(t, "555-0199", settings["cook_admin_phone"])

	// Validation: missing field is rejected; invalid JSON is rejected.
	e.mustCall(t, http.StatusBadRequest, http.MethodPut, "/api/v1/tenant/settings", `{}`)
	e.mustCall(t, http.StatusBadRequest, http.MethodPut, "/api/v1/tenant/settings", `{"cook_admin_phone": 5}`)

	// REQNOTIF002S01: create enqueues to the cook.
	orderID := newOrder()
	rows := e.notifications(t, orderID)
	require.Len(t, rows, 1)
	assert.Equal(t, "555-0199", rows[0]["recipient_phone"])
	assert.Equal(t, "pending", rows[0]["status"])
	assert.Equal(t, "sms", rows[0]["channel"])

	// REQNOTIF002S03, S06, S05: accept → (preparing: none) → ready → pickup.
	e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+orderID+"/accept", nil)
	e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+orderID+"/start-preparing", nil)
	require.Len(t, e.notifications(t, orderID), 2)
	e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+orderID+"/ready", nil)
	e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+orderID+"/pickup", nil)
	rows = e.notifications(t, orderID)
	require.Len(t, rows, 4)
	for _, ev := range []string{"order.accepted", "order.ready", "order.picked_up"} {
		r := byEvent(rows, ev)
		require.NotNil(t, r, ev)
		assert.Equal(t, "555-0100", r["recipient_phone"], ev)
	}
	assert.Nil(t, byEvent(rows, "order.in_progress"))

	// REQNOTIF002S04: decline carries the reason.
	declined := newOrder()
	e.mustCall(t, http.StatusOK, http.MethodPost, "/api/v1/orders/"+declined+"/refuse", map[string]interface{}{"reason": "Out of rice"})
	r := byEvent(e.notifications(t, declined), "order.declined")
	require.NotNil(t, r)
	assert.Contains(t, r["body"], "Out of rice")

	// REQNOTIF003S01 / REQNOTIF004S01: a worker cycle delivers via the local provider.
	w := notification.NewWorker(e.notifRepo, notification.NewLocalSmsProvider(e.notifRepo, e.logger), e.logger, notification.WorkerConfig{})
	res, err := w.DrainOnce(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 6, res.Delivered) // main order: created+accepted+ready+picked_up; declined order: created+declined
	for _, row := range e.notifications(t, orderID) {
		assert.Equal(t, "delivered", row["status"])
	}
	sink, err := e.notifRepo.ListSmsSink(t.Context(), testTenantID, byEvent(e.notifications(t, declined), "order.declined")["id"].(string))
	require.NoError(t, err)
	require.Len(t, sink, 1)

	// REQNOTIF005S01: cross-tenant / unknown order never exposes rows.
	e.mustCall(t, http.StatusNotFound, http.MethodGet, "/api/v1/orders/does-not-exist/notifications", nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+orderID+"/notifications", strings.NewReader(""))
	req.Header.Set("X-Tenant-ID", "another-tenant")
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// REQNOTIF001S02: clear.
	cleared := e.mustCall(t, http.StatusOK, http.MethodPut, "/api/v1/tenant/settings", `{"cook_admin_phone": null}`).Data()
	assert.Nil(t, cleared["cook_admin_phone"])
	e.mustCall(t, http.StatusOK, http.MethodPut, "/api/v1/tenant/settings", map[string]interface{}{"cook_admin_phone": "555-0199"})
	e.mustCall(t, http.StatusOK, http.MethodPut, "/api/v1/tenant/settings", map[string]interface{}{"cook_admin_phone": ""})
	settings = e.mustCall(t, http.StatusOK, http.MethodGet, "/api/v1/tenant/settings", nil).Data()
	assert.Nil(t, settings["cook_admin_phone"])
}
