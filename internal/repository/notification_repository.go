package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// PostgreSQLNotificationRepository persists the notification outbox, the SMS dev sink, and
// tenant settings. Methods take tenantID explicitly because the delivery worker is cross-tenant.
// REQNOTIF001, REQNOTIF002, REQNOTIF003, REQNOTIF004, REQNOTIF005
type PostgreSQLNotificationRepository struct {
	DB     *sqlx.DB
	Logger *logger.Logger
}

// NewNotificationRepository constructs the outbox repository.
func NewNotificationRepository(db *sqlx.DB, l *logger.Logger) *PostgreSQLNotificationRepository {
	return &PostgreSQLNotificationRepository{DB: db, Logger: l}
}

const outboxSelectCols = `
	id, tenant_id, order_id, event_type, channel, recipient_phone, body, status, attempts,
	next_attempt_at, last_error, provider_ref, created_at, updated_at`

// insertOutboxPending inserts a pending outbox row on the given executor (a *sqlx.Tx when
// called from an order write). It is idempotent per (tenant_id, order_id, event_type).
// REQNOTIF002
func insertOutboxPending(ctx context.Context, ext sqlx.ExtContext, ob *models.NotificationOutbox, tenantID string, now int64) error {
	if ob.ID == "" {
		ob.ID = uuid.New().String()
	}
	ob.TenantID = tenantID
	if ob.Channel == "" {
		ob.Channel = models.ChannelSMS
	}
	ob.Status = models.OutboxStatusPending
	ob.Attempts = 0
	ob.NextAttemptAt = now
	ob.CreatedAt = now
	ob.UpdatedAt = now
	_, err := ext.ExecContext(ctx, `
		INSERT INTO notification_outbox (
			id, tenant_id, order_id, event_type, channel, recipient_phone, body,
			status, attempts, next_attempt_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$11)
		ON CONFLICT (tenant_id, order_id, event_type) DO NOTHING`,
		ob.ID, ob.TenantID, ob.OrderID, ob.EventType, ob.Channel, ob.RecipientPhone, ob.Body,
		ob.Status, ob.Attempts, ob.NextAttemptAt, now)
	if err != nil {
		return fmt.Errorf("insert notification outbox: %w", err)
	}
	return nil
}

// InsertPending enqueues a pending row outside an order transaction (tests, backfills).
// Order writes use InsertPendingTx via CreateOrder/UpdateOrder instead.
// REQNOTIF002
func (r *PostgreSQLNotificationRepository) InsertPending(ctx context.Context, tenantID string, ob *models.NotificationOutbox) error {
	return insertOutboxPending(ctx, r.DB, ob, tenantID, time.Now().UTC().UnixMilli())
}

// InsertPendingTx enqueues a pending row inside the caller's transaction.
// REQNOTIF002
func (r *PostgreSQLNotificationRepository) InsertPendingTx(ctx context.Context, tx *sqlx.Tx, tenantID string, ob *models.NotificationOutbox) error {
	return insertOutboxPending(ctx, tx, ob, tenantID, time.Now().UTC().UnixMilli())
}

