package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	apperrors "github.com/pareshvernekar/homecooked/internal/errors"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// Repository port for orders, lines, payments, and menu lookups.
type Repository interface {
	CreateOrder(ctx context.Context, o *models.CustomerOrder) error
	ListOrders(ctx context.Context) ([]*models.CustomerOrder, error)
	GetActiveOrder(ctx context.Context, id string) (*models.CustomerOrder, error)
	UpdateOrder(ctx context.Context, o *models.CustomerOrder, freeze []models.OrderLineFreeze) error

	InsertLine(ctx context.Context, item *models.OrderItem) error
	UpdateLine(ctx context.Context, item *models.OrderItem, replaceSelections bool) error
	DeactivateLine(ctx context.Context, orderID, itemID string) error
	GetActiveLine(ctx context.Context, orderID, itemID string) (*models.OrderItem, error)
	ListActiveLines(ctx context.Context, orderID string) ([]models.OrderItem, error)

	InsertPayment(ctx context.Context, p *models.OrderPayment) error
	ListPayments(ctx context.Context, orderID string) ([]models.OrderPayment, error)

	GetMenuItemOnMenu(ctx context.Context, menuID, itemID string) (*models.MenuItem, error)
	GetSizeOptionPrices(ctx context.Context, optionIDs []string) (map[string]float64, error)
}

// MenuReader loads the menu an order binds to.
type MenuReader interface {
	GetActiveMenu(ctx context.Context, tenantID, id string) (*models.Menu, error)
}

// Service implements order intake, lines, pricing, and payments.
// REQORDER001–REQORDER005, REQOLINE001–REQOLINE004, REQPAY001–REQPAY004
type Service struct {
	repo   Repository
	menus  MenuReader
	logger *logger.Logger
	now    func() time.Time
}

// NewService constructs an order service.
func NewService(repo Repository, menus MenuReader, l *logger.Logger) *Service {
	return &Service{repo: repo, menus: menus, logger: l, now: time.Now}
}

func (s *Service) nowMillis() int64 { return s.now().UTC().UnixMilli() }

// round2 rounds to cents; money columns are NUMERIC(10,2).
func round2(v float64) float64 { return math.Round(v*100) / 100 }

func validation(msg string, field string) error {
	return apperrors.CreateValidationError(msg, map[string]interface{}{"field": field})
}

// Create creates an order against an active published daily|catering menu.
// REQORDER001, REQORDER004, REQORDER005
func (s *Service) Create(ctx context.Context, tenantID string, req *models.OrderCreateRequest) (*models.CustomerOrder, error) {
	name := strings.TrimSpace(req.CustomerName)
	phone := strings.TrimSpace(req.CustomerPhone)
	if name == "" {
		return nil, validation("customer_name is required", "customer_name")
	}
	if phone == "" {
		return nil, validation("customer_phone is required", "customer_phone")
	}
	if req.ExpectedAt == nil {
		return nil, validation("expected_at is required", "expected_at")
	}
	if strings.TrimSpace(req.MenuID) == "" {
		return nil, validation("menu_id is required", "menu_id")
	}

	m, err := s.menus.GetActiveMenu(ctx, tenantID, req.MenuID)
	if err != nil {
		return nil, err
	}
	if m.Status != models.MenuStatusPublished {
		return nil, validation("orders require a published menu", "menu_id")
	}
	if m.MenuType != models.MenuTypeDaily && m.MenuType != models.MenuTypeCatering {
		return nil, validation("orders are only supported for daily or catering menus", "menu_id")
	}

	received := s.nowMillis()
	if req.ReceivedAt != nil {
		received = *req.ReceivedAt
	}
	o := &models.CustomerOrder{
		MenuID: m.ID, CustomerName: name, CustomerPhone: phone,
		ReceivedAt: received, ExpectedAt: *req.ExpectedAt,
		Status: models.OrderStatusReceived, CustomizationText: textPtr(req.CustomizationText),
	}
	if err := s.repo.CreateOrder(ctx, o); err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to create order", err, tenantID)
	}
	return o, nil
}

// List lists active orders for the tenant.
// REQORDER002
func (s *Service) List(ctx context.Context, tenantID string) ([]*models.CustomerOrder, error) {
	orders, err := s.repo.ListOrders(ctx)
	if err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to list orders", err, tenantID)
	}
	if orders == nil {
		orders = []*models.CustomerOrder{}
	}
	return orders, nil
}

