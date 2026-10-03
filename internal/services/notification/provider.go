// Package notification implements the order-notification outbox: message building,
// SMS provider abstraction, async delivery worker, and tenant notification settings.
package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// SmsMessage is one outbound SMS handed to a provider.
// REQNOTIF004
type SmsMessage struct {
	TenantID string
	OutboxID string
	To       string
	Body     string
}

// SmsProvider sends an SMS and returns a provider-side reference.
// REQNOTIF004
type SmsProvider interface {
	Send(ctx context.Context, msg SmsMessage) (providerRef string, err error)
}

// Supported SMS_PROVIDER values.
const ProviderLocal = "local"

// ErrUnknownProvider is returned by NewProvider for an unsupported provider name.
var ErrUnknownProvider = errors.New("unknown SMS provider")

// NewProvider selects the SmsProvider by configuration name ("" defaults to local).
// REQNOTIF004
func NewProvider(name string, sink SinkRecorder, l Logger) (SmsProvider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", ProviderLocal:
		return NewLocalSmsProvider(sink, l), nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownProvider, name)
	}
}

// FailSmsProvider always fails; inject it to exercise retry and dead-letter paths.
// REQNOTIF003
type FailSmsProvider struct {
	Err   error
	Calls int
}

// Send implements SmsProvider.
func (f *FailSmsProvider) Send(context.Context, SmsMessage) (string, error) {
	f.Calls++
	if f.Err != nil {
		return "", f.Err
	}
	return "", errors.New("sms provider unavailable")
}
