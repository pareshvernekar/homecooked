package notification

import (
	"context"
	"time"

	"github.com/pareshvernekar/homecooked/internal/models"
)

// WorkerStore is the outbox persistence the worker needs.
type WorkerStore interface {
	ClaimPending(ctx context.Context, now int64, limit int, lease time.Duration) ([]models.NotificationOutbox, error)
	MarkDelivered(ctx context.Context, tenantID, id, providerRef string, now int64) error
	MarkRetry(ctx context.Context, tenantID, id, lastError string, nextAttemptAt, now int64) error
	MarkDead(ctx context.Context, tenantID, id, lastError string, now int64) error
}

// WorkerConfig tunes delivery. Zero values take the defaults.
type WorkerConfig struct {
	MaxAttempts  int           // total send attempts before dead (default 5)
	BatchSize    int           // rows claimed per cycle (default 20)
	BaseBackoff  time.Duration // delay after the 1st failure, doubled per attempt (default 5s)
	MaxBackoff   time.Duration // backoff cap (default 5m)
	Lease        time.Duration // 'processing' rows older than this are reclaimed (default 5m)
	PollInterval time.Duration // Run loop interval (default 5s)
	SendTimeout  time.Duration // per-send timeout (default 15s)
}

func (c WorkerConfig) withDefaults() WorkerConfig {
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 5
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 20
	}
	if c.BaseBackoff <= 0 {
		c.BaseBackoff = 5 * time.Second
	}
	if c.MaxBackoff <= 0 {
		c.MaxBackoff = 5 * time.Minute
	}
	if c.Lease <= 0 {
		c.Lease = 5 * time.Minute
	}
	if c.PollInterval <= 0 {
		c.PollInterval = 5 * time.Second
	}
	if c.SendTimeout <= 0 {
		c.SendTimeout = 15 * time.Second
	}
	return c
}

// DrainResult summarises one delivery cycle.
type DrainResult struct {
	Claimed, Delivered, Retried, Dead int
}

// Worker claims pending outbox rows and delivers them through an SmsProvider.
// Delivery is at-least-once. REQNOTIF003, REQNOTIF004
type Worker struct {
	store    WorkerStore
	provider SmsProvider
	logger   Logger
	cfg      WorkerConfig
	now      func() time.Time
}

// NewWorker constructs a delivery worker.
func NewWorker(store WorkerStore, provider SmsProvider, l Logger, cfg WorkerConfig) *Worker {
	return &Worker{store: store, provider: provider, logger: l, cfg: cfg.withDefaults(), now: time.Now}
}

// WithClock overrides the clock (tests).
func (w *Worker) WithClock(now func() time.Time) *Worker {
	w.now = now
	return w
}

func (w *Worker) millis() int64 { return w.now().UTC().UnixMilli() }

// backoff returns the delay after the given (1-based) failed attempt count.
func (w *Worker) backoff(attempt int) time.Duration {
	d := w.cfg.BaseBackoff
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= w.cfg.MaxBackoff {
			return w.cfg.MaxBackoff
		}
	}
	if d > w.cfg.MaxBackoff {
		return w.cfg.MaxBackoff
	}
	return d
}

// DrainOnce runs one delivery cycle: claim up to BatchSize due rows, send each, and mark it
// delivered, retry (with backoff), or dead once MaxAttempts is reached.
// REQNOTIF003, REQNOTIF003S01, REQNOTIF003S02
func (w *Worker) DrainOnce(ctx context.Context) (DrainResult, error) {
	var res DrainResult
	rows, err := w.store.ClaimPending(ctx, w.millis(), w.cfg.BatchSize, w.cfg.Lease)
	if err != nil {
		return res, err
	}
	res.Claimed = len(rows)
	for i := range rows {
		row := &rows[i]
		sendCtx, cancel := context.WithTimeout(ctx, w.cfg.SendTimeout)
		ref, sendErr := w.provider.Send(sendCtx, SmsMessage{
			TenantID: row.TenantID, OutboxID: row.ID, To: row.RecipientPhone, Body: row.Body,
		})
		cancel()

		// Persist the outcome even if ctx was cancelled during the send.
		markCtx, markCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		var markErr error
		switch {
		case sendErr == nil:
			markErr = w.store.MarkDelivered(markCtx, row.TenantID, row.ID, ref, w.millis())
			res.Delivered++
		case row.Attempts+1 >= w.cfg.MaxAttempts:
			markErr = w.store.MarkDead(markCtx, row.TenantID, row.ID, truncateErr(sendErr), w.millis())
			res.Dead++
			w.logger.Error(ctx, "notification dead-lettered", "tenant_id", row.TenantID, "outbox_id", row.ID,
				"event_type", row.EventType, "attempts", row.Attempts+1)
		default:
			next := w.now().Add(w.backoff(row.Attempts + 1)).UTC().UnixMilli()
			markErr = w.store.MarkRetry(markCtx, row.TenantID, row.ID, truncateErr(sendErr), next, w.millis())
			res.Retried++
			w.logger.Warn(ctx, "notification send failed; will retry", "tenant_id", row.TenantID,
				"outbox_id", row.ID, "event_type", row.EventType, "attempts", row.Attempts+1)
		}
		markCancel()
		if markErr != nil {
			// Row stays 'processing' and is reclaimed after the lease (at-least-once).
			w.logger.Error(ctx, "failed to record notification outcome", "tenant_id", row.TenantID,
				"outbox_id", row.ID, "error", markErr.Error())
		}
	}
	return res, nil
}

// Run polls DrainOnce every PollInterval until ctx is cancelled.
// REQNOTIF003
func (w *Worker) Run(ctx context.Context) {
	w.logger.Info(ctx, "notification worker started", "poll_interval", w.cfg.PollInterval.String(),
		"max_attempts", w.cfg.MaxAttempts)
	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()
	for {
		for {
			res, err := w.DrainOnce(ctx)
			if err != nil {
				if ctx.Err() == nil {
					w.logger.Error(ctx, "notification drain failed", "error", err.Error())
				}
				break
			}
			if res.Claimed < w.cfg.BatchSize {
				break
			}
		}
		select {
		case <-ctx.Done():
			w.logger.Info(context.WithoutCancel(ctx), "notification worker stopped")
			return
		case <-ticker.C:
		}
	}
}

func truncateErr(err error) string {
	const max = 1000
	s := err.Error()
	if len(s) > max {
		return s[:max]
	}
	return s
}
