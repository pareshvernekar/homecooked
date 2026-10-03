package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// PostgreSQLOrderRepository implements order, line, and payment persistence.
// REQORDER001–REQORDER005, REQOLINE001–REQOLINE004, REQPAY001–REQPAY004
type PostgreSQLOrderRepository struct {
	DB       *sqlx.DB
	TenantID string
	Logger   *logger.Logger
}

// NewOrderRepository constructs a tenant-scoped order repository.
func NewOrderRepository(db *sqlx.DB, l *logger.Logger, tenantID string) *PostgreSQLOrderRepository {
	return &PostgreSQLOrderRepository{DB: db, TenantID: tenantID, Logger: l}
}

const orderSelectCols = `
	id, tenant_id, menu_id, customer_name, customer_phone, received_at, expected_at, pickedup_at,
	status, customization_text, refuse_reason, total_override::float8 AS total_override, frozen_total::float8 AS frozen_total,
	is_active, created_at, updated_at`

const orderItemSelectCols = `
	id, tenant_id, order_id, menu_item_id, quantity, customization_text,
	unit_price_override::float8 AS unit_price_override,
	frozen_unit_price::float8 AS frozen_unit_price,
	frozen_extended_amount::float8 AS frozen_extended_amount,
	is_active, created_at, updated_at`

const orderSelectionSelectCols = `
	id, tenant_id, order_item_id, menu_item_component_id, size_option_id, created_at, updated_at`

const orderPaymentSelectCols = `
	id, tenant_id, order_id, mode, amount::float8 AS amount, reference_text, created_at, updated_at`

// CreateOrder inserts a new order header.
// REQORDER001
func (r *PostgreSQLOrderRepository) CreateOrder(ctx context.Context, o *models.CustomerOrder) error {
	if err := requireTenantExists(ctx, r.DB, r.TenantID); err != nil {
		return err
	}
	now := time.Now().UTC().UnixMilli()
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	o.TenantID = r.TenantID
	o.IsActive = true
	o.CreatedAt = now
	o.UpdatedAt = now

	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO customer_order (
			id, tenant_id, menu_id, customer_name, customer_phone, received_at, expected_at, pickedup_at,
			status, customization_text, refuse_reason, total_override, frozen_total, is_active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$15)`,
		o.ID, o.TenantID, o.MenuID, o.CustomerName, o.CustomerPhone, o.ReceivedAt, o.ExpectedAt, o.PickedupAt,
		o.Status, o.CustomizationText, o.RefuseReason, o.TotalOverride, o.FrozenTotal, o.IsActive, now,
	)
	return err
}

// ListOrders returns active orders for the tenant, newest first.
// REQORDER002
func (r *PostgreSQLOrderRepository) ListOrders(ctx context.Context) ([]*models.CustomerOrder, error) {
	var orders []*models.CustomerOrder
	err := r.DB.SelectContext(ctx, &orders, `
		SELECT `+orderSelectCols+` FROM customer_order
		WHERE tenant_id = $1 AND is_active = TRUE
		ORDER BY created_at DESC, id ASC`, r.TenantID)
	return orders, err
}

// GetActiveOrder returns an active order for the tenant or sql.ErrNoRows.
// REQORDER002
func (r *PostgreSQLOrderRepository) GetActiveOrder(ctx context.Context, id string) (*models.CustomerOrder, error) {
	var o models.CustomerOrder
	err := r.DB.GetContext(ctx, &o, `
		SELECT `+orderSelectCols+` FROM customer_order
		WHERE tenant_id = $1 AND id = $2 AND is_active = TRUE`, r.TenantID, id)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// UpdateOrder writes the mutable header columns. Any per-line frozen amounts in freeze
// are written in the same transaction (freeze at PICKEDUP).
// REQORDER003, REQORDER004
func (r *PostgreSQLOrderRepository) UpdateOrder(ctx context.Context, o *models.CustomerOrder, freeze []models.OrderLineFreeze) error {
	now := time.Now().UTC().UnixMilli()
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE customer_order SET
			customer_name = $1, customer_phone = $2, received_at = $3, expected_at = $4, pickedup_at = $5,
			status = $6, customization_text = $7, refuse_reason = $8, total_override = $9, frozen_total = $10, updated_at = $11
		WHERE tenant_id = $12 AND id = $13 AND is_active = TRUE`,
		o.CustomerName, o.CustomerPhone, o.ReceivedAt, o.ExpectedAt, o.PickedupAt,
		o.Status, o.CustomizationText, o.RefuseReason, o.TotalOverride, o.FrozenTotal, now,
		r.TenantID, o.ID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}

	for _, f := range freeze {
		if _, err := tx.ExecContext(ctx, `
			UPDATE order_item SET frozen_unit_price = $1, frozen_extended_amount = $2, updated_at = $3
			WHERE tenant_id = $4 AND order_id = $5 AND id = $6 AND is_active = TRUE`,
			f.UnitPrice, f.ExtendedAmount, now, r.TenantID, o.ID, f.ItemID,
		); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	o.UpdatedAt = now
	return nil
}

