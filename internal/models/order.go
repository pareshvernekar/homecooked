package models

import (
	"bytes"
	"encoding/json"
)

// Order status constants.
// REQORDER004, REQLIFE001
const (
	OrderStatusReceived   = "RECEIVED"
	OrderStatusAccepted   = "ACCEPTED"
	OrderStatusDeclined   = "DECLINED"
	OrderStatusInProgress = "IN_PROGRESS"
	OrderStatusReady      = "READY"
	OrderStatusPickedUp   = "PICKEDUP"

	DefaultRefuseReason = "No available slots"
)

// Payment mode constants.
// REQPAY001
const (
	PaymentModeCash   = "cash"
	PaymentModeCredit = "credit"
	PaymentModePayPal = "paypal"
	PaymentModeZelle  = "zelle"
	PaymentModeVenmo  = "venmo"
)

// IsValidOrderStatus reports whether s is a supported order status.
func IsValidOrderStatus(s string) bool {
	switch s {
	case OrderStatusReceived, OrderStatusAccepted, OrderStatusDeclined,
		OrderStatusInProgress, OrderStatusReady, OrderStatusPickedUp:
		return true
	}
	return false
}

// IsUnfulfilledOrderStatus is true when live pricing and line edits apply.
// REQORDER004, REQOLINE001
func IsUnfulfilledOrderStatus(s string) bool {
	switch s {
	case OrderStatusReceived, OrderStatusAccepted, OrderStatusInProgress, OrderStatusReady:
		return true
	}
	return false
}

// IsValidPaymentMode reports whether m is a supported payment mode.
func IsValidPaymentMode(m string) bool {
	switch m {
	case PaymentModeCash, PaymentModeCredit, PaymentModePayPal, PaymentModeZelle, PaymentModeVenmo:
		return true
	}
	return false
}

// CustomerOrder is the tenant-scoped order header (table customer_order).
// Timestamps are Unix epoch milliseconds.
// REQORDER001–REQORDER005
type CustomerOrder struct {
	ID                string   `db:"id" json:"id"`
	TenantID          string   `db:"tenant_id" json:"tenant_id"`
	MenuID            string   `db:"menu_id" json:"menu_id"`
	CustomerName      string   `db:"customer_name" json:"customer_name"`
	CustomerPhone     string   `db:"customer_phone" json:"customer_phone"`
	ReceivedAt        int64    `db:"received_at" json:"received_at"`
	ExpectedAt        int64    `db:"expected_at" json:"expected_at"`
	PickedupAt        *int64   `db:"pickedup_at" json:"pickedup_at,omitempty"`
	Status            string   `db:"status" json:"status"`
	CustomizationText *string  `db:"customization_text" json:"customization_text,omitempty"`
	RefuseReason      *string  `db:"refuse_reason" json:"refuse_reason,omitempty"`
	TotalOverride     *float64 `db:"total_override" json:"total_override,omitempty"`
	FrozenTotal       *float64 `db:"frozen_total" json:"frozen_total,omitempty"`
	IsActive          bool     `db:"is_active" json:"is_active"`
	CreatedAt         int64    `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt         int64    `db:"updated_at" json:"updated_at,omitempty"`
}

// OrderItem is an order line (table order_item).
// REQOLINE001–REQOLINE004
type OrderItem struct {
	ID                   string   `db:"id" json:"id"`
	TenantID             string   `db:"tenant_id" json:"tenant_id"`
	OrderID              string   `db:"order_id" json:"order_id"`
	MenuItemID           string   `db:"menu_item_id" json:"menu_item_id"`
	Quantity             int      `db:"quantity" json:"quantity"`
	CustomizationText    *string  `db:"customization_text" json:"customization_text,omitempty"`
	UnitPriceOverride    *float64 `db:"unit_price_override" json:"unit_price_override,omitempty"`
	FrozenUnitPrice      *float64 `db:"frozen_unit_price" json:"frozen_unit_price,omitempty"`
	FrozenExtendedAmount *float64 `db:"frozen_extended_amount" json:"frozen_extended_amount,omitempty"`
	IsActive             bool     `db:"is_active" json:"is_active"`
	CreatedAt            int64    `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt            int64    `db:"updated_at" json:"updated_at,omitempty"`

	// Resolved at read time (REQOLINE003): live or frozen.
	UnitPrice      float64                       `db:"-" json:"unit_price"`
	ExtendedAmount float64                       `db:"-" json:"extended_amount"`
	Selections     []OrderItemComponentSelection `db:"-" json:"selections"`
}

