package notification_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/pareshvernekar/homecooked/internal/errors"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/services/notification"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
)

// --- unit tests (no DB) ---

type fakeCook struct {
	phone string
	err   error
}

func (f fakeCook) GetCookAdminPhone(context.Context, string) (string, error) { return f.phone, f.err }

// REQNOTIF002
func TestBuilder_Events(t *testing.T) {
	ctx := context.Background()
	reason := "Out of rice"
	o := &models.CustomerOrder{ID: "o-1", CustomerName: "Asha", CustomerPhone: "555-0100", RefuseReason: &reason}
	b := notification.NewBuilder(fakeCook{phone: "555-0199"}, logger.NewLogger())

	t.Run("REQNOTIF002S01_created_goes_to_cook", func(t *testing.T) {
		ob, err := b.Build(ctx, "t1", models.EventOrderCreated, o)
		require.NoError(t, err)
		require.NotNil(t, ob)
		assert.Equal(t, "555-0199", ob.RecipientPhone)
		assert.Equal(t, "o-1", ob.OrderID)
		assert.Equal(t, models.ChannelSMS, ob.Channel)
	})

	t.Run("REQNOTIF002S02_created_skipped_without_cook_phone", func(t *testing.T) {
		ob, err := notification.NewBuilder(fakeCook{}, logger.NewLogger()).Build(ctx, "t1", models.EventOrderCreated, o)
		require.NoError(t, err)
		assert.Nil(t, ob)
	})

	t.Run("cook_lookup_error_is_returned", func(t *testing.T) {
		_, err := notification.NewBuilder(fakeCook{err: errors.New("boom")}, logger.NewLogger()).Build(ctx, "t1", models.EventOrderCreated, o)
		require.Error(t, err)
	})

	t.Run("REQNOTIF002S03_S04_S05_customer_events", func(t *testing.T) {
		for _, ev := range []string{models.EventOrderAccepted, models.EventOrderDeclined, models.EventOrderReady, models.EventOrderPickedUp} {
			ob, err := b.Build(ctx, "t1", ev, o)
			require.NoError(t, err, ev)
			require.NotNil(t, ob, ev)
			assert.Equal(t, "555-0100", ob.RecipientPhone, ev)
			assert.Contains(t, ob.Body, "o-1", ev)
		}
		ob, _ := b.Build(ctx, "t1", models.EventOrderDeclined, o)
		assert.Contains(t, ob.Body, reason)
	})

	t.Run("unsupported_event_rejected", func(t *testing.T) {
		_, err := b.Build(ctx, "t1", "order.preparing", o)
		require.Error(t, err)
	})
}

// REQNOTIF004S01
func TestNewProvider_SelectedByConfig(t *testing.T) {
	for _, name := range []string{"", "local", " LOCAL "} {
		p, err := notification.NewProvider(name, nil, logger.NewLogger())
		require.NoError(t, err, name)
		assert.IsType(t, &notification.LocalSmsProvider{}, p)
	}
	_, err := notification.NewProvider("twilio", nil, logger.NewLogger())
	require.ErrorIs(t, err, notification.ErrUnknownProvider)
}

func TestLocalSmsProvider_RejectsEmptyRecipient(t *testing.T) {
	_, err := notification.NewLocalSmsProvider(nil, logger.NewLogger()).Send(context.Background(), notification.SmsMessage{Body: "x"})
	require.Error(t, err)
}

// --- DB-backed tests ---

type fixture struct {
	db   *sqlx.DB
	repo *repository.PostgreSQLNotificationRepository
	now  int64
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	helper, err := testdb.NewDatabaseHelper(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = helper.Terminate(ctx) })
	require.NoError(t, testdb.InitializeSchema(ctx, helper.DB))
	return &fixture{db: helper.DB, repo: repository.NewNotificationRepository(helper.DB, logger.NewLogger()), now: time.Now().UTC().UnixMilli()}
}

// seedTenant inserts an active tenant.
func (f *fixture) seedTenant(t *testing.T, id string) {
	t.Helper()
	_, err := f.db.Exec(`INSERT INTO tenant (id, name, is_active, created_at, updated_at) VALUES ($1,$1,TRUE,$2,$2)`, id, f.now)
	require.NoError(t, err)
}

// seedOrder inserts a minimal menu + order for the tenant and returns the order id.
func (f *fixture) seedOrder(t *testing.T, tenantID string) string {
	t.Helper()
	menuID, orderID := uuid.New().String(), uuid.New().String()
	_, err := f.db.Exec(`
		INSERT INTO menu (id, tenant_id, name, menu_type, status, is_active, menu_date, created_at, updated_at)
		VALUES ($1,$2,'Daily','daily','draft',TRUE,CURRENT_DATE,$3,$3)`, menuID, tenantID, f.now)
	require.NoError(t, err)
	_, err = f.db.Exec(`
		INSERT INTO customer_order (id, tenant_id, menu_id, customer_name, customer_phone, received_at, expected_at, status, created_at, updated_at)
		VALUES ($1,$2,$3,'Asha','555-0100',$4,$4,'RECEIVED',$4,$4)`, orderID, tenantID, menuID, f.now)
	require.NoError(t, err)
	return orderID
}

