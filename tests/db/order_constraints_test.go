package db_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testdb "github.com/pareshvernekar/homecooked/tests/db"
)

// REQORDER001, REQORDER003, REQORDER004, REQOLINE001, REQOLINE002, REQPAY001, REQPAY002
func TestOrderSchemaConstraints(t *testing.T) {
	ctx := t.Context()
	helper, err := testdb.NewDatabaseHelper(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = helper.Terminate(ctx) })
	require.NoError(t, testdb.InitializeSchema(ctx, helper.DB))

	now := time.Now().UTC().UnixMilli()
	for _, id := range []string{"order-tenant-a", "order-tenant-b"} {
		_, err = helper.DB.ExecContext(ctx, `
			INSERT INTO tenant (id, name, description, is_active, created_at, updated_at)
			VALUES ($1, $1, '', TRUE, $2, $2)`, id, now)
		require.NoError(t, err)
	}
	const tenantA, tenantB = "order-tenant-a", "order-tenant-b"

	type scaffold struct{ menuID, itemID, componentID, optionID string }
	seed := func(tenantID string) scaffold {
		_, menuID, _, itemID, componentID := seedMenuScaffold(t, helper, tenantID, now)
		optionID := uuid.New().String()
		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO menu_item_component_size_option
			(id, tenant_id, component_id, size_unit_id, qty, price, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, 'su_serving', 1, 5.00, TRUE, $4, $4)`, optionID, tenantID, componentID, now)
		require.NoError(t, err)
		return scaffold{menuID, itemID, componentID, optionID}
	}
	a, b := seed(tenantA), seed(tenantB)

	insertOrder := func(tenantID, menuID, status string, override interface{}) (string, error) {
		id := uuid.New().String()
		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO customer_order
			(id, tenant_id, menu_id, customer_name, customer_phone, received_at, expected_at, status, total_override, created_at, updated_at)
			VALUES ($1,$2,$3,'Asha','555',$4,$4,$5,$6,$4,$4)`, id, tenantID, menuID, now, status, override)
		return id, err
	}
	insertItem := func(tenantID, orderID, menuItemID string, qty int, override interface{}) (string, error) {
		id := uuid.New().String()
		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO order_item (id, tenant_id, order_id, menu_item_id, quantity, unit_price_override, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$7)`, id, tenantID, orderID, menuItemID, qty, override, now)
		return id, err
	}
	insertSelection := func(tenantID, itemID, componentID, optionID string) error {
		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO order_item_component_selection
			(id, tenant_id, order_item_id, menu_item_component_id, size_option_id, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$6)`, uuid.New().String(), tenantID, itemID, componentID, optionID, now)
		return err
	}
	insertPayment := func(tenantID, orderID, mode string, amount float64) error {
		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO order_payment (id, tenant_id, order_id, mode, amount, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$6)`, uuid.New().String(), tenantID, orderID, mode, amount, now)
		return err
	}

	t.Run("REQORDER001_valid_order_and_status_values", func(t *testing.T) {
		for _, st := range []string{"RECEIVED", "IN_PROGRESS", "COMPLETE", "PICKEDUP"} {
			_, err := insertOrder(tenantA, a.menuID, st, nil)
			require.NoError(t, err, st)
		}
	})

	t.Run("REQORDER004_status_check", func(t *testing.T) {
		_, err := insertOrder(tenantA, a.menuID, "DELIVERED", nil)
		require.Error(t, err)
		_, err = insertOrder(tenantA, a.menuID, "received", nil)
		require.Error(t, err)
	})

	t.Run("REQORDER003_total_override_non_negative", func(t *testing.T) {
		_, err := insertOrder(tenantA, a.menuID, "RECEIVED", -1)
		require.Error(t, err)
		_, err = insertOrder(tenantA, a.menuID, "RECEIVED", 0)
		require.NoError(t, err)
	})

	t.Run("REQORDER001_menu_fk", func(t *testing.T) {
		_, err := insertOrder(tenantA, uuid.New().String(), "RECEIVED", nil)
		require.Error(t, err, "unknown menu must fail")
	})

	t.Run("tenant_isolation_cross_tenant_menu_fk", func(t *testing.T) {
		_, err := insertOrder(tenantA, b.menuID, "RECEIVED", nil)
		require.Error(t, err, "tenant A order must not reference tenant B menu")
	})

	t.Run("tenant_isolation_rows_scoped_by_tenant", func(t *testing.T) {
		idB, err := insertOrder(tenantB, b.menuID, "RECEIVED", nil)
		require.NoError(t, err)
		var count int
		require.NoError(t, helper.DB.GetContext(ctx, &count,
			`SELECT COUNT(*) FROM customer_order WHERE tenant_id = $1 AND id = $2`, tenantA, idB))
		assert.Equal(t, 0, count)
		require.NoError(t, helper.DB.GetContext(ctx, &count,
			`SELECT COUNT(*) FROM customer_order WHERE tenant_id = $1 AND id = $2`, tenantB, idB))
		assert.Equal(t, 1, count)
	})

	orderA, err := insertOrder(tenantA, a.menuID, "RECEIVED", nil)
	require.NoError(t, err)

	t.Run("REQOLINE001_quantity_positive", func(t *testing.T) {
		_, err := insertItem(tenantA, orderA, a.itemID, 0, nil)
		require.Error(t, err)
		_, err = insertItem(tenantA, orderA, a.itemID, -2, nil)
		require.Error(t, err)
	})

	t.Run("REQOLINE002_unit_price_override_non_negative", func(t *testing.T) {
		_, err := insertItem(tenantA, orderA, a.itemID, 1, -0.01)
		require.Error(t, err)
		_, err = insertItem(tenantA, orderA, a.itemID, 1, 0)
		require.NoError(t, err)
	})

	t.Run("REQOLINE001_item_fks", func(t *testing.T) {
		_, err := insertItem(tenantA, uuid.New().String(), a.itemID, 1, nil)
		require.Error(t, err, "unknown order must fail")
		_, err = insertItem(tenantA, orderA, uuid.New().String(), 1, nil)
		require.Error(t, err, "unknown menu item must fail")
		_, err = insertItem(tenantA, orderA, b.itemID, 1, nil)
		require.Error(t, err, "cross-tenant menu item must fail")
	})

	t.Run("REQOLINE001_selection_fks_and_unique", func(t *testing.T) {
		itemID, err := insertItem(tenantA, orderA, a.itemID, 1, nil)
		require.NoError(t, err)

		require.Error(t, insertSelection(tenantA, itemID, a.componentID, b.optionID), "cross-tenant option must fail")
		require.Error(t, insertSelection(tenantA, itemID, b.componentID, a.optionID), "cross-tenant component must fail")
		require.Error(t, insertSelection(tenantA, itemID, a.componentID, uuid.New().String()), "unknown option must fail")
		require.Error(t, insertSelection(tenantA, uuid.New().String(), a.componentID, a.optionID), "unknown line must fail")

		require.NoError(t, insertSelection(tenantA, itemID, a.componentID, a.optionID))
		require.Error(t, insertSelection(tenantA, itemID, a.componentID, a.optionID), "one selection per component per line")
	})

	t.Run("REQPAY001_mode_check", func(t *testing.T) {
		for _, mode := range []string{"cash", "credit", "paypal", "zelle", "venmo"} {
			require.NoError(t, insertPayment(tenantA, orderA, mode, 10), mode)
		}
		require.Error(t, insertPayment(tenantA, orderA, "bitcoin", 10))
		require.Error(t, insertPayment(tenantA, orderA, "CASH", 10))
	})

	t.Run("REQPAY001_amount_positive", func(t *testing.T) {
		require.Error(t, insertPayment(tenantA, orderA, "cash", 0))
		require.Error(t, insertPayment(tenantA, orderA, "cash", -5))
		require.NoError(t, insertPayment(tenantA, orderA, "cash", 0.01))
	})

	t.Run("REQPAY001_order_fk_and_tenant_isolation", func(t *testing.T) {
		require.Error(t, insertPayment(tenantA, uuid.New().String(), "cash", 5), "unknown order must fail")
		require.Error(t, insertPayment(tenantB, orderA, "cash", 5), "tenant B must not pay tenant A's order")
	})

	t.Run("REQPAY002_payments_sum_per_order", func(t *testing.T) {
		o, err := insertOrder(tenantA, a.menuID, "RECEIVED", nil)
		require.NoError(t, err)
		require.NoError(t, insertPayment(tenantA, o, "cash", 20))
		require.NoError(t, insertPayment(tenantA, o, "venmo", 30))
		var paid float64
		require.NoError(t, helper.DB.GetContext(ctx, &paid,
			`SELECT COALESCE(SUM(amount),0)::float8 FROM order_payment WHERE tenant_id=$1 AND order_id=$2`, tenantA, o))
		assert.Equal(t, 50.0, paid)
	})
}