// OrderItemComponentSelection is the chosen size option for one component of a line.
// REQOLINE001
type OrderItemComponentSelection struct {
	ID                  string `db:"id" json:"id"`
	TenantID            string `db:"tenant_id" json:"tenant_id"`
	OrderItemID         string `db:"order_item_id" json:"order_item_id"`
	MenuItemComponentID string `db:"menu_item_component_id" json:"menu_item_component_id"`
	SizeOptionID        string `db:"size_option_id" json:"size_option_id"`
	CreatedAt           int64  `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt           int64  `db:"updated_at" json:"updated_at,omitempty"`

	// Price is the resolved option price at read time (live unless the order is frozen).
	Price float64 `db:"-" json:"price"`
}

// OrderPayment is one entry in the payment ledger (table order_payment).
// REQPAY001, REQPAY004
type OrderPayment struct {
	ID            string  `db:"id" json:"id"`
	TenantID      string  `db:"tenant_id" json:"tenant_id"`
	OrderID       string  `db:"order_id" json:"order_id"`
	Mode          string  `db:"mode" json:"mode"`
	Amount        float64 `db:"amount" json:"amount"`
	ReferenceText *string `db:"reference_text" json:"reference,omitempty"`
	CreatedAt     int64   `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt     int64   `db:"updated_at" json:"updated_at,omitempty"`
}

// OrderDetail is the get-order payload: header, lines, charged total and derived payment aggregates.
// REQORDER002, REQOLINE003, REQPAY002
type OrderDetail struct {
	CustomerOrder
	Lines           []OrderItem `json:"lines"`
	Subtotal        float64     `json:"subtotal"`
	ChargedTotal    float64     `json:"charged_total"`
	PaidAmount      float64     `json:"paid_amount"`
	Balance         float64     `json:"balance"`
	PaymentReceived bool        `json:"payment_received"`
	OverpaidAmount  float64     `json:"overpaid_amount"`
}

// OrderLineFreeze carries the per-line amounts persisted at PICKEDUP.
// REQORDER004
type OrderLineFreeze struct {
	ItemID         string
	UnitPrice      float64
	ExtendedAmount float64
}

// NullableFloat64 distinguishes an absent JSON field from an explicit null,
// so clients can clear an override by sending null.
// REQORDER003, REQOLINE002
type NullableFloat64 struct {
	Set   bool
	Value *float64
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullableFloat64) UnmarshalJSON(b []byte) error {
	n.Set = true
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		n.Value = nil
		return nil
	}
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	n.Value = &f
	return nil
}

// MarshalJSON implements json.Marshaler.
func (n NullableFloat64) MarshalJSON() ([]byte, error) {
	if n.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*n.Value)
}

// --- Request DTOs ---

// OrderCreateRequest creates an order against a published daily|catering menu.
// REQORDER001
type OrderCreateRequest struct {
	MenuID            string  `json:"menu_id"`
	CustomerName      string  `json:"customer_name"`
	CustomerPhone     string  `json:"customer_phone"`
	ReceivedAt        *int64  `json:"received_at,omitempty"`
	ExpectedAt        *int64  `json:"expected_at"`
	CustomizationText *string `json:"customization_text,omitempty"`
}

// OrderUpdateRequest patches the order header.
// REQORDER003
type OrderUpdateRequest struct {
	CustomerName      *string         `json:"customer_name,omitempty"`
	CustomerPhone     *string         `json:"customer_phone,omitempty"`
	ReceivedAt        *int64          `json:"received_at,omitempty"`
	ExpectedAt        *int64          `json:"expected_at,omitempty"`
	PickedupAt        *int64          `json:"pickedup_at,omitempty"`
	CustomizationText *string         `json:"customization_text,omitempty"`
	TotalOverride     NullableFloat64 `json:"total_override"`
	Status            *string         `json:"status,omitempty"`
}

// OrderSelectionInput selects a size option for one menu-item component.
// REQOLINE001
type OrderSelectionInput struct {
	MenuItemComponentID string `json:"menu_item_component_id"`
	SizeOptionID        string `json:"size_option_id"`
}

// OrderLineCreateRequest adds a line to an order.
// REQOLINE001, REQOLINE002
type OrderLineCreateRequest struct {
	MenuItemID        string                `json:"menu_item_id"`
	Quantity          int                   `json:"quantity"`
	CustomizationText *string               `json:"customization_text,omitempty"`
	UnitPriceOverride *float64              `json:"unit_price_override,omitempty"`
	Selections        []OrderSelectionInput `json:"selections,omitempty"`
}

// OrderLineUpdateRequest patches a line.
// REQOLINE002, REQOLINE004
type OrderLineUpdateRequest struct {
	Quantity          *int                  `json:"quantity,omitempty"`
	CustomizationText *string               `json:"customization_text,omitempty"`
	UnitPriceOverride NullableFloat64       `json:"unit_price_override"`
	Selections        []OrderSelectionInput `json:"selections,omitempty"`
}

// PaymentCreateRequest records a payment.
// REQPAY001
type PaymentCreateRequest struct {
	Mode          string  `json:"mode"`
	Amount        float64 `json:"amount"`
	ReferenceText *string `json:"reference,omitempty"`
}