// ClaimPending atomically moves up to limit due rows to 'processing' and returns them.
// Due rows are pending with next_attempt_at <= now, plus 'processing' rows whose lease
// (updated_at) is older than lease (a crashed worker), preserving at-least-once delivery.
// Uses FOR UPDATE SKIP LOCKED so concurrent workers never claim the same row.
// REQNOTIF003
func (r *PostgreSQLNotificationRepository) ClaimPending(ctx context.Context, now int64, limit int, lease time.Duration) ([]models.NotificationOutbox, error) {
	if limit <= 0 {
		return nil, nil
	}
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var rows []models.NotificationOutbox
	if err := tx.SelectContext(ctx, &rows, `
		SELECT `+outboxSelectCols+` FROM notification_outbox
		WHERE (status = 'pending' AND next_attempt_at <= $1)
		   OR (status = 'processing' AND updated_at <= $2)
		ORDER BY next_attempt_at ASC, created_at ASC, id ASC
		LIMIT $3
		FOR UPDATE SKIP LOCKED`, now, now-lease.Milliseconds(), limit); err != nil {
		return nil, err
	}
	for i := range rows {
		if _, err := tx.ExecContext(ctx, `
			UPDATE notification_outbox SET status = 'processing', updated_at = $1
			WHERE tenant_id = $2 AND id = $3`, now, rows[i].TenantID, rows[i].ID); err != nil {
			return nil, err
		}
		rows[i].Status = models.OutboxStatusProcessing
		rows[i].UpdatedAt = now
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgreSQLNotificationRepository) markOne(ctx context.Context, query string, args ...interface{}) error {
	res, err := r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// MarkDelivered records a successful send. REQNOTIF003
func (r *PostgreSQLNotificationRepository) MarkDelivered(ctx context.Context, tenantID, id, providerRef string, now int64) error {
	return r.markOne(ctx, `
		UPDATE notification_outbox
		SET status = 'delivered', attempts = attempts + 1, provider_ref = $1, last_error = NULL, updated_at = $2
		WHERE tenant_id = $3 AND id = $4`, providerRef, now, tenantID, id)
}

// MarkRetry records a failed attempt and schedules the next one. REQNOTIF003
func (r *PostgreSQLNotificationRepository) MarkRetry(ctx context.Context, tenantID, id, lastError string, nextAttemptAt, now int64) error {
	return r.markOne(ctx, `
		UPDATE notification_outbox
		SET status = 'pending', attempts = attempts + 1, last_error = $1, next_attempt_at = $2, updated_at = $3
		WHERE tenant_id = $4 AND id = $5`, lastError, nextAttemptAt, now, tenantID, id)
}

// MarkDead records the final failed attempt; the row is never retried again. REQNOTIF003
func (r *PostgreSQLNotificationRepository) MarkDead(ctx context.Context, tenantID, id, lastError string, now int64) error {
	return r.markOne(ctx, `
		UPDATE notification_outbox
		SET status = 'dead', attempts = attempts + 1, last_error = $1, updated_at = $2
		WHERE tenant_id = $3 AND id = $4`, lastError, now, tenantID, id)
}

// ListByOrder lists a tenant's outbox rows for an order, oldest first. REQNOTIF005
func (r *PostgreSQLNotificationRepository) ListByOrder(ctx context.Context, tenantID, orderID string) ([]models.NotificationOutbox, error) {
	rows := []models.NotificationOutbox{}
	err := r.DB.SelectContext(ctx, &rows, `
		SELECT `+outboxSelectCols+` FROM notification_outbox
		WHERE tenant_id = $1 AND order_id = $2
		ORDER BY created_at ASC, id ASC`, tenantID, orderID)
	return rows, err
}

// OrderExists reports whether an active order exists for the tenant. REQNOTIF005
func (r *PostgreSQLNotificationRepository) OrderExists(ctx context.Context, tenantID, orderID string) (bool, error) {
	var ok bool
	err := r.DB.GetContext(ctx, &ok, `
		SELECT EXISTS(SELECT 1 FROM customer_order WHERE tenant_id = $1 AND id = $2 AND is_active = TRUE)`,
		tenantID, orderID)
	return ok, err
}

// GetCookAdminPhone returns the tenant's cook admin phone ("" when unset). REQNOTIF001
// Returns sql.ErrNoRows when the tenant does not exist or is inactive.
func (r *PostgreSQLNotificationRepository) GetCookAdminPhone(ctx context.Context, tenantID string) (string, error) {
	var phone sql.NullString
	if err := r.DB.GetContext(ctx, &phone, `
		SELECT cook_admin_phone FROM tenant WHERE id = $1 AND is_active = TRUE`, tenantID); err != nil {
		return "", err
	}
	return strings.TrimSpace(phone.String), nil
}

// UpdateCookAdminPhone sets (non-empty) or clears (empty) the tenant cook admin phone. REQNOTIF001
// Returns sql.ErrNoRows when the tenant does not exist or is inactive.
func (r *PostgreSQLNotificationRepository) UpdateCookAdminPhone(ctx context.Context, tenantID, phone string) error {
	var v interface{}
	if p := strings.TrimSpace(phone); p != "" {
		v = p
	}
	return r.markOne(ctx, `
		UPDATE tenant SET cook_admin_phone = $1, updated_at = $2
		WHERE id = $3 AND is_active = TRUE`, v, time.Now().UTC().UnixMilli(), tenantID)
}

// RecordSmsSink writes a message captured by the local SMS provider. REQNOTIF004
func (r *PostgreSQLNotificationRepository) RecordSmsSink(ctx context.Context, tenantID, outboxID, phone, body string) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sms_dev_sink (id, tenant_id, outbox_id, phone, body, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		uuid.New().String(), tenantID, outboxID, phone, body, time.Now().UTC().UnixMilli())
	return err
}

// ListSmsSink lists local-provider messages for an outbox row (tests/dev inspection). REQNOTIF004
func (r *PostgreSQLNotificationRepository) ListSmsSink(ctx context.Context, tenantID, outboxID string) ([]models.SmsDevSinkEntry, error) {
	rows := []models.SmsDevSinkEntry{}
	err := r.DB.SelectContext(ctx, &rows, `
		SELECT id, tenant_id, outbox_id, phone, body, created_at FROM sms_dev_sink
		WHERE tenant_id = $1 AND outbox_id = $2
		ORDER BY created_at ASC, id ASC`, tenantID, outboxID)
	return rows, err
}
