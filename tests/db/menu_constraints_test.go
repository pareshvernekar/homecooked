package db_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testdb "github.com/pareshvernekar/homecooked/tests/db"
)

// REQSIZE001, REQSIZE002, REQITEM003, REQITEM004, REQMENU009
func TestMenuSchemaConstraints(t *testing.T) {
	ctx := t.Context()
	helper, err := testdb.NewDatabaseHelper(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = helper.Terminate(ctx) })

	require.NoError(t, testdb.InitializeSchema(ctx, helper.DB))

	tenantID := "constraint-tenant"
	now := time.Now().UTC().UnixMilli()
	_, err = helper.DB.ExecContext(ctx, `
		INSERT INTO tenant (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, 'Constraint Tenant', '', TRUE, $2, $2)
	`, tenantID, now)
	require.NoError(t, err)

	t.Run("REQSIZE001_system_units_seeded", func(t *testing.T) {
		var count int
		err := helper.DB.GetContext(ctx, &count, `
			SELECT COUNT(*) FROM size_unit WHERE is_system = TRUE AND code IN
			('serving','tray','piece','dozen','kg','g','liter')
		`)
		require.NoError(t, err)
		assert.Equal(t, 7, count)
	})

	t.Run("REQSIZE002_tenant_custom_code_unique", func(t *testing.T) {
		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO size_unit (id, tenant_id, code, display_name, is_system, is_active, created_at, updated_at)
			VALUES ($1, $2, 'party-pan', 'Party Pan', FALSE, TRUE, $3, $3)
		`, uuid.New().String(), tenantID, now)
		require.NoError(t, err)

		_, err = helper.DB.ExecContext(ctx, `
			INSERT INTO size_unit (id, tenant_id, code, display_name, is_system, is_active, created_at, updated_at)
			VALUES ($1, $2, 'party-pan', 'Party Pan 2', FALSE, TRUE, $3, $3)
		`, uuid.New().String(), tenantID, now)
		require.Error(t, err, "duplicate custom code for tenant must fail")
	})

	t.Run("REQITEM004_size_option_price_and_qty_checks", func(t *testing.T) {
		foodID, menuID, categoryID, itemID, componentID := seedMenuScaffold(t, helper, tenantID, now)

		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO menu_item_component_size_option
			(id, tenant_id, component_id, size_unit_id, qty, price, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, 'su_tray', 0, 10.00, TRUE, $4, $4)
		`, uuid.New().String(), tenantID, componentID, now)
		require.Error(t, err, "qty <= 0 must fail")

		_, err = helper.DB.ExecContext(ctx, `
			INSERT INTO menu_item_component_size_option
			(id, tenant_id, component_id, size_unit_id, qty, price, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, 'su_tray', 1, -1.00, TRUE, $4, $4)
		`, uuid.New().String(), tenantID, componentID, now)
		require.Error(t, err, "price < 0 must fail")

		optID := uuid.New().String()
		_, err = helper.DB.ExecContext(ctx, `
			INSERT INTO menu_item_component_size_option
			(id, tenant_id, component_id, size_unit_id, qty, price, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, 'su_tray', 1, 12.50, TRUE, $4, $4)
		`, optID, tenantID, componentID, now)
		require.NoError(t, err)

		_, err = helper.DB.ExecContext(ctx, `
			UPDATE menu_item_component SET default_size_option_id = $1
			WHERE tenant_id = $2 AND id = $3
		`, optID, tenantID, componentID)
		require.NoError(t, err)

		_ = foodID
		_ = menuID
		_ = categoryID
		_ = itemID
	})

	t.Run("REQMENU009_weekly_published_start_date_unique", func(t *testing.T) {
		start := "2026-10-05"
		end := "2026-10-11"
		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO menu (id, tenant_id, name, menu_type, status, is_active, start_date, end_date, created_at, updated_at)
			VALUES ($1, $2, 'Week A', 'weekly', 'published', TRUE, $3::date, $4::date, $5, $5)
		`, uuid.New().String(), tenantID, start, end, now)
		require.NoError(t, err)

		_, err = helper.DB.ExecContext(ctx, `
			INSERT INTO menu (id, tenant_id, name, menu_type, status, is_active, start_date, end_date, created_at, updated_at)
			VALUES ($1, $2, 'Week A Dup', 'weekly', 'published', TRUE, $3::date, $4::date, $5, $5)
		`, uuid.New().String(), tenantID, start, end, now)
		require.Error(t, err, "second active published weekly with same start_date must fail")

		// Drafts may share start_date.
		_, err = helper.DB.ExecContext(ctx, `
			INSERT INTO menu (id, tenant_id, name, menu_type, status, is_active, start_date, end_date, created_at, updated_at)
			VALUES ($1, $2, 'Week A Draft', 'weekly', 'draft', TRUE, $3::date, $4::date, $5, $5)
		`, uuid.New().String(), tenantID, start, end, now)
		require.NoError(t, err)
	})

	t.Run("food_item_has_no_price_column_REQFOOD001", func(t *testing.T) {
		var exists bool
		err := helper.DB.GetContext(ctx, &exists, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'food_item' AND column_name = 'price'
			)
		`)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func seedMenuScaffold(t *testing.T, helper *testdb.DatabaseHelper, tenantID string, now int64) (foodID, menuID, categoryID, itemID, componentID string) {
	t.Helper()
	foodCatID := uuid.New().String()
	foodID = uuid.New().String()
	menuID = uuid.New().String()
	categoryID = uuid.New().String()
	itemID = uuid.New().String()
	componentID = uuid.New().String()

	_, err := helper.DB.ExecContext(t.Context(), `
		INSERT INTO food_category (id, tenant_id, name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, TRUE, $4, $4)
	`, foodCatID, tenantID, "mains-"+uuid.New().String()[:8], now)
	require.NoError(t, err)

	_, err = helper.DB.ExecContext(t.Context(), `
		INSERT INTO food_item (id, tenant_id, name, availability_status, category_id, is_active, created_at, updated_at)
		VALUES ($1, $2, 'Rice', 'available', $3, TRUE, $4, $4)
	`, foodID, tenantID, foodCatID, now)
	require.NoError(t, err)

	_, err = helper.DB.ExecContext(t.Context(), `
		INSERT INTO menu (id, tenant_id, name, menu_type, status, is_active, menu_date, created_at, updated_at)
		VALUES ($1, $2, 'Daily', 'daily', 'draft', TRUE, CURRENT_DATE, $3, $3)
	`, menuID, tenantID, now)
	require.NoError(t, err)

	_, err = helper.DB.ExecContext(t.Context(), `
		INSERT INTO menu_category (id, tenant_id, menu_id, name, sequence, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, 'Mains', 1, TRUE, $4, $4)
	`, categoryID, tenantID, menuID, now)
	require.NoError(t, err)

	_, err = helper.DB.ExecContext(t.Context(), `
		INSERT INTO menu_item (id, tenant_id, menu_id, category_id, kind, name, sequence, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'simple', 'Rice Bowl', 1, TRUE, $5, $5)
	`, itemID, tenantID, menuID, categoryID, now)
	require.NoError(t, err)

	_, err = helper.DB.ExecContext(t.Context(), `
		INSERT INTO menu_item_component (id, tenant_id, menu_item_id, food_item_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, TRUE, $5, $5)
	`, componentID, tenantID, itemID, foodID, now)
	require.NoError(t, err)

	return foodID, menuID, categoryID, itemID, componentID
}