// Get returns an order with lines, resolved prices, and payment aggregates.
// REQORDER002, REQOLINE003, REQPAY002
func (s *Service) Get(ctx context.Context, tenantID, id string) (*models.OrderDetail, error) {
	o, err := s.loadOrder(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	return s.buildDetail(ctx, tenantID, o)
}

// Update patches the order header; moving to PICKEDUP freezes the charged total.
// REQORDER003, REQORDER004, REQORDER005, REQPAY003S02
func (s *Service) Update(ctx context.Context, tenantID, id string, req *models.OrderUpdateRequest) (*models.OrderDetail, error) {
	o, err := s.loadOrder(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	var newStatus string
	if req.Status != nil {
		newStatus = strings.ToUpper(strings.TrimSpace(*req.Status))
		if !models.IsValidOrderStatus(newStatus) {
			return nil, validation("status must be RECEIVED, IN_PROGRESS, COMPLETE, or PICKEDUP", "status")
		}
	}

	if o.Status == models.OrderStatusPickedUp {
		// REQORDER003: after PICKEDUP only pickedup_at may be corrected.
		if req.TotalOverride.Set {
			return nil, validation("total_override cannot be changed after PICKEDUP", "total_override")
		}
		if req.CustomerName != nil || req.CustomerPhone != nil || req.ReceivedAt != nil ||
			req.ExpectedAt != nil || req.CustomizationText != nil ||
			(req.Status != nil && newStatus != models.OrderStatusPickedUp) {
			return nil, validation("order is PICKEDUP; only pickedup_at may be updated", "status")
		}
		if req.PickedupAt != nil {
			o.PickedupAt = req.PickedupAt
			if err := s.repo.UpdateOrder(ctx, o, nil); err != nil {
				return nil, s.mapOrderWriteErr(err, tenantID, id)
			}
		}
		return s.buildDetail(ctx, tenantID, o)
	}

	if req.CustomerName != nil {
		n := strings.TrimSpace(*req.CustomerName)
		if n == "" {
			return nil, validation("customer_name cannot be empty", "customer_name")
		}
		o.CustomerName = n
	}
	if req.CustomerPhone != nil {
		p := strings.TrimSpace(*req.CustomerPhone)
		if p == "" {
			return nil, validation("customer_phone cannot be empty", "customer_phone")
		}
		o.CustomerPhone = p
	}
	if req.ReceivedAt != nil {
		o.ReceivedAt = *req.ReceivedAt
	}
	if req.ExpectedAt != nil {
		o.ExpectedAt = *req.ExpectedAt
	}
	if req.PickedupAt != nil {
		o.PickedupAt = req.PickedupAt
	}
	if req.CustomizationText != nil {
		o.CustomizationText = textPtr(req.CustomizationText)
	}
	if req.TotalOverride.Set {
		if req.TotalOverride.Value == nil {
			o.TotalOverride = nil
		} else {
			if *req.TotalOverride.Value < 0 {
				return nil, validation("total_override must be >= 0", "total_override")
			}
			v := round2(*req.TotalOverride.Value)
			o.TotalOverride = &v
		}
	}

	var freeze []models.OrderLineFreeze
	if req.Status != nil {
		o.Status = newStatus
		if newStatus == models.OrderStatusPickedUp {
			// REQORDER004: freeze charged total and line amounts as resolved now.
			lines, prices, err := s.loadLinesAndPrices(ctx, tenantID, o.ID)
			if err != nil {
				return nil, err
			}
			_, charged := applyPricing(o, lines, prices)
			o.FrozenTotal = &charged
			for _, l := range lines {
				freeze = append(freeze, models.OrderLineFreeze{ItemID: l.ID, UnitPrice: l.UnitPrice, ExtendedAmount: l.ExtendedAmount})
			}
			if o.PickedupAt == nil {
				now := s.nowMillis()
				o.PickedupAt = &now
			}
		}
	}

	if err := s.repo.UpdateOrder(ctx, o, freeze); err != nil {
		return nil, s.mapOrderWriteErr(err, tenantID, id)
	}
	return s.buildDetail(ctx, tenantID, o)
}

// AddLine adds a line with per-component size selections.
// REQOLINE001, REQOLINE002
func (s *Service) AddLine(ctx context.Context, tenantID, orderID string, req *models.OrderLineCreateRequest) (*models.OrderItem, error) {
	o, err := s.loadOrder(ctx, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	if err := requireUnfulfilled(o); err != nil {
		return nil, err
	}
	if req.Quantity <= 0 {
		return nil, validation("quantity must be a positive integer", "quantity")
	}
	override, err := normalizeOverride(req.UnitPriceOverride, "unit_price_override")
	if err != nil {
		return nil, err
	}
	menuItem, err := s.loadMenuItem(ctx, tenantID, o.MenuID, req.MenuItemID)
	if err != nil {
		return nil, err
	}
	sels, err := resolveSelections(menuItem, nil, req.Selections)
	if err != nil {
		return nil, err
	}

	item := &models.OrderItem{
		OrderID: o.ID, MenuItemID: menuItem.ID, Quantity: req.Quantity,
		CustomizationText: textPtr(req.CustomizationText), UnitPriceOverride: override,
		Selections: sels,
	}
	if err := s.repo.InsertLine(ctx, item); err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to add order line", err, tenantID)
	}
	return s.resolveSingleLine(ctx, tenantID, o, item)
}

// UpdateLine updates selections, quantity, customization, and unit price override.
// REQOLINE002, REQOLINE004
func (s *Service) UpdateLine(ctx context.Context, tenantID, orderID, itemID string, req *models.OrderLineUpdateRequest) (*models.OrderItem, error) {
	o, err := s.loadOrder(ctx, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	if err := requireUnfulfilled(o); err != nil {
		return nil, err
	}
	item, err := s.repo.GetActiveLine(ctx, orderID, itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.CreateNotFoundError("order_item", itemID, tenantID)
		}
		return nil, apperrors.CreateDatabaseError("Failed to load order line", err, tenantID)
	}

	if req.Quantity != nil {
		if *req.Quantity <= 0 {
			return nil, validation("quantity must be a positive integer", "quantity")
		}
		item.Quantity = *req.Quantity
	}
	if req.CustomizationText != nil {
		item.CustomizationText = textPtr(req.CustomizationText)
	}
	if req.UnitPriceOverride.Set {
		ov, err := normalizeOverride(req.UnitPriceOverride.Value, "unit_price_override")
		if err != nil {
			return nil, err
		}
		item.UnitPriceOverride = ov
	}

	replace := len(req.Selections) > 0
	if replace {
		menuItem, err := s.loadMenuItem(ctx, tenantID, o.MenuID, item.MenuItemID)
		if err != nil {
			return nil, err
		}
		sels, err := resolveSelections(menuItem, item.Selections, req.Selections)
		if err != nil {
			return nil, err
		}
		item.Selections = sels
	}

	if err := s.repo.UpdateLine(ctx, item, replace); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.CreateNotFoundError("order_item", itemID, tenantID)
		}
		return nil, apperrors.CreateDatabaseError("Failed to update order line", err, tenantID)
	}
	return s.resolveSingleLine(ctx, tenantID, o, item)
}

