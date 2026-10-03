package order_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/services/order"
)

func (e *tenantEnv) setCookPhone(t *testing.T, phone string) {
	t.Helper()
	require.NoError(t, e.notifRepo.UpdateCookAdminPhone(context.Background(), e.tenantID, phone))
}

func (e *tenantEnv) outbox(t *testing.T, orderID string) map[string]models.NotificationOutbox {
	t.Helper()
	rows, err := e.notifRepo.ListByOrder(context.Background(), e.tenantID, orderID)
	require.NoError(t, err)
	out := make(map[string]models.NotificationOutbox, len(rows))
	for _, r := range rows {
		out[r.EventType] = r
	}
	return out
}

// REQNOTIF002S01, REQNOTIF002S02
func TestNotifications_CreateCookAlert(t *testing.T) {
	db := newDB(t)

	t.Run("REQNOTIF002S02_skips_when_cook_phone_unset", func(t *testing.T) {
		e := newTenant(t, db, "notif-unset")
		o := e.createOrder(t) // must still succeed
		assert.Empty(t, e.outbox(t, o.ID))
		assert.Equal(t, models.OrderStatusReceived, e.get(t, o.ID).Status)
	})

	t.Run("REQNOTIF002S01_enqueues_when_cook_phone_set", func(t *testing.T) {
		e := newTenant(t, db, "notif-set")
		e.setCookPhone(t, "555-0199")
		o := e.createOrder(t)
		rows := e.outbox(t, o.ID)
		require.Len(t, rows, 1)
		row := rows[models.EventOrderCreated]
		assert.Equal(t, models.OutboxStatusPending, row.Status)
		assert.Equal(t, "555-0199", row.RecipientPhone, "recipient is the cook, not the customer")
		assert.Equal(t, models.ChannelSMS, row.Channel)
		assert.Equal(t, e.tenantID, row.TenantID)
		assert.Contains(t, row.Body, o.ID)
	})
}

