package db_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testdb "github.com/pareshvernekar/homecooked/tests/db"
)

// REQNOTIF001, REQNOTIF002, REQNOTIF003, REQNOTIF004
func TestNotificationSchemaConstraints(t *testing.T) {
	ctx := t.Context()
	helper, err := testdb.NewDatabaseHelper(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = helper.Terminate(ctx) })
	require.NoError(t, testdb.InitializeSchema(ctx, helper.DB))

	now := time.Now().UTC().UnixMilli()
	const tenantA, tenantB = "notif-tenant-a", "notif-tenant-b"
	for _, id := range []string{tenantA, tenantB} {
		_, err = helper.DB.ExecContext(ctx, `
			INSERT INTO tenant (id, name, description, is_active, created_at, updated_at)
			VALUES ($1, $1, '', TRUE, $2, $2)`, id, now)
		require.NoError(t, err)
	}

	insertOrder := func(tenantID string) string {
		_, menuID, _, _, _ := seedMenuScaffold(t, helper, tenantID, now)
		id := uuid.New().String()
		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO customer_order
			(id, tenant_id, menu_id, customer_name, customer_phone, received_at, expected_at, status, created_at, updated_at)
			VALUES ($1,$2,$3,'Asha','555',$4,$4,'RECEIVED',$4,$4)`, id, tenantID, menuID, now)
		require.NoError(t, err)
		return id
	}
	orderA, orderB := insertOrder(tenantA), insertOrder(tenantB)

	insertOutbox := func(tenantID, orderID, event, channel, phone, status string, attempts int) (string, error) {
		id := uuid.New().String()
		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO notification_outbox
			(id, tenant_id, order_id, event_type, channel, recipient_phone, body, status, attempts, next_attempt_at, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,'hello',$7,$8,$9,$9,$9)`,
			id, tenantID, orderID, event, channel, phone, status, attempts, now)
		return id, err
	}

	t.Run("REQNOTIF001_tenant_cook_admin_phone_nullable", func(t *testing.T) {
		var phone *string
		require.NoError(t, helper.DB.GetContext(ctx, &phone, `SELECT cook_admin_phone FROM tenant WHERE id = $1`, tenantA))
		assert.Nil(t, phone, "unset by default")
		_, err := helper.DB.ExecContext(ctx, `UPDATE tenant SET cook_admin_phone = '555-0199' WHERE id = $1`, tenantA)
		require.NoError(t, err)
		_, err = helper.DB.ExecContext(ctx, `UPDATE tenant SET cook_admin_phone = NULL WHERE id = $1`, tenantA)
		require.NoError(t, err)
	})

	t.Run("REQNOTIF003_status_values", func(t *testing.T) {
		for i, st := range []string{"pending", "processing", "delivered", "dead"} {
			_, err := insertOutbox(tenantA, orderA, "evt."+st+string(rune('a'+i)), "sms", "555", st, 0)
			require.NoError(t, err, st)
		}
		_, err := insertOutbox(tenantA, orderA, "evt.bad", "sms", "555", "sent", 0)
		require.Error(t, err, "unknown status must fail")
		_, err = insertOutbox(tenantA, orderA, "evt.bad2", "sms", "555", "PENDING", 0)
		require.Error(t, err, "status is lower-case only")
	})

	t.Run("REQNOTIF002_channel_and_recipient_checks", func(t *testing.T) {
		_, err := insertOutbox(tenantA, orderA, "evt.chan", "email", "555", "pending", 0)
		require.Error(t, err, "only sms channel in v1")
		_, err = insertOutbox(tenantA, orderA, "evt.phone", "sms", "", "pending", 0)
		require.Error(t, err, "recipient phone must be non-empty")
		_, err = insertOutbox(tenantA, orderA, "", "sms", "555", "pending", 0)
		require.Error(t, err, "event type must be non-empty")
	})

	t.Run("REQNOTIF003_attempts_non_negative", func(t *testing.T) {
		_, err := insertOutbox(tenantA, orderA, "evt.attempts", "sms", "555", "pending", -1)
		require.Error(t, err)
	})

	t.Run("REQNOTIF002_unique_event_per_order", func(t *testing.T) {
		_, err := insertOutbox(tenantA, orderA, "order.created", "sms", "555", "pending", 0)
		require.NoError(t, err)
		_, err = insertOutbox(tenantA, orderA, "order.created", "sms", "555", "pending", 0)
		require.Error(t, err, "duplicate (tenant, order, event) must fail")
		_, err = insertOutbox(tenantA, orderA, "order.accepted", "sms", "555", "pending", 0)
		require.NoError(t, err, "different event for same order is allowed")
	})

	t.Run("REQNOTIF002_order_fk_is_tenant_scoped", func(t *testing.T) {
		_, err := insertOutbox(tenantA, uuid.New().String(), "order.created", "sms", "555", "pending", 0)
		require.Error(t, err, "unknown order must fail")
		_, err = insertOutbox(tenantA, orderB, "order.created", "sms", "555", "pending", 0)
		require.Error(t, err, "tenant A outbox must not reference tenant B order")
		_, err = insertOutbox("no-such-tenant", orderA, "order.created", "sms", "555", "pending", 0)
		require.Error(t, err, "unknown tenant must fail")
		_, err = insertOutbox(tenantB, orderB, "order.created", "sms", "555", "pending", 0)
		require.NoError(t, err)
	})

	t.Run("REQNOTIF004_sms_dev_sink_fk_and_cascade", func(t *testing.T) {
		outboxID, err := insertOutbox(tenantA, orderA, "order.ready", "sms", "555", "delivered", 1)
		require.NoError(t, err)
		_, err = helper.DB.ExecContext(ctx, `
			INSERT INTO sms_dev_sink (id, tenant_id, outbox_id, phone, body, created_at)
			VALUES ($1,$2,$3,'555','hello',$4)`, uuid.New().String(), tenantA, outboxID, now)
		require.NoError(t, err)
		_, err = helper.DB.ExecContext(ctx, `
			INSERT INTO sms_dev_sink (id, tenant_id, outbox_id, phone, body, created_at)
			VALUES ($1,$2,$3,'555','hello',$4)`, uuid.New().String(), tenantB, outboxID, now)
		require.Error(t, err, "sink row must reference an outbox row of the same tenant")

		_, err = helper.DB.ExecContext(ctx, `DELETE FROM notification_outbox WHERE tenant_id = $1 AND id = $2`, tenantA, outboxID)
		require.NoError(t, err)
		var n int
		require.NoError(t, helper.DB.GetContext(ctx, &n, `SELECT COUNT(*) FROM sms_dev_sink WHERE outbox_id = $1`, outboxID))
		assert.Equal(t, 0, n, "sink rows cascade with the outbox row")
	})

	t.Run("REQNOTIF002_outbox_cascades_with_order", func(t *testing.T) {
		o := insertOrder(tenantB)
		_, err := insertOutbox(tenantB, o, "order.created", "sms", "555", "pending", 0)
		require.NoError(t, err)
		_, err = helper.DB.ExecContext(ctx, `DELETE FROM customer_order WHERE tenant_id = $1 AND id = $2`, tenantB, o)
		require.NoError(t, err)
		var n int
		require.NoError(t, helper.DB.GetContext(ctx, &n, `SELECT COUNT(*) FROM notification_outbox WHERE order_id = $1`, o))
		assert.Equal(t, 0, n)
	})
}