func (f *fixture) enqueue(t *testing.T, tenantID, orderID, event string) *models.NotificationOutbox {
	t.Helper()
	ob := &models.NotificationOutbox{OrderID: orderID, EventType: event, RecipientPhone: "555-0100", Body: "Order " + orderID + " " + event}
	require.NoError(t, f.repo.InsertPending(context.Background(), tenantID, ob))
	return ob
}

func (f *fixture) row(t *testing.T, tenantID, orderID, event string) models.NotificationOutbox {
	t.Helper()
	rows, err := f.repo.ListByOrder(context.Background(), tenantID, orderID)
	require.NoError(t, err)
	for _, r := range rows {
		if r.EventType == event {
			return r
		}
	}
	t.Fatalf("no outbox row for %s", event)
	return models.NotificationOutbox{}
}

type countingProvider struct{ sends atomic.Int64 }

func (c *countingProvider) Send(context.Context, notification.SmsMessage) (string, error) {
	c.sends.Add(1)
	return "ref", nil
}

// REQNOTIF003, REQNOTIF004
func TestWorker_Delivery(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	l := logger.NewLogger()
	f.seedTenant(t, "w-tenant")

	t.Run("REQNOTIF003S01_REQNOTIF004S01_local_provider_delivers", func(t *testing.T) {
		orderID := f.seedOrder(t, "w-tenant")
		ob := f.enqueue(t, "w-tenant", orderID, models.EventOrderAccepted)

		w := notification.NewWorker(f.repo, notification.NewLocalSmsProvider(f.repo, l), l, notification.WorkerConfig{})
		res, err := w.DrainOnce(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, res.Claimed)
		assert.Equal(t, 1, res.Delivered)

		row := f.row(t, "w-tenant", orderID, models.EventOrderAccepted)
		assert.Equal(t, models.OutboxStatusDelivered, row.Status)
		assert.Equal(t, 1, row.Attempts)
		require.NotNil(t, row.ProviderRef)
		assert.Contains(t, *row.ProviderRef, "local-")

		sink, err := f.repo.ListSmsSink(ctx, "w-tenant", ob.ID)
		require.NoError(t, err)
		require.Len(t, sink, 1, "local provider records the message")
		assert.Equal(t, "555-0100", sink[0].Phone)
		assert.Equal(t, ob.Body, sink[0].Body)

		res, err = w.DrainOnce(ctx) // delivered rows are never re-sent
		require.NoError(t, err)
		assert.Equal(t, 0, res.Claimed)
	})

	t.Run("REQNOTIF003S02_failure_retries_with_backoff_then_dead", func(t *testing.T) {
		orderID := f.seedOrder(t, "w-tenant")
		f.enqueue(t, "w-tenant", orderID, models.EventOrderDeclined)

		clock := time.Now()
		fail := &notification.FailSmsProvider{Err: errors.New("carrier down")}
		w := notification.NewWorker(f.repo, fail, l, notification.WorkerConfig{MaxAttempts: 3, BaseBackoff: time.Minute}).
			WithClock(func() time.Time { return clock })

		res, err := w.DrainOnce(ctx) // attempt 1
		require.NoError(t, err)
		assert.Equal(t, notification.DrainResult{Claimed: 1, Retried: 1}, res)
		row := f.row(t, "w-tenant", orderID, models.EventOrderDeclined)
		assert.Equal(t, models.OutboxStatusPending, row.Status)
		assert.Equal(t, 1, row.Attempts)
		require.NotNil(t, row.LastError)
		assert.Contains(t, *row.LastError, "carrier down")
		assert.Equal(t, clock.Add(time.Minute).UnixMilli(), row.NextAttemptAt, "1st backoff = base")

		res, err = w.DrainOnce(ctx) // not due yet
		require.NoError(t, err)
		assert.Equal(t, 0, res.Claimed, "row must wait for next_attempt_at")

		clock = clock.Add(time.Minute)
		res, err = w.DrainOnce(ctx) // attempt 2
		require.NoError(t, err)
		assert.Equal(t, 1, res.Retried)
		row = f.row(t, "w-tenant", orderID, models.EventOrderDeclined)
		assert.Equal(t, 2, row.Attempts)
		assert.Equal(t, clock.Add(2*time.Minute).UnixMilli(), row.NextAttemptAt, "backoff doubles")

		clock = clock.Add(2 * time.Minute)
		res, err = w.DrainOnce(ctx) // attempt 3 = max
		require.NoError(t, err)
		assert.Equal(t, notification.DrainResult{Claimed: 1, Dead: 1}, res)
		row = f.row(t, "w-tenant", orderID, models.EventOrderDeclined)
		assert.Equal(t, models.OutboxStatusDead, row.Status)
		assert.Equal(t, 3, row.Attempts)
		require.NotNil(t, row.LastError)
		assert.Contains(t, *row.LastError, "carrier down")

		clock = clock.Add(24 * time.Hour)
		res, err = w.DrainOnce(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, res.Claimed, "dead rows are never retried")
		assert.Equal(t, 3, fail.Calls)
	})

	t.Run("REQNOTIF003_retry_then_success", func(t *testing.T) {
		orderID := f.seedOrder(t, "w-tenant")
		f.enqueue(t, "w-tenant", orderID, models.EventOrderReady)
		clock := time.Now()
		flaky := &flakyProvider{failFirst: 1}
		w := notification.NewWorker(f.repo, flaky, l, notification.WorkerConfig{BaseBackoff: time.Second}).
			WithClock(func() time.Time { return clock })
		_, err := w.DrainOnce(ctx)
		require.NoError(t, err)
		clock = clock.Add(time.Second)
		res, err := w.DrainOnce(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, res.Delivered)
		row := f.row(t, "w-tenant", orderID, models.EventOrderReady)
		assert.Equal(t, models.OutboxStatusDelivered, row.Status)
		assert.Equal(t, 2, row.Attempts)
		assert.Nil(t, row.LastError, "last_error cleared on success")
	})

	t.Run("REQNOTIF003_stale_processing_row_is_reclaimed", func(t *testing.T) {
		orderID := f.seedOrder(t, "w-tenant")
		f.enqueue(t, "w-tenant", orderID, models.EventOrderPickedUp)
		// Simulate a worker that crashed after claiming: processing with an old lease.
		_, err := f.db.Exec(`UPDATE notification_outbox SET status='processing', updated_at = $1 WHERE order_id = $2`,
			time.Now().Add(-time.Hour).UnixMilli(), orderID)
		require.NoError(t, err)
		p := &countingProvider{}
		w := notification.NewWorker(f.repo, p, l, notification.WorkerConfig{Lease: time.Minute})
		res, err := w.DrainOnce(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, res.Delivered)
		assert.Equal(t, models.OutboxStatusDelivered, f.row(t, "w-tenant", orderID, models.EventOrderPickedUp).Status)

		// A fresh 'processing' row (live lease) is left alone.
		orderID2 := f.seedOrder(t, "w-tenant")
		f.enqueue(t, "w-tenant", orderID2, models.EventOrderPickedUp)
		_, err = f.db.Exec(`UPDATE notification_outbox SET status='processing', updated_at = $1 WHERE order_id = $2`,
			time.Now().UnixMilli(), orderID2)
		require.NoError(t, err)
		res, err = w.DrainOnce(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, res.Claimed)
	})

	t.Run("REQNOTIF003_concurrent_workers_never_double_send", func(t *testing.T) {
		const n = 12
		for i := 0; i < n; i++ {
			f.enqueue(t, "w-tenant", f.seedOrder(t, "w-tenant"), models.EventOrderAccepted)
		}
		p := &countingProvider{}
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w := notification.NewWorker(f.repo, p, l, notification.WorkerConfig{BatchSize: 4})
				for {
					res, err := w.DrainOnce(ctx)
					if err != nil || res.Claimed == 0 {
						return
					}
				}
			}()
		}
		wg.Wait()
		assert.EqualValues(t, n, p.sends.Load(), "each pending row is sent exactly once across workers")
	})

	t.Run("REQNOTIF002_enqueue_is_idempotent_per_order_event", func(t *testing.T) {
		orderID := f.seedOrder(t, "w-tenant")
		f.enqueue(t, "w-tenant", orderID, models.EventOrderAccepted)
		f.enqueue(t, "w-tenant", orderID, models.EventOrderAccepted)
		rows, err := f.repo.ListByOrder(ctx, "w-tenant", orderID)
		require.NoError(t, err)
		assert.Len(t, rows, 1)
	})
}

