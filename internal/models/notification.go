package models

import (
	"bytes"
	"encoding/json"
)

// Outbox event types (REQNOTIF002). Keys are extensible for later channels/templates.
const (
	EventOrderCreated  = "order.created"
	EventOrderAccepted = "order.accepted"
	EventOrderDeclined = "order.declined"
	EventOrderReady    = "order.ready"
	EventOrderPickedUp = "order.picked_up"
)

// Outbox channels.
const ChannelSMS = "sms"

// Outbox statuses (REQNOTIF003).
const (
	OutboxStatusPending    = "pending"
	OutboxStatusProcessing = "processing"
	OutboxStatusDelivered  = "delivered"
	OutboxStatusDead       = "dead"
)

// NotificationOutbox is one transactional-outbox row for an outbound notification.
// REQNOTIF002, REQNOTIF003, REQNOTIF005
type NotificationOutbox struct {
	ID             string  `db:"id" json:"id"`
	TenantID       string  `db:"tenant_id" json:"tenant_id"`
	OrderID        string  `db:"order_id" json:"order_id"`
	EventType      string  `db:"event_type" json:"event_type"`
	Channel        string  `db:"channel" json:"channel"`
	RecipientPhone string  `db:"recipient_phone" json:"recipient_phone"`
	Body           string  `db:"body" json:"body"`
	Status         string  `db:"status" json:"status"`
	Attempts       int     `db:"attempts" json:"attempts"`
	NextAttemptAt  int64   `db:"next_attempt_at" json:"next_attempt_at"`
	LastError      *string `db:"last_error" json:"last_error,omitempty"`
	ProviderRef    *string `db:"provider_ref" json:"provider_ref,omitempty"`
	CreatedAt      int64   `db:"created_at" json:"created_at"`
	UpdatedAt      int64   `db:"updated_at" json:"updated_at"`
}

// SmsDevSinkEntry is a message recorded by the local SMS provider (dev/tests only).
// REQNOTIF004
type SmsDevSinkEntry struct {
	ID        string `db:"id" json:"id"`
	TenantID  string `db:"tenant_id" json:"tenant_id"`
	OutboxID  string `db:"outbox_id" json:"outbox_id"`
	Phone     string `db:"phone" json:"phone"`
	Body      string `db:"body" json:"body"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

// TenantSettings is the tenant-level configuration exposed over the API.
// REQNOTIF001
type TenantSettings struct {
	CookAdminPhone *string  `json:"cook_admin_phone"`
	Warnings       []string `json:"warnings,omitempty"`
}

// NullableString distinguishes an absent JSON field from an explicit null/empty string.
// REQNOTIF001
type NullableString struct {
	Set   bool
	Value *string
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullableString) UnmarshalJSON(b []byte) error {
	n.Set = true
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		n.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	n.Value = &s
	return nil
}

// MarshalJSON implements json.Marshaler.
func (n NullableString) MarshalJSON() ([]byte, error) {
	if n.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*n.Value)
}

// TenantSettingsUpdateRequest sets or clears the tenant cook admin phone.
// null or an empty string clears it. REQNOTIF001
type TenantSettingsUpdateRequest struct {
	CookAdminPhone NullableString `json:"cook_admin_phone"`
}