// RemoveLine soft-deletes a line from an unfulfilled order.
// REQOLINE004
func (s *Service) RemoveLine(ctx context.Context, tenantID, orderID, itemID string) error {
	o, err := s.loadOrder(ctx, tenantID, orderID)
	if err != nil {
		return err
	}
	if err := requireUnfulfilled(o); err != nil {
		return err
	}
	if err := s.repo.DeactivateLine(ctx, orderID, itemID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.CreateNotFoundError("order_item", itemID, tenantID)
		}
		return apperrors.CreateDatabaseError("Failed to remove order line", err, tenantID)
	}
	return nil
}

// RecordPayment records a payment on an active order in any status.
// REQPAY001
func (s *Service) RecordPayment(ctx context.Context, tenantID, orderID string, req *models.PaymentCreateRequest) (*models.OrderPayment, error) {
	if _, err := s.loadOrder(ctx, tenantID, orderID); err != nil {
		return nil, err
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if !models.IsValidPaymentMode(mode) {
		return nil, validation("mode must be cash, credit, paypal, zelle, or venmo", "mode")
	}
	amount := round2(req.Amount)
	if amount <= 0 {
		return nil, validation("amount must be > 0", "amount")
	}
	p := &models.OrderPayment{OrderID: orderID, Mode: mode, Amount: amount, ReferenceText: textPtr(req.ReferenceText)}
	if err := s.repo.InsertPayment(ctx, p); err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to record payment", err, tenantID)
	}
	return p, nil
}