type flakyProvider struct {
	failFirst int
	calls     int
}

func (p *flakyProvider) Send(context.Context, notification.SmsMessage) (string, error) {
	p.calls++
	if p.calls <= p.failFirst {
		return "", errors.New("transient")
	}
	return "ok", nil
}

// REQNOTIF001
func TestSettings(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.seedTenant(t, "s-tenant")
	svc := notification.NewService(f.repo)
	ptr := func(s string) *string { return &s }
	set := func(v *string) *models.TenantSettingsUpdateRequest {
		return &models.TenantSettingsUpdateRequest{CookAdminPhone: models.NullableString{Set: true, Value: v}}
	}

	got, err := svc.GetSettings(ctx, "s-tenant")
	require.NoError(t, err)
	assert.Nil(t, got.CookAdminPhone, "unset by default")
	assert.NotEmpty(t, got.Warnings, "warns that cook alerts are skipped")

	t.Run("REQNOTIF001S01_set", func(t *testing.T) {
		got, err := svc.UpdateSettings(ctx, "s-tenant", set(ptr("  555-0199 ")))
		require.NoError(t, err)
		require.NotNil(t, got.CookAdminPhone)
		assert.Equal(t, "555-0199", *got.CookAdminPhone)
		assert.Empty(t, got.Warnings)

		got, err = svc.GetSettings(ctx, "s-tenant")
		require.NoError(t, err)
		require.NotNil(t, got.CookAdminPhone)
		assert.Equal(t, "555-0199", *got.CookAdminPhone)
	})

	t.Run("REQNOTIF001S02_clear_with_null_or_empty", func(t *testing.T) {
		for _, v := range []*string{nil, ptr(""), ptr("   ")} {
			_, err := svc.UpdateSettings(ctx, "s-tenant", set(ptr("555-0199")))
			require.NoError(t, err)
			_, err = svc.UpdateSettings(ctx, "s-tenant", set(v))
			require.NoError(t, err)
			got, err := svc.GetSettings(ctx, "s-tenant")
			require.NoError(t, err)
			assert.Nil(t, got.CookAdminPhone)
		}
	})

	t.Run("validation_and_unknown_tenant", func(t *testing.T) {
		_, err := svc.UpdateSettings(ctx, "s-tenant", &models.TenantSettingsUpdateRequest{})
		requireCode(t, err, 400) // field absent
		long := make([]byte, 51)
		for i := range long {
			long[i] = '1'
		}
		_, err = svc.UpdateSettings(ctx, "s-tenant", set(ptr(string(long))))
		requireCode(t, err, 400)
		_, err = svc.GetSettings(ctx, "missing")
		requireCode(t, err, 404)
		_, err = svc.UpdateSettings(ctx, "missing", set(ptr("555")))
		requireCode(t, err, 404)
	})

	t.Run("settings_are_tenant_scoped", func(t *testing.T) {
		f.seedTenant(t, "s-other")
		_, err := svc.UpdateSettings(ctx, "s-tenant", set(ptr("555-0001")))
		require.NoError(t, err)
		got, err := svc.GetSettings(ctx, "s-other")
		require.NoError(t, err)
		assert.Nil(t, got.CookAdminPhone)
	})
}

