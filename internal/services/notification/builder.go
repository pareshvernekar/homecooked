package notification

import (
	"context"
	"fmt"
	"strings"

	"github.com/pareshvernekar/homecooked/internal/models"
)

// CookPhoneReader reads the tenant's cook admin phone ("" when unset).
type CookPhoneReader interface {
	GetCookAdminPhone(ctx context.Context, tenantID string) (string, error)
}

// Builder turns an order lifecycle event into an outbox row (not yet persisted).
// The order repository inserts the row in the same transaction as the order write.
// REQNOTIF002
type Builder struct {
	cook   CookPhoneReader
	logger Logger
}

// NewBuilder constructs an outbox builder.
func NewBuilder(cook CookPhoneReader, l Logger) *Builder {
	return &Builder{cook: cook, logger: l}
}

// Build returns the outbox row for event on order o, or nil when nothing should be enqueued
// (order.created with no cook_admin_phone, or an event with no notification such as preparing).
// o.ID must be set; for order.declined o.RefuseReason is included in the body.
// REQNOTIF002, REQNOTIF002S01, REQNOTIF002S02, REQNOTIF002S04
func (b *Builder) Build(ctx context.Context, tenantID, eventType string, o *models.CustomerOrder) (*models.NotificationOutbox, error) {
	var phone, body string
	switch eventType {
	case models.EventOrderCreated:
		cook, err := b.cook.GetCookAdminPhone(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("load cook_admin_phone: %w", err)
		}
		if cook == "" {
			// REQNOTIF002S02: skip, order create still succeeds.
			if b.logger != nil {
				b.logger.Warn(ctx, "cook_admin_phone not set; skipping order.created notification",
					"tenant_id", tenantID, "order_id", o.ID)
			}
			return nil, nil
		}
		phone = cook
		body = fmt.Sprintf("New order %s from %s", o.ID, o.CustomerName)
	case models.EventOrderAccepted:
		phone, body = o.CustomerPhone, fmt.Sprintf("Order %s accepted", o.ID)
	case models.EventOrderDeclined:
		reason := models.DefaultRefuseReason
		if o.RefuseReason != nil && strings.TrimSpace(*o.RefuseReason) != "" {
			reason = *o.RefuseReason
		}
		phone, body = o.CustomerPhone, fmt.Sprintf("Order %s declined: %s", o.ID, reason)
	case models.EventOrderReady:
		phone, body = o.CustomerPhone, fmt.Sprintf("Order %s is ready for pickup", o.ID)
	case models.EventOrderPickedUp:
		phone, body = o.CustomerPhone, fmt.Sprintf("Order %s picked up", o.ID)
	default:
		return nil, fmt.Errorf("unsupported notification event %q", eventType)
	}
	return &models.NotificationOutbox{
		TenantID: tenantID, OrderID: o.ID, EventType: eventType, Channel: models.ChannelSMS,
		RecipientPhone: strings.TrimSpace(phone), Body: body,
	}, nil
}