// ListPayments lists an order's payments in recorded order.
// REQPAY004
func (s *Service) ListPayments(ctx context.Context, tenantID, orderID string) ([]models.OrderPayment, error) {
	if _, err := s.loadOrder(ctx, tenantID, orderID); err != nil {
		return nil, err
	}
	payments, err := s.repo.ListPayments(ctx, orderID)
	if err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to list payments", err, tenantID)
	}
	return payments, nil
}

// --- helpers ---

func (s *Service) loadOrder(ctx context.Context, tenantID, id string) (*models.CustomerOrder, error) {
	o, err := s.repo.GetActiveOrder(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.CreateNotFoundError("order", id, tenantID)
		}
		return nil, apperrors.CreateDatabaseError("Failed to load order", err, tenantID)
	}
	return o, nil
}

func (s *Service) mapOrderWriteErr(err error, tenantID, id string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return apperrors.CreateNotFoundError("order", id, tenantID)
	}
	return apperrors.CreateDatabaseError("Failed to update order", err, tenantID)
}

func (s *Service) loadMenuItem(ctx context.Context, tenantID, menuID, menuItemID string) (*models.MenuItem, error) {
	if strings.TrimSpace(menuItemID) == "" {
		return nil, validation("menu_item_id is required", "menu_item_id")
	}
	mi, err := s.repo.GetMenuItemOnMenu(ctx, menuID, menuItemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.CreateValidationError("menu_item_id is not on the order's menu", map[string]interface{}{"menu_item_id": menuItemID})
		}
		return nil, apperrors.CreateDatabaseError("Failed to load menu item", err, tenantID)
	}
	return mi, nil
}

func (s *Service) loadLinesAndPrices(ctx context.Context, tenantID, orderID string) ([]models.OrderItem, map[string]float64, error) {
	lines, err := s.repo.ListActiveLines(ctx, orderID)
	if err != nil {
		return nil, nil, apperrors.CreateDatabaseError("Failed to load order lines", err, tenantID)
	}
	var ids []string
	for _, l := range lines {
		for _, sel := range l.Selections {
			ids = append(ids, sel.SizeOptionID)
		}
	}
	prices, err := s.repo.GetSizeOptionPrices(ctx, ids)
	if err != nil {
		return nil, nil, apperrors.CreateDatabaseError("Failed to load size option prices", err, tenantID)
	}
	return lines, prices, nil
}

// buildDetail resolves pricing and payment aggregates.
// REQORDER002, REQOLINE003, REQPAY002, REQPAY003
func (s *Service) buildDetail(ctx context.Context, tenantID string, o *models.CustomerOrder) (*models.OrderDetail, error) {
	lines, prices, err := s.loadLinesAndPrices(ctx, tenantID, o.ID)
	if err != nil {
		return nil, err
	}
	subtotal, charged := applyPricing(o, lines, prices)

	payments, err := s.repo.ListPayments(ctx, o.ID)
	if err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to load payments", err, tenantID)
	}
	var paid float64
	for _, p := range payments {
		paid += p.Amount
	}
	paid = round2(paid)
	balance := round2(charged - paid)

	return &models.OrderDetail{
		CustomerOrder:   *o,
		Lines:           lines,
		Subtotal:        subtotal,
		ChargedTotal:    charged,
		PaidAmount:      paid,
		Balance:         balance,
		PaymentReceived: balance <= 0,
		OverpaidAmount:  math.Max(0, -balance),
	}, nil
}

// resolveSingleLine reloads a just-written line and applies pricing to it.
func (s *Service) resolveSingleLine(ctx context.Context, tenantID string, o *models.CustomerOrder, item *models.OrderItem) (*models.OrderItem, error) {
	line, err := s.repo.GetActiveLine(ctx, o.ID, item.ID)
	if err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to load order line", err, tenantID)
	}
	var ids []string
	for _, sel := range line.Selections {
		ids = append(ids, sel.SizeOptionID)
	}
	prices, err := s.repo.GetSizeOptionPrices(ctx, ids)
	if err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to load size option prices", err, tenantID)
	}
	lines := []models.OrderItem{*line}
	applyPricing(o, lines, prices)
	return &lines[0], nil
}

func requireUnfulfilled(o *models.CustomerOrder) error {
	if o.Status == models.OrderStatusPickedUp {
		return apperrors.CreateValidationError("order is PICKEDUP; lines can no longer be changed", map[string]interface{}{"order_id": o.ID})
	}
	return nil
}