// REQNOTIF005
func TestListByOrder_TenantScoped(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.seedTenant(t, "l-a")
	f.seedTenant(t, "l-b")
	svc := notification.NewService(f.repo)

	orderA, orderB := f.seedOrder(t, "l-a"), f.seedOrder(t, "l-b")
	f.enqueue(t, "l-a", orderA, models.EventOrderAccepted)
	f.enqueue(t, "l-a", orderA, models.EventOrderReady)
	f.enqueue(t, "l-b", orderB, models.EventOrderAccepted)

	t.Run("REQNOTIF005S01_lists_rows_for_order", func(t *testing.T) {
		rows, err := svc.ListByOrder(ctx, "l-a", orderA)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		for _, r := range rows {
			assert.Equal(t, orderA, r.OrderID)
			assert.Equal(t, "l-a", r.TenantID)
			assert.Equal(t, models.ChannelSMS, r.Channel)
			assert.Equal(t, models.OutboxStatusPending, r.Status)
			assert.NotEmpty(t, r.RecipientPhone)
			assert.NotZero(t, r.CreatedAt)
		}
	})

	t.Run("REQNOTIF005S01_cross_tenant_order_is_not_found", func(t *testing.T) {
		_, err := svc.ListByOrder(ctx, "l-a", orderB)
		requireCode(t, err, 404)
		_, err = svc.ListByOrder(ctx, "l-b", orderA)
		requireCode(t, err, 404)
	})

	t.Run("order_without_notifications_returns_empty_list", func(t *testing.T) {
		rows, err := svc.ListByOrder(ctx, "l-a", f.seedOrder(t, "l-a"))
		require.NoError(t, err)
		assert.NotNil(t, rows)
		assert.Empty(t, rows)
	})
}

func requireCode(t *testing.T, err error, code int) {
	t.Helper()
	require.Error(t, err)
	var se *apperrors.ServiceError
	require.ErrorAs(t, err, &se)
	assert.Equal(t, code, se.GetStatusCode())
}