// REQNOTIF002S03, REQNOTIF002S04, REQNOTIF002S05, REQNOTIF002S06
func TestNotifications_LifecycleEnqueueMatrix(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	e := newTenant(t, db, "notif-life")
	// cook phone intentionally unset: customer events are always enqueued.

	t.Run("REQNOTIF002S03_accept_enqueues_to_customer", func(t *testing.T) {
		o := e.createOrder(t)
		_, err := e.svc.Accept(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		rows := e.outbox(t, o.ID)
		require.Len(t, rows, 1)
		row := rows[models.EventOrderAccepted]
		assert.Equal(t, models.OutboxStatusPending, row.Status)
		assert.Equal(t, o.CustomerPhone, row.RecipientPhone)
		assert.Equal(t, "Order "+o.ID+" accepted", row.Body)
	})

	t.Run("REQNOTIF002S04_decline_body_includes_reason", func(t *testing.T) {
		o := e.createOrder(t)
		_, err := e.svc.Refuse(ctx, e.tenantID, o.ID, "  Out of rice  ")
		require.NoError(t, err)
		rows := e.outbox(t, o.ID)
		require.Len(t, rows, 1)
		row := rows[models.EventOrderDeclined]
		assert.Equal(t, o.CustomerPhone, row.RecipientPhone)
		assert.Contains(t, row.Body, "Out of rice")
		assert.Contains(t, row.Body, o.ID)
	})

	t.Run("REQNOTIF002S04_decline_default_reason", func(t *testing.T) {
		o := e.createOrder(t)
		_, err := e.svc.Refuse(ctx, e.tenantID, o.ID, "")
		require.NoError(t, err)
		assert.Contains(t, e.outbox(t, o.ID)[models.EventOrderDeclined].Body, models.DefaultRefuseReason)
	})

	t.Run("REQNOTIF002S05_S06_preparing_no_enqueue_ready_and_pickup_enqueue", func(t *testing.T) {
		o := e.createOrder(t)
		e.addRice(t, o.ID, 1)

		_, err := e.svc.Accept(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		require.Len(t, e.outbox(t, o.ID), 1)

		_, err = e.svc.StartPreparing(ctx, e.tenantID, o.ID) // REQNOTIF002S06
		require.NoError(t, err)
		require.Len(t, e.outbox(t, o.ID), 1, "start-preparing must not enqueue")

		_, err = e.svc.Ready(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		rows := e.outbox(t, o.ID)
		require.Len(t, rows, 2)
		assert.Equal(t, o.CustomerPhone, rows[models.EventOrderReady].RecipientPhone)

		_, err = e.svc.Pickup(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		rows = e.outbox(t, o.ID)
		require.Len(t, rows, 3)
		assert.Equal(t, o.CustomerPhone, rows[models.EventOrderPickedUp].RecipientPhone)
		for _, ev := range []string{models.EventOrderAccepted, models.EventOrderReady, models.EventOrderPickedUp} {
			assert.Equal(t, models.OutboxStatusPending, rows[ev].Status, ev)
		}
	})

	t.Run("rejected_transition_does_not_enqueue", func(t *testing.T) {
		o := e.createOrder(t)
		_, err := e.svc.Ready(ctx, e.tenantID, o.ID) // RECEIVED → READY is illegal
		require.Error(t, err)
		assert.Empty(t, e.outbox(t, o.ID))
	})

	t.Run("picked_up_correction_does_not_enqueue", func(t *testing.T) {
		o := e.createOrder(t)
		e.pickup(t, o.ID)
		before := len(e.outbox(t, o.ID))
		at := int64(1_700_000_000_000)
		_, err := e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{PickedupAt: &at})
		require.NoError(t, err)
		assert.Len(t, e.outbox(t, o.ID), before)
	})
}

// failingBuilder returns an outbox row that violates a DB constraint, forcing the enqueue to fail.
type failingBuilder struct{}

func (failingBuilder) Build(_ context.Context, tenantID, event string, o *models.CustomerOrder) (*models.NotificationOutbox, error) {
	return &models.NotificationOutbox{
		TenantID: tenantID, OrderID: o.ID, EventType: event, Channel: "carrier-pigeon",
		RecipientPhone: "555", Body: "x",
	}, nil
}

type errBuilder struct{}

func (errBuilder) Build(context.Context, string, string, *models.CustomerOrder) (*models.NotificationOutbox, error) {
	return nil, errors.New("settings unavailable")
}

// REQNOTIF002: an outbox insert failure rolls back the order write.
func TestNotifications_SameTransactionRollback(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	e := newTenant(t, db, "notif-tx")
	good := e.createOrder(t)

	bad := order.NewService(e.orderRepo, e.menuSvc, logger.NewLogger(), order.WithNotifications(failingBuilder{}))

	t.Run("create_rolled_back", func(t *testing.T) {
		before, err := bad.List(ctx, e.tenantID)
		require.NoError(t, err)
		_, err = bad.Create(ctx, e.tenantID, &models.OrderCreateRequest{
			MenuID: e.menuID, CustomerName: "Asha", CustomerPhone: "555-0100", ExpectedAt: ptr(int64(1_900_000_000_000)),
		})
		require.Error(t, err)
		after, err := bad.List(ctx, e.tenantID)
		require.NoError(t, err)
		assert.Len(t, after, len(before), "no order row may persist when the outbox insert fails")
	})

	t.Run("accept_rolled_back", func(t *testing.T) {
		_, err := bad.Accept(ctx, e.tenantID, good.ID)
		require.Error(t, err)
		assert.Equal(t, models.OrderStatusReceived, e.get(t, good.ID).Status, "status write must roll back")
		assert.Empty(t, e.outbox(t, good.ID))
	})

	t.Run("refuse_rolled_back", func(t *testing.T) {
		_, err := bad.Refuse(ctx, e.tenantID, good.ID, "nope")
		require.Error(t, err)
		d := e.get(t, good.ID)
		assert.Equal(t, models.OrderStatusReceived, d.Status)
		assert.Nil(t, d.RefuseReason)
	})

	t.Run("builder_error_aborts_write", func(t *testing.T) {
		svc := order.NewService(e.orderRepo, e.menuSvc, logger.NewLogger(), order.WithNotifications(errBuilder{}))
		_, err := svc.Accept(ctx, e.tenantID, good.ID)
		require.Error(t, err)
		assert.Equal(t, models.OrderStatusReceived, e.get(t, good.ID).Status)
	})
}

// Without notification wiring the service behaves as before (no enqueue, no panic).
func TestNotifications_DisabledWhenNotWired(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	e := newTenant(t, db, "notif-off")
	svc := order.NewService(e.orderRepo, e.menuSvc, logger.NewLogger())
	o, err := svc.Create(ctx, e.tenantID, &models.OrderCreateRequest{
		MenuID: e.menuID, CustomerName: "Asha", CustomerPhone: "555", ExpectedAt: ptr(int64(1_900_000_000_000)),
	})
	require.NoError(t, err)
	_, err = svc.Accept(ctx, e.tenantID, o.ID)
	require.NoError(t, err)
	assert.Empty(t, e.outbox(t, o.ID))
}