func normalizeOverride(v *float64, field string) (*float64, error) {
	if v == nil {
		return nil, nil
	}
	if *v < 0 {
		return nil, validation(field+" must be >= 0", field)
	}
	r := round2(*v)
	return &r, nil
}

// textPtr trims free text; empty becomes nil (cleared).
func textPtr(p *string) *string {
	if p == nil {
		return nil
	}
	t := strings.TrimSpace(*p)
	if t == "" {
		return nil
	}
	return &t
}

// resolveSelections builds one selection per active component: supplied > existing > component default.
// Every supplied selection must reference a component of the item and one of its size options.
// REQOLINE001, REQOLINE004S01
func resolveSelections(item *models.MenuItem, existing []models.OrderItemComponentSelection, supplied []models.OrderSelectionInput) ([]models.OrderItemComponentSelection, error) {
	byComp := make(map[string]*models.MenuItemComponent, len(item.Components))
	for i := range item.Components {
		byComp[item.Components[i].ID] = &item.Components[i]
	}

	chosen := make(map[string]string, len(item.Components))
	for _, e := range existing {
		chosen[e.MenuItemComponentID] = e.SizeOptionID
	}
	seen := make(map[string]bool, len(supplied))
	for _, in := range supplied {
		comp, ok := byComp[in.MenuItemComponentID]
		if !ok {
			return nil, apperrors.CreateValidationError("menu_item_component_id is not a component of the menu-item",
				map[string]interface{}{"menu_item_component_id": in.MenuItemComponentID})
		}
		if seen[comp.ID] {
			return nil, apperrors.CreateValidationError("duplicate selection for component",
				map[string]interface{}{"menu_item_component_id": comp.ID})
		}
		seen[comp.ID] = true
		if !hasOption(comp, in.SizeOptionID) {
			return nil, apperrors.CreateValidationError("size_option_id is not an option of the component on the order's menu",
				map[string]interface{}{"menu_item_component_id": comp.ID, "size_option_id": in.SizeOptionID})
		}
		chosen[comp.ID] = in.SizeOptionID
	}

	out := make([]models.OrderItemComponentSelection, 0, len(item.Components))
	for _, comp := range item.Components {
		optID := chosen[comp.ID]
		if optID == "" || !hasOption(&comp, optID) {
			if comp.DefaultSizeOptionID == nil || !hasOption(&comp, *comp.DefaultSizeOptionID) {
				return nil, apperrors.CreateValidationError(fmt.Sprintf("component %s has no usable default size option", comp.ID),
					map[string]interface{}{"menu_item_component_id": comp.ID})
			}
			optID = *comp.DefaultSizeOptionID
		}
		out = append(out, models.OrderItemComponentSelection{MenuItemComponentID: comp.ID, SizeOptionID: optID})
	}
	return out, nil
}

func hasOption(c *models.MenuItemComponent, optionID string) bool {
	for _, o := range c.SizeOptions {
		if o.ID == optionID {
			return true
		}
	}
	return false
}

// applyPricing fills each line's resolved unit/extended amounts and selection prices, and returns
// (subtotal, charged_total). Unfulfilled: live option prices unless line override; order override wins.
// PICKEDUP: frozen values. REQOLINE003, REQORDER004
func applyPricing(o *models.CustomerOrder, lines []models.OrderItem, prices map[string]float64) (subtotal, charged float64) {
	frozen := o.Status == models.OrderStatusPickedUp && o.FrozenTotal != nil
	for i := range lines {
		l := &lines[i]
		for j := range l.Selections {
			l.Selections[j].Price = prices[l.Selections[j].SizeOptionID]
		}
		if frozen && l.FrozenUnitPrice != nil && l.FrozenExtendedAmount != nil {
			l.UnitPrice = *l.FrozenUnitPrice
			l.ExtendedAmount = *l.FrozenExtendedAmount
		} else {
			if l.UnitPriceOverride != nil {
				l.UnitPrice = *l.UnitPriceOverride
			} else {
				var sum float64
				for _, sel := range l.Selections {
					sum += sel.Price
				}
				l.UnitPrice = round2(sum)
			}
			l.ExtendedAmount = round2(l.UnitPrice * float64(l.Quantity))
		}
		subtotal += l.ExtendedAmount
	}
	subtotal = round2(subtotal)
	switch {
	case frozen:
		charged = *o.FrozenTotal
	case o.TotalOverride != nil:
		charged = *o.TotalOverride
	default:
		charged = subtotal
	}
	return subtotal, charged
}