// DeactivateOrder soft-deletes an order.
// REQORDER002
func (r *PostgreSQLOrderRepository) DeactivateOrder(ctx context.Context, id string) error {
	now := time.Now().UTC().UnixMilli()
	res, err := r.DB.ExecContext(ctx, `
		UPDATE customer_order SET is_active = FALSE, updated_at = $1
		WHERE tenant_id = $2 AND id = $3 AND is_active = TRUE`, now, r.TenantID, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// InsertLine inserts a line and its selections in one transaction.
// REQOLINE001
func (r *PostgreSQLOrderRepository) InsertLine(ctx context.Context, item *models.OrderItem) error {
	now := time.Now().UTC().UnixMilli()
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	item.TenantID = r.TenantID
	item.IsActive = true
	item.CreatedAt = now
	item.UpdatedAt = now

	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO order_item (
			id, tenant_id, order_id, menu_item_id, quantity, customization_text, unit_price_override,
			is_active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,TRUE,$8,$8)`,
		item.ID, r.TenantID, item.OrderID, item.MenuItemID, item.Quantity,
		item.CustomizationText, item.UnitPriceOverride, now,
	); err != nil {
		return err
	}
	if err := insertSelections(ctx, tx, r.TenantID, item, now); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateLine writes mutable line columns and replaces selections with item.Selections
// when replaceSelections is true.
// REQOLINE004
func (r *PostgreSQLOrderRepository) UpdateLine(ctx context.Context, item *models.OrderItem, replaceSelections bool) error {
	now := time.Now().UTC().UnixMilli()
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE order_item SET quantity = $1, customization_text = $2, unit_price_override = $3, updated_at = $4
		WHERE tenant_id = $5 AND order_id = $6 AND id = $7 AND is_active = TRUE`,
		item.Quantity, item.CustomizationText, item.UnitPriceOverride, now,
		r.TenantID, item.OrderID, item.ID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	if replaceSelections {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM order_item_component_selection WHERE tenant_id = $1 AND order_item_id = $2`,
			r.TenantID, item.ID); err != nil {
			return err
		}
		if err := insertSelections(ctx, tx, r.TenantID, item, now); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	item.UpdatedAt = now
	return nil
}

func insertSelections(ctx context.Context, tx *sqlx.Tx, tenantID string, item *models.OrderItem, now int64) error {
	for i := range item.Selections {
		sel := &item.Selections[i]
		if sel.ID == "" {
			sel.ID = uuid.New().String()
		}
		sel.TenantID = tenantID
		sel.OrderItemID = item.ID
		sel.CreatedAt = now
		sel.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO order_item_component_selection (
				id, tenant_id, order_item_id, menu_item_component_id, size_option_id, created_at, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$6)`,
			sel.ID, tenantID, item.ID, sel.MenuItemComponentID, sel.SizeOptionID, now,
		); err != nil {
			return err
		}
	}
	return nil
}

// DeactivateLine soft-deletes a line.
// REQOLINE004
func (r *PostgreSQLOrderRepository) DeactivateLine(ctx context.Context, orderID, itemID string) error {
	now := time.Now().UTC().UnixMilli()
	res, err := r.DB.ExecContext(ctx, `
		UPDATE order_item SET is_active = FALSE, updated_at = $1
		WHERE tenant_id = $2 AND order_id = $3 AND id = $4 AND is_active = TRUE`,
		now, r.TenantID, orderID, itemID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetActiveLine returns an active line with its selections or sql.ErrNoRows.
func (r *PostgreSQLOrderRepository) GetActiveLine(ctx context.Context, orderID, itemID string) (*models.OrderItem, error) {
	var item models.OrderItem
	if err := r.DB.GetContext(ctx, &item, `
		SELECT `+orderItemSelectCols+` FROM order_item
		WHERE tenant_id = $1 AND order_id = $2 AND id = $3 AND is_active = TRUE`,
		r.TenantID, orderID, itemID); err != nil {
		return nil, err
	}
	sels, err := r.listSelections(ctx, orderID)
	if err != nil {
		return nil, err
	}
	item.Selections = sels[item.ID]
	if item.Selections == nil {
		item.Selections = []models.OrderItemComponentSelection{}
	}
	return &item, nil
}

// ListActiveLines returns active lines (oldest first) with selections.
// REQORDER002
func (r *PostgreSQLOrderRepository) ListActiveLines(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	items := []models.OrderItem{}
	if err := r.DB.SelectContext(ctx, &items, `
		SELECT `+orderItemSelectCols+` FROM order_item
		WHERE tenant_id = $1 AND order_id = $2 AND is_active = TRUE
		ORDER BY created_at ASC, id ASC`, r.TenantID, orderID); err != nil {
		return nil, err
	}
	sels, err := r.listSelections(ctx, orderID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Selections = sels[items[i].ID]
		if items[i].Selections == nil {
			items[i].Selections = []models.OrderItemComponentSelection{}
		}
	}
	return items, nil
}

// listSelections loads selections for all active lines of an order, keyed by order_item_id.
func (r *PostgreSQLOrderRepository) listSelections(ctx context.Context, orderID string) (map[string][]models.OrderItemComponentSelection, error) {
	var rows []models.OrderItemComponentSelection
	if err := r.DB.SelectContext(ctx, &rows, `
		SELECT `+orderSelectionSelectCols+` FROM order_item_component_selection
		WHERE tenant_id = $1 AND order_item_id IN (
			SELECT id FROM order_item WHERE tenant_id = $1 AND order_id = $2 AND is_active = TRUE
		)
		ORDER BY created_at ASC, id ASC`, r.TenantID, orderID); err != nil {
		return nil, err
	}
	out := make(map[string][]models.OrderItemComponentSelection)
	for _, s := range rows {
		out[s.OrderItemID] = append(out[s.OrderItemID], s)
	}
	return out, nil
}

// InsertPayment records a payment.
// REQPAY001
func (r *PostgreSQLOrderRepository) InsertPayment(ctx context.Context, p *models.OrderPayment) error {
	now := time.Now().UTC().UnixMilli()
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	p.TenantID = r.TenantID
	p.CreatedAt = now
	p.UpdatedAt = now
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO order_payment (id, tenant_id, order_id, mode, amount, reference_text, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$7)`,
		p.ID, r.TenantID, p.OrderID, p.Mode, p.Amount, p.ReferenceText, now)
	return err
}

// ListPayments returns an order's payments in recorded order.
// REQPAY004
func (r *PostgreSQLOrderRepository) ListPayments(ctx context.Context, orderID string) ([]models.OrderPayment, error) {
	payments := []models.OrderPayment{}
	err := r.DB.SelectContext(ctx, &payments, `
		SELECT `+orderPaymentSelectCols+` FROM order_payment
		WHERE tenant_id = $1 AND order_id = $2
		ORDER BY created_at ASC, id ASC`, r.TenantID, orderID)
	return payments, err
}

// GetMenuItemOnMenu loads an active menu-item on the menu with its active components and
// active size options (IsDefault set from the component default). Returns sql.ErrNoRows if absent.
// REQOLINE001
func (r *PostgreSQLOrderRepository) GetMenuItemOnMenu(ctx context.Context, menuID, itemID string) (*models.MenuItem, error) {
	var item models.MenuItem
	if err := r.DB.GetContext(ctx, &item, `
		SELECT id, tenant_id, menu_id, category_id, kind, name, COALESCE(description,'') AS description,
		       sequence, is_active, created_at, updated_at
		FROM menu_item
		WHERE tenant_id = $1 AND menu_id = $2 AND id = $3 AND is_active = TRUE`,
		r.TenantID, menuID, itemID); err != nil {
		return nil, err
	}

	var comps []models.MenuItemComponent
	if err := r.DB.SelectContext(ctx, &comps, `
		SELECT id, tenant_id, menu_item_id, food_item_id, default_size_option_id, is_active, created_at, updated_at
		FROM menu_item_component
		WHERE tenant_id = $1 AND menu_item_id = $2 AND is_active = TRUE
		ORDER BY created_at ASC, id ASC`, r.TenantID, itemID); err != nil {
		return nil, err
	}
	if len(comps) == 0 {
		item.Components = comps
		return &item, nil
	}

	ids := make([]string, len(comps))
	for i, c := range comps {
		ids[i] = c.ID
	}
	var opts []models.MenuItemComponentSizeOption
	if err := r.DB.SelectContext(ctx, &opts, `
		SELECT id, tenant_id, component_id, size_unit_id, qty, price::float8 AS price, is_active, created_at, updated_at
		FROM menu_item_component_size_option
		WHERE tenant_id = $1 AND component_id = ANY($2) AND is_active = TRUE
		ORDER BY created_at ASC, id ASC`, r.TenantID, pq.Array(ids)); err != nil {
		return nil, err
	}
	for i := range comps {
		for _, o := range opts {
			if o.ComponentID != comps[i].ID {
				continue
			}
			o.IsDefault = comps[i].DefaultSizeOptionID != nil && *comps[i].DefaultSizeOptionID == o.ID
			comps[i].SizeOptions = append(comps[i].SizeOptions, o)
		}
	}
	item.Components = comps
	return &item, nil
}

// GetSizeOptionPrices returns current prices keyed by size option id (REQOLINE003).
func (r *PostgreSQLOrderRepository) GetSizeOptionPrices(ctx context.Context, optionIDs []string) (map[string]float64, error) {
	out := make(map[string]float64, len(optionIDs))
	if len(optionIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		ID    string  `db:"id"`
		Price float64 `db:"price"`
	}
	if err := r.DB.SelectContext(ctx, &rows, `
		SELECT id, price::float8 AS price FROM menu_item_component_size_option
		WHERE tenant_id = $1 AND id = ANY($2)`, r.TenantID, pq.Array(optionIDs)); err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ID] = row.Price
	}
	return out, nil
}
