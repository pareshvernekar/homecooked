package notification

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Logger is the minimal structured logger used by this package (satisfied by *logger.Logger).
type Logger interface {
	Info(ctx context.Context, msg string, keysAndValues ...interface{})
	Warn(ctx context.Context, msg string, keysAndValues ...interface{})
	Error(ctx context.Context, msg string, keysAndValues ...interface{})
}

// SinkRecorder persists messages captured by the local provider (sms_dev_sink).
type SinkRecorder interface {
	RecordSmsSink(ctx context.Context, tenantID, outboxID, phone, body string) error
}

// LocalSmsProvider delivers to a dev sink table and the structured log; no external vendor.
// REQNOTIF004
type LocalSmsProvider struct {
	sink   SinkRecorder
	logger Logger
}

// NewLocalSmsProvider constructs the local provider. sink may be nil (log only).
func NewLocalSmsProvider(sink SinkRecorder, l Logger) *LocalSmsProvider {
	return &LocalSmsProvider{sink: sink, logger: l}
}

// Send records the message and returns a synthetic provider reference.
// REQNOTIF004
func (p *LocalSmsProvider) Send(ctx context.Context, msg SmsMessage) (string, error) {
	if msg.To == "" {
		return "", errors.New("recipient phone is empty")
	}
	if p.sink != nil {
		if err := p.sink.RecordSmsSink(ctx, msg.TenantID, msg.OutboxID, msg.To, msg.Body); err != nil {
			return "", fmt.Errorf("record sms dev sink: %w", err)
		}
	}
	if p.logger != nil {
		// Phone is masked and the body is omitted from logs; the full message is in sms_dev_sink.
		p.logger.Info(ctx, "local sms sent", "tenant_id", msg.TenantID, "outbox_id", msg.OutboxID,
			"to", MaskPhone(msg.To), "body_len", len(msg.Body))
	}
	return "local-" + uuid.New().String(), nil
}

// MaskPhone keeps only the last 4 characters for logs.
func MaskPhone(p string) string {
	if len(p) <= 4 {
		return "****"
	}
	return "****" + p[len(p)-4:]
}
