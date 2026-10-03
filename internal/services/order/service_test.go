package order_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/pareshvernekar/homecooked/internal/errors"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/services/menu"
	"github.com/pareshvernekar/homecooked/internal/services/menuitem"
	"github.com/pareshvernekar/homecooked/internal/services/order"
	"github.com/pareshvernekar/homecooked/internal/services/sizeunit"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
)

// tenantEnv wires repositories/services for one tenant on a shared DB.
type tenantEnv struct {
	tenantID  string
	orderRepo *repository.PostgreSQLOrderRepository
	menuSvc   *menu.Service
	itemSvc   *menuitem.Service
	svc       *order.Service

	// Published daily menu seeded for the tenant.
	menuID string
	// Simple item "Rice Bowl": one component, options serving=5 (default) and tray=20.
	riceItemID, riceCompID, riceServing, riceTray string
	// Combo "Thali": comp A options serving=5 (default) and tray=8; comp B serving=7 (default).
	thaliItemID, thaliCompA, thaliCompB, thaliAServing, thaliATray, thaliBServing string
}

func newDB(t *testing.T) *sqlx.DB {
	t.Helper()
	ctx := context.Background()
	helper, err := testdb.NewDatabaseHelper(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = helper.Terminate(ctx) })
	require.NoError(t, testdb.InitializeSchema(ctx, helper.DB))
	return helper.DB
}

func newTenant(t *testing.T, db *sqlx.DB, tenantID string) *tenantEnv {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC().UnixMilli()
	_, err := db.ExecContext(ctx, `INSERT INTO tenant (id, name, is_active, created_at, updated_at) VALUES ($1,$1,TRUE,$2,$2)`, tenantID, now)
	require.NoError(t, err)

	catID := uuid.New().String()
	_, err = db.ExecContext(ctx, `
		INSERT INTO food_category (id, tenant_id, name, is_active, created_at, updated_at)
		VALUES ($1,$2,'mains',TRUE,$3,$3)`, catID, tenantID, now)
	require.NoError(t, err)
	foodIDs := make([]string, 3)
	for i, name := range []string{"Rice", "Dal", "Curry"} {
		foodIDs[i] = uuid.New().String()
		_, err = db.ExecContext(ctx, `
			INSERT INTO food_item (id, tenant_id, name, availability_status, category_id, is_active, created_at, updated_at)
			VALUES ($1,$2,$3,'available',$4,TRUE,$5,$5)`, foodIDs[i], tenantID, name, catID, now)
		require.NoError(t, err)
	}

	l := logger.NewLogger()
	menuRepo := repository.NewMenuRepository(db, l, tenantID)
	itemRepo := repository.NewMenuItemRepository(db, l, tenantID)
	sizeSvc := sizeunit.NewService(repository.NewSizeUnitRepository(db, l, tenantID), l)
	menuSvc := menu.NewService(menuRepo, itemRepo, l)
	itemSvc := menuitem.NewService(itemRepo, menuSvc, sizeSvc, l)
	orderRepo := repository.NewOrderRepository(db, l, tenantID)

	e := &tenantEnv{
		tenantID: tenantID, orderRepo: orderRepo, menuSvc: menuSvc, itemSvc: itemSvc,
		svc: order.NewService(orderRepo, menuSvc, l),
	}

	date := "2026-11-10"
	m, err := menuSvc.Create(ctx, tenantID, &models.MenuCreateRequest{
		Name: "Daily", MenuType: models.MenuTypeDaily, MenuDate: &date,
		Categories: []models.MenuCategoryInput{{Name: "Mains", Sequence: 1}},
	})
	require.NoError(t, err)
	e.menuID = m.ID
	tree, err := menuSvc.GetTree(ctx, tenantID, m.ID)
	require.NoError(t, err)
	menuCat := tree.Categories[0].ID

	idx0 := 0
	_, err = itemSvc.AddItem(ctx, tenantID, m.ID, &models.MenuItemCreateRequest{
		CategoryID: menuCat, Kind: models.MenuItemKindSimple, Name: "Rice Bowl",
		Components: []models.MenuItemComponentInput{
			{FoodItemID: foodIDs[0], DefaultSizeIndex: &idx0, SizeOptions: []models.MenuItemSizeOptionInput{
				{SizeUnitID: "su_serving", Qty: 1, Price: 5},
				{SizeUnitID: "su_tray", Qty: 1, Price: 20},
			}},
		},
	})
	require.NoError(t, err)
	_, err = itemSvc.AddItem(ctx, tenantID, m.ID, &models.MenuItemCreateRequest{
		CategoryID: menuCat, Kind: models.MenuItemKindCombo, Name: "Thali",
		Components: []models.MenuItemComponentInput{
			{FoodItemID: foodIDs[0], DefaultSizeIndex: &idx0, SizeOptions: []models.MenuItemSizeOptionInput{
				{SizeUnitID: "su_serving", Qty: 1, Price: 5},
				{SizeUnitID: "su_tray", Qty: 1, Price: 8},
			}},
			{FoodItemID: foodIDs[1], DefaultSizeIndex: &idx0, SizeOptions: []models.MenuItemSizeOptionInput{
				{SizeUnitID: "su_serving", Qty: 1, Price: 7},
			}},
		},
	})
	require.NoError(t, err)
	require.NoError(t, menuSvc.Publish(ctx, tenantID, m.ID))

	tree, err = menuSvc.GetTree(ctx, tenantID, m.ID)
	require.NoError(t, err)
	for _, it := range tree.Categories[0].Items {
		switch it.Name {
		case "Rice Bowl":
			e.riceItemID = it.ID
			e.riceCompID = it.Components[0].ID
			e.riceServing = it.Components[0].SizeOptions[0].ID
			e.riceTray = it.Components[0].SizeOptions[1].ID
		case "Thali":
			e.thaliItemID = it.ID
			e.thaliCompA = it.Components[0].ID
			e.thaliAServing = it.Components[0].SizeOptions[0].ID
			e.thaliATray = it.Components[0].SizeOptions[1].ID
			e.thaliCompB = it.Components[1].ID
			e.thaliBServing = it.Components[0+1].SizeOptions[0].ID
		}
	}
	require.NotEmpty(t, e.riceItemID)
	require.NotEmpty(t, e.thaliItemID)
	return e
}

func ptr[T any](v T) *T { return &v }

func (e *tenantEnv) createOrder(t *testing.T) *models.CustomerOrder {
	t.Helper()
	o, err := e.svc.Create(context.Background(), e.tenantID, &models.OrderCreateRequest{
		MenuID: e.menuID, CustomerName: "Asha", CustomerPhone: "555-0100",
		ExpectedAt: ptr(time.Now().Add(2 * time.Hour).UnixMilli()),
	})
	require.NoError(t, err)
	return o
}

func (e *tenantEnv) addRice(t *testing.T, orderID string, qty int) *models.OrderItem {
	t.Helper()
	item, err := e.svc.AddLine(context.Background(), e.tenantID, orderID, &models.OrderLineCreateRequest{
		MenuItemID: e.riceItemID, Quantity: qty,
	})
	require.NoError(t, err)
	return item
}

func (e *tenantEnv) setPrice(t *testing.T, itemID, compID, optID string, price float64) {
	t.Helper()
	require.NoError(t, e.itemSvc.UpdateSizeOption(context.Background(), e.tenantID, e.menuID, itemID, compID, optID,
		&models.SizeOptionUpdateRequest{Price: &price}))
}

func (e *tenantEnv) get(t *testing.T, id string) *models.OrderDetail {
	t.Helper()
	d, err := e.svc.Get(context.Background(), e.tenantID, id)
	require.NoError(t, err)
	return d
}

func (e *tenantEnv) pay(t *testing.T, id, mode string, amount float64) {
	t.Helper()
	_, err := e.svc.RecordPayment(context.Background(), e.tenantID, id, &models.PaymentCreateRequest{Mode: mode, Amount: amount})
	require.NoError(t, err)
}

// toReady walks an order RECEIVED → ACCEPTED → IN_PROGRESS → READY via lifecycle actions.
func (e *tenantEnv) toReady(t *testing.T, id string) {
	t.Helper()
	ctx := context.Background()
	_, err := e.svc.Accept(ctx, e.tenantID, id)
	require.NoError(t, err)
	_, err = e.svc.StartPreparing(ctx, e.tenantID, id)
	require.NoError(t, err)
	_, err = e.svc.Ready(ctx, e.tenantID, id)
	require.NoError(t, err)
}

// pickup walks the order to READY then picks it up.
func (e *tenantEnv) pickup(t *testing.T, id string) *models.OrderDetail {
	t.Helper()
	e.toReady(t, id)
	d, err := e.svc.Pickup(context.Background(), e.tenantID, id)
	require.NoError(t, err)
	return d
}

func requireStatus(t *testing.T, err error, code int) {
	t.Helper()
	require.Error(t, err)
	var se *apperrors.ServiceError
	require.ErrorAs(t, err, &se)
	assert.Equal(t, code, se.GetStatusCode())
}

// REQORDER001, REQORDER004S01, REQORDER005
func TestCreate_Gates(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	a := newTenant(t, db, "tenant-a")
	b := newTenant(t, db, "tenant-b")
	expected := ptr(time.Now().Add(time.Hour).UnixMilli())

	t.Run("REQORDER001S01_success_defaults", func(t *testing.T) {
		before := time.Now().UnixMilli()
		o, err := a.svc.Create(ctx, a.tenantID, &models.OrderCreateRequest{
			MenuID: a.menuID, CustomerName: " Asha ", CustomerPhone: "555", ExpectedAt: expected,
			CustomizationText: ptr("nut-free"),
		})
		require.NoError(t, err)
		assert.NotEmpty(t, o.ID)
		assert.Equal(t, "tenant-a", o.TenantID)
		assert.Equal(t, models.OrderStatusReceived, o.Status) // REQORDER004S01
		assert.True(t, o.IsActive)
		assert.Equal(t, "Asha", o.CustomerName)
		assert.GreaterOrEqual(t, o.ReceivedAt, before)
		assert.Equal(t, *expected, o.ExpectedAt)
		require.NotNil(t, o.CustomizationText)
		assert.Equal(t, "nut-free", *o.CustomizationText) // REQORDER005S01
	})

	t.Run("REQORDER001S01_client_received_at", func(t *testing.T) {
		o, err := a.svc.Create(ctx, a.tenantID, &models.OrderCreateRequest{
			MenuID: a.menuID, CustomerName: "N", CustomerPhone: "1", ExpectedAt: expected, ReceivedAt: ptr(int64(1234)),
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1234), o.ReceivedAt)
	})

	t.Run("REQORDER001S02_missing_header_fields", func(t *testing.T) {
		for name, req := range map[string]*models.OrderCreateRequest{
			"name":     {MenuID: a.menuID, CustomerPhone: "1", ExpectedAt: expected},
			"phone":    {MenuID: a.menuID, CustomerName: "n", ExpectedAt: expected},
			"expected": {MenuID: a.menuID, CustomerName: "n", CustomerPhone: "1"},
			"blank":    {MenuID: a.menuID, CustomerName: "  ", CustomerPhone: "1", ExpectedAt: expected},
		} {
			_, err := a.svc.Create(ctx, a.tenantID, req)
			requireStatus(t, err, 400)
			_ = name
		}
	})

	t.Run("REQORDER001S03_draft_menu_rejected", func(t *testing.T) {
		date := "2026-11-11"
		draft, err := a.menuSvc.Create(ctx, a.tenantID, &models.MenuCreateRequest{
			Name: "Draft", MenuType: models.MenuTypeDaily, MenuDate: &date,
		})
		require.NoError(t, err)
		_, err = a.svc.Create(ctx, a.tenantID, &models.OrderCreateRequest{
			MenuID: draft.ID, CustomerName: "n", CustomerPhone: "1", ExpectedAt: expected,
		})
		requireStatus(t, err, 400)
	})

	t.Run("REQORDER001S03_weekly_published_rejected", func(t *testing.T) {
		start, end := "2026-12-07", "2026-12-13"
		wk, err := a.menuSvc.Create(ctx, a.tenantID, &models.MenuCreateRequest{
			Name: "Week", MenuType: models.MenuTypeWeekly, StartDate: &start, EndDate: &end,
			Categories: []models.MenuCategoryInput{{Name: "Mains", Sequence: 1}},
		})
		require.NoError(t, err)
		tree, err := a.menuSvc.GetTree(ctx, a.tenantID, wk.ID)
		require.NoError(t, err)
		foodID := ""
		require.NoError(t, db.GetContext(ctx, &foodID, `SELECT id FROM food_item WHERE tenant_id=$1 LIMIT 1`, a.tenantID))
		idx0 := 0
		_, err = a.itemSvc.AddItem(ctx, a.tenantID, wk.ID, &models.MenuItemCreateRequest{
			CategoryID: tree.Categories[0].ID, Kind: models.MenuItemKindSimple, Name: "Rice",
			Components: []models.MenuItemComponentInput{{FoodItemID: foodID, DefaultSizeIndex: &idx0,
				SizeOptions: []models.MenuItemSizeOptionInput{{SizeUnitID: "su_serving", Qty: 1, Price: 5}}}},
		})
		require.NoError(t, err)
		require.NoError(t, a.menuSvc.Publish(ctx, a.tenantID, wk.ID))

		_, err = a.svc.Create(ctx, a.tenantID, &models.OrderCreateRequest{
			MenuID: wk.ID, CustomerName: "n", CustomerPhone: "1", ExpectedAt: expected,
		})
		requireStatus(t, err, 400)
	})

	t.Run("REQORDER001S03_unpublished_or_inactive_menu_rejected", func(t *testing.T) {
		require.NoError(t, a.menuSvc.Unpublish(ctx, a.tenantID, a.menuID))
		_, err := a.svc.Create(ctx, a.tenantID, &models.OrderCreateRequest{
			MenuID: a.menuID, CustomerName: "n", CustomerPhone: "1", ExpectedAt: expected,
		})
		requireStatus(t, err, 400)
		require.NoError(t, a.menuSvc.Deactivate(ctx, a.tenantID, a.menuID))
		_, err = a.svc.Create(ctx, a.tenantID, &models.OrderCreateRequest{
			MenuID: a.menuID, CustomerName: "n", CustomerPhone: "1", ExpectedAt: expected,
		})
		requireStatus(t, err, 404)
	})

	t.Run("REQORDER001S03_cross_tenant_menu_rejected", func(t *testing.T) {
		_, err := a.svc.Create(ctx, a.tenantID, &models.OrderCreateRequest{
			MenuID: b.menuID, CustomerName: "n", CustomerPhone: "1", ExpectedAt: expected,
		})
		requireStatus(t, err, 404)
	})
}

// REQORDER002
func TestListGet_TenantScoping(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	a := newTenant(t, db, "tenant-a")
	b := newTenant(t, db, "tenant-b")

	oa := a.createOrder(t)
	ob := b.createOrder(t)
	inactive := a.createOrder(t)
	require.NoError(t, a.orderRepo.DeactivateOrder(ctx, inactive.ID))

	t.Run("REQORDER002S01_list_tenant_only", func(t *testing.T) {
		list, err := a.svc.List(ctx, a.tenantID)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, oa.ID, list[0].ID)
		listB, err := b.svc.List(ctx, b.tenantID)
		require.NoError(t, err)
		require.Len(t, listB, 1)
		assert.Equal(t, ob.ID, listB[0].ID)
	})

	t.Run("REQORDER002S03_not_found_cross_tenant_or_inactive", func(t *testing.T) {
		_, err := a.svc.Get(ctx, a.tenantID, ob.ID)
		requireStatus(t, err, 404)
		_, err = a.svc.Get(ctx, a.tenantID, inactive.ID)
		requireStatus(t, err, 404)
		_, err = a.svc.Get(ctx, a.tenantID, uuid.New().String())
		requireStatus(t, err, 404)
	})

	t.Run("REQPAY001_cross_tenant_payment_rejected", func(t *testing.T) {
		_, err := a.svc.RecordPayment(ctx, a.tenantID, ob.ID, &models.PaymentCreateRequest{Mode: "cash", Amount: 5})
		requireStatus(t, err, 404)
		_, err = a.svc.ListPayments(ctx, a.tenantID, ob.ID)
		requireStatus(t, err, 404)
	})
}

// REQOLINE001, REQOLINE002, REQOLINE003, REQOLINE004, REQORDER002S02
func TestLines_SelectionsAndLivePricing(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	e := newTenant(t, db, "tenant-a")
	o := e.createOrder(t)

	t.Run("REQOLINE001S01_defaults", func(t *testing.T) {
		line := e.addRice(t, o.ID, 2)
		require.Len(t, line.Selections, 1)
		assert.Equal(t, e.riceServing, line.Selections[0].SizeOptionID)
		assert.Equal(t, 5.0, line.UnitPrice)
		assert.Equal(t, 10.0, line.ExtendedAmount)

		thali, err := e.svc.AddLine(ctx, e.tenantID, o.ID, &models.OrderLineCreateRequest{MenuItemID: e.thaliItemID, Quantity: 1})
		require.NoError(t, err)
		require.Len(t, thali.Selections, 2)
		assert.Equal(t, 12.0, thali.UnitPrice) // 5 + 7
	})

	t.Run("REQOLINE001S02_explicit_selection", func(t *testing.T) {
		thali, err := e.svc.AddLine(ctx, e.tenantID, o.ID, &models.OrderLineCreateRequest{
			MenuItemID: e.thaliItemID, Quantity: 1,
			Selections: []models.OrderSelectionInput{{MenuItemComponentID: e.thaliCompA, SizeOptionID: e.thaliATray}},
		})
		require.NoError(t, err)
		assert.Equal(t, 15.0, thali.UnitPrice) // 8 + 7 (comp B default)
	})

	t.Run("REQOLINE001S03_invalid_selection_rejected", func(t *testing.T) {
		before := len(e.get(t, o.ID).Lines)
		// option from a different component
		_, err := e.svc.AddLine(ctx, e.tenantID, o.ID, &models.OrderLineCreateRequest{
			MenuItemID: e.thaliItemID, Quantity: 1,
			Selections: []models.OrderSelectionInput{{MenuItemComponentID: e.thaliCompA, SizeOptionID: e.riceTray}},
		})
		requireStatus(t, err, 400)
		// unknown option
		_, err = e.svc.AddLine(ctx, e.tenantID, o.ID, &models.OrderLineCreateRequest{
			MenuItemID: e.riceItemID, Quantity: 1,
			Selections: []models.OrderSelectionInput{{MenuItemComponentID: e.riceCompID, SizeOptionID: uuid.New().String()}},
		})
		requireStatus(t, err, 400)
		// component not on item
		_, err = e.svc.AddLine(ctx, e.tenantID, o.ID, &models.OrderLineCreateRequest{
			MenuItemID: e.riceItemID, Quantity: 1,
			Selections: []models.OrderSelectionInput{{MenuItemComponentID: e.thaliCompB, SizeOptionID: e.thaliBServing}},
		})
		requireStatus(t, err, 400)
		// menu item not on the order's menu
		_, err = e.svc.AddLine(ctx, e.tenantID, o.ID, &models.OrderLineCreateRequest{MenuItemID: uuid.New().String(), Quantity: 1})
		requireStatus(t, err, 400)
		// non-positive quantity
		_, err = e.svc.AddLine(ctx, e.tenantID, o.ID, &models.OrderLineCreateRequest{MenuItemID: e.riceItemID, Quantity: 0})
		requireStatus(t, err, 400)
		assert.Len(t, e.get(t, o.ID).Lines, before)
	})

	o2 := e.createOrder(t)
	line := e.addRice(t, o2.ID, 2)

	t.Run("REQOLINE002S01_override_and_customization", func(t *testing.T) {
		got, err := e.svc.UpdateLine(ctx, e.tenantID, o2.ID, line.ID, &models.OrderLineUpdateRequest{
			CustomizationText: ptr("spicy"),
			UnitPriceOverride: models.NullableFloat64{Set: true, Value: ptr(10.0)},
		})
		require.NoError(t, err)
		assert.Equal(t, 10.0, got.UnitPrice)
		d := e.get(t, o2.ID)
		assert.Equal(t, 20.0, d.Subtotal)
		assert.Equal(t, 20.0, d.ChargedTotal)
		require.NotNil(t, d.Lines[0].CustomizationText)
		assert.Equal(t, "spicy", *d.Lines[0].CustomizationText)
	})

	t.Run("REQOLINE002S02_clear_override_restores_live", func(t *testing.T) {
		got, err := e.svc.UpdateLine(ctx, e.tenantID, o2.ID, line.ID, &models.OrderLineUpdateRequest{
			UnitPriceOverride: models.NullableFloat64{Set: true, Value: nil},
		})
		require.NoError(t, err)
		assert.Equal(t, 5.0, got.UnitPrice)
		assert.Equal(t, 10.0, e.get(t, o2.ID).ChargedTotal)
	})

	t.Run("REQOLINE003S01_live_price_change_moves_total", func(t *testing.T) {
		e.setPrice(t, e.riceItemID, e.riceCompID, e.riceServing, 6)
		d := e.get(t, o2.ID)
		assert.Equal(t, 6.0, d.Lines[0].UnitPrice)
		assert.Equal(t, 12.0, d.ChargedTotal)
		e.setPrice(t, e.riceItemID, e.riceCompID, e.riceServing, 5)
	})

	t.Run("REQOLINE003S02_order_override_wins", func(t *testing.T) {
		d, err := e.svc.Update(ctx, e.tenantID, o2.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: ptr(7.5)},
		})
		require.NoError(t, err)
		assert.Equal(t, 10.0, d.Subtotal)
		assert.Equal(t, 7.5, d.ChargedTotal) // REQORDER003S02

		d, err = e.svc.Update(ctx, e.tenantID, o2.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: nil},
		})
		require.NoError(t, err)
		assert.Equal(t, 10.0, d.ChargedTotal)
	})

	t.Run("REQOLINE004S01_update_selection", func(t *testing.T) {
		got, err := e.svc.UpdateLine(ctx, e.tenantID, o2.ID, line.ID, &models.OrderLineUpdateRequest{
			Selections: []models.OrderSelectionInput{{MenuItemComponentID: e.riceCompID, SizeOptionID: e.riceTray}},
		})
		require.NoError(t, err)
		require.Len(t, got.Selections, 1)
		assert.Equal(t, e.riceTray, got.Selections[0].SizeOptionID)
		assert.Equal(t, 40.0, e.get(t, o2.ID).ChargedTotal) // 20 x 2

		_, err = e.svc.UpdateLine(ctx, e.tenantID, o2.ID, line.ID, &models.OrderLineUpdateRequest{
			Selections: []models.OrderSelectionInput{{MenuItemComponentID: e.riceCompID, SizeOptionID: e.thaliATray}},
		})
		requireStatus(t, err, 400)
	})

	t.Run("REQOLINE004S02_remove_line", func(t *testing.T) {
		require.NoError(t, e.svc.RemoveLine(ctx, e.tenantID, o2.ID, line.ID))
		d := e.get(t, o2.ID)
		assert.Empty(t, d.Lines)
		assert.Equal(t, 0.0, d.ChargedTotal)
		requireStatus(t, e.svc.RemoveLine(ctx, e.tenantID, o2.ID, line.ID), 404)
	})
}

// REQORDER003, REQORDER004, REQOLINE003S03, REQOLINE001S04, REQPAY003 (unpaid PICKEDUP)
func TestFreezeAtPickedUp(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	e := newTenant(t, db, "tenant-a")

	t.Run("REQORDER003S01_update_header", func(t *testing.T) {
		o := e.createOrder(t)
		newExpected := time.Now().Add(5 * time.Hour).UnixMilli()
		d, err := e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{
			ExpectedAt: &newExpected, CustomizationText: ptr("nut-free"), CustomerName: ptr("Bina"),
		})
		require.NoError(t, err)
		assert.Equal(t, newExpected, d.ExpectedAt)
		require.NotNil(t, d.CustomizationText)
		assert.Equal(t, "nut-free", *d.CustomizationText)
		assert.Equal(t, "Bina", d.CustomerName)

		_, err = e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{CustomerName: ptr(" ")})
		requireStatus(t, err, 400)
		_, err = e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: ptr(-1.0)},
		})
		requireStatus(t, err, 400)
	})

	t.Run("REQORDER004S02_freeze_ignores_later_price_changes", func(t *testing.T) {
		o := e.createOrder(t)
		e.addRice(t, o.ID, 2) // live 5 x 2 = 10
		before := time.Now().UnixMilli()
		d := e.pickup(t, o.ID)
		assert.Equal(t, models.OrderStatusPickedUp, d.Status)
		require.NotNil(t, d.PickedupAt)
		assert.GreaterOrEqual(t, *d.PickedupAt, before)
		require.NotNil(t, d.FrozenTotal)
		assert.Equal(t, 10.0, *d.FrozenTotal)
		assert.Equal(t, 10.0, d.ChargedTotal)

		e.setPrice(t, e.riceItemID, e.riceCompID, e.riceServing, 9) // REQOLINE003S03
		defer e.setPrice(t, e.riceItemID, e.riceCompID, e.riceServing, 5)
		d = e.get(t, o.ID)
		assert.Equal(t, 10.0, d.ChargedTotal)
		assert.Equal(t, 5.0, d.Lines[0].UnitPrice)
		assert.Equal(t, 10.0, d.Lines[0].ExtendedAmount)
	})

	t.Run("REQORDER004S02_freeze_uses_overrides", func(t *testing.T) {
		o := e.createOrder(t)
		e.addRice(t, o.ID, 2)
		_, err := e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: ptr(8.0)},
		})
		require.NoError(t, err)
		d := e.pickup(t, o.ID)
		assert.Equal(t, 8.0, d.ChargedTotal)
		assert.Equal(t, 8.0, *d.FrozenTotal)
	})

	t.Run("REQORDER004_pickedup_at_supplied_is_kept", func(t *testing.T) {
		o := e.createOrder(t)
		at := int64(1_700_000_000_000)
		_, err := e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{PickedupAt: &at})
		require.NoError(t, err)
		d := e.pickup(t, o.ID)
		assert.Equal(t, at, *d.PickedupAt)
	})

	t.Run("REQORDER003S03_money_overrides_rejected_after_pickedup", func(t *testing.T) {
		o := e.createOrder(t)
		e.addRice(t, o.ID, 1)
		e.pickup(t, o.ID)

		_, err := e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: ptr(1.0)},
		})
		requireStatus(t, err, 400)
		_, err = e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: nil},
		})
		requireStatus(t, err, 400)

		// pickedup_at may still be corrected
		fix := int64(1_700_000_123_000)
		d, err := e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{PickedupAt: &fix})
		require.NoError(t, err)
		assert.Equal(t, fix, *d.PickedupAt)
		assert.Equal(t, 5.0, d.ChargedTotal)
	})

	t.Run("REQOLINE001S04_REQOLINE004_line_mutations_rejected_after_pickedup", func(t *testing.T) {
		o := e.createOrder(t)
		line := e.addRice(t, o.ID, 1)
		e.pickup(t, o.ID)

		_, err := e.svc.AddLine(ctx, e.tenantID, o.ID, &models.OrderLineCreateRequest{MenuItemID: e.riceItemID, Quantity: 1})
		requireStatus(t, err, 400)
		_, err = e.svc.UpdateLine(ctx, e.tenantID, o.ID, line.ID, &models.OrderLineUpdateRequest{
			UnitPriceOverride: models.NullableFloat64{Set: true, Value: ptr(1.0)},
		})
		requireStatus(t, err, 400) // REQOLINE002
		_, err = e.svc.UpdateLine(ctx, e.tenantID, o.ID, line.ID, &models.OrderLineUpdateRequest{CustomizationText: ptr("x")})
		requireStatus(t, err, 400)
		requireStatus(t, e.svc.RemoveLine(ctx, e.tenantID, o.ID, line.ID), 400)
		assert.Len(t, e.get(t, o.ID).Lines, 1)
	})

	t.Run("REQORDER004S03_unpaid_pickedup_allowed", func(t *testing.T) {
		o := e.createOrder(t)
		e.addRice(t, o.ID, 1)
		d := e.pickup(t, o.ID)
		assert.False(t, d.PaymentReceived)
		assert.Equal(t, 5.0, d.Balance)
	})
}

// REQPAY001–REQPAY004
func TestPayments(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	e := newTenant(t, db, "tenant-a")

	// order with charged total 50: 5 x 10
	newOrder50 := func(t *testing.T) *models.CustomerOrder {
		o := e.createOrder(t)
		e.addRice(t, o.ID, 10)
		return o
	}

	t.Run("REQPAY001S02_invalid_mode_or_amount", func(t *testing.T) {
		o := newOrder50(t)
		for _, req := range []*models.PaymentCreateRequest{
			{Mode: "bitcoin", Amount: 5},
			{Mode: "", Amount: 5},
			{Mode: "cash", Amount: 0},
			{Mode: "cash", Amount: -3},
			{Mode: "cash", Amount: 0.001},
		} {
			_, err := e.svc.RecordPayment(ctx, e.tenantID, o.ID, req)
			requireStatus(t, err, 400)
		}
		ps, err := e.svc.ListPayments(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		assert.Empty(t, ps)
	})

	t.Run("REQPAY001S01_REQPAY002S01_partial_payment", func(t *testing.T) {
		o := newOrder50(t)
		p, err := e.svc.RecordPayment(ctx, e.tenantID, o.ID, &models.PaymentCreateRequest{
			Mode: "cash", Amount: 20, ReferenceText: ptr("till 3"),
		})
		require.NoError(t, err)
		assert.NotEmpty(t, p.ID)
		d := e.get(t, o.ID)
		assert.Equal(t, 50.0, d.ChargedTotal)
		assert.Equal(t, 20.0, d.PaidAmount)
		assert.Equal(t, 30.0, d.Balance)
		assert.False(t, d.PaymentReceived)
		assert.Equal(t, 0.0, d.OverpaidAmount)
	})

	t.Run("REQPAY002S02_REQPAY004S01_split_modes", func(t *testing.T) {
		o := newOrder50(t)
		e.pay(t, o.ID, "cash", 20)
		e.pay(t, o.ID, "venmo", 30)
		d := e.get(t, o.ID)
		assert.Equal(t, 50.0, d.PaidAmount)
		assert.Equal(t, 0.0, d.Balance)
		assert.True(t, d.PaymentReceived)

		ps, err := e.svc.ListPayments(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		require.Len(t, ps, 2)
		assert.Equal(t, "cash", ps[0].Mode)
		assert.Equal(t, "venmo", ps[1].Mode)
		assert.Equal(t, 30.0, ps[1].Amount)
		assert.NotZero(t, ps[0].CreatedAt)
	})

	t.Run("REQPAY003S01_cash_overpay", func(t *testing.T) {
		o := e.createOrder(t)
		_, err := e.svc.AddLine(ctx, e.tenantID, o.ID, &models.OrderLineCreateRequest{
			MenuItemID: e.riceItemID, Quantity: 1, UnitPriceOverride: ptr(47.0),
		})
		require.NoError(t, err)
		e.pay(t, o.ID, "cash", 50)
		d := e.get(t, o.ID)
		assert.True(t, d.PaymentReceived)
		assert.Equal(t, 3.0, d.OverpaidAmount)
		assert.Equal(t, -3.0, d.Balance)
	})

	t.Run("REQPAY003S02_S03_override_drop_and_price_rise", func(t *testing.T) {
		o := newOrder50(t)
		e.pay(t, o.ID, "zelle", 50)

		d, err := e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: ptr(40.0)},
		})
		require.NoError(t, err)
		assert.Equal(t, 10.0, d.OverpaidAmount)
		assert.True(t, d.PaymentReceived)

		d, err = e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: ptr(55.0)},
		})
		require.NoError(t, err)
		assert.Equal(t, 5.0, d.Balance)
		assert.False(t, d.PaymentReceived)
		assert.Equal(t, 0.0, d.OverpaidAmount)
	})

	t.Run("REQPAY001S03_payment_after_pickedup", func(t *testing.T) {
		o := newOrder50(t)
		d := e.pickup(t, o.ID)
		assert.False(t, d.PaymentReceived)

		e.pay(t, o.ID, "paypal", 50)
		d = e.get(t, o.ID)
		assert.True(t, d.PaymentReceived)
		assert.Equal(t, 0.0, d.Balance)
	})

	t.Run("REQPAY001_credit_mode_accepted_case_insensitive", func(t *testing.T) {
		o := newOrder50(t)
		e.pay(t, o.ID, "CREDIT", 1)
		ps, err := e.svc.ListPayments(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		require.Len(t, ps, 1)
		assert.Equal(t, "credit", ps[0].Mode)
	})
}

// REQLIFE001–REQLIFE005
func TestLifecycle(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	e := newTenant(t, db, "tenant-a")

	t.Run("REQLIFE001S01_REQLIFE002S01_REQLIFE004_happy_path", func(t *testing.T) {
		o := e.createOrder(t)
		e.addRice(t, o.ID, 2)

		d, err := e.svc.Accept(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		assert.Equal(t, models.OrderStatusAccepted, d.Status)

		d, err = e.svc.StartPreparing(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		assert.Equal(t, models.OrderStatusInProgress, d.Status)

		d, err = e.svc.Ready(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		assert.Equal(t, models.OrderStatusReady, d.Status)

		d, err = e.svc.Pickup(ctx, e.tenantID, o.ID) // REQLIFE004S03
		require.NoError(t, err)
		assert.Equal(t, models.OrderStatusPickedUp, d.Status)
		require.NotNil(t, d.PickedupAt)
		require.NotNil(t, d.FrozenTotal)
		assert.Equal(t, 10.0, *d.FrozenTotal)
		assert.Equal(t, models.OrderStatusPickedUp, e.get(t, o.ID).Status)
	})

	t.Run("REQLIFE001S02_reject_skip_received_to_ready", func(t *testing.T) {
		o := e.createOrder(t)
		_, err := e.svc.Ready(ctx, e.tenantID, o.ID)
		requireStatus(t, err, 400)
		_, err = e.svc.StartPreparing(ctx, e.tenantID, o.ID)
		requireStatus(t, err, 400)
		_, err = e.svc.Pickup(ctx, e.tenantID, o.ID)
		requireStatus(t, err, 400)
		assert.Equal(t, models.OrderStatusReceived, e.get(t, o.ID).Status)
	})

	t.Run("REQLIFE001_reject_illegal_edges", func(t *testing.T) {
		o := e.createOrder(t)
		_, err := e.svc.Accept(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		_, err = e.svc.Accept(ctx, e.tenantID, o.ID) // REQLIFE002: only from RECEIVED
		requireStatus(t, err, 400)
		_, err = e.svc.Refuse(ctx, e.tenantID, o.ID, "") // REQLIFE003: only from RECEIVED
		requireStatus(t, err, 400)
		_, err = e.svc.Ready(ctx, e.tenantID, o.ID)
		requireStatus(t, err, 400)
		assert.Equal(t, models.OrderStatusAccepted, e.get(t, o.ID).Status)

		_, err = e.svc.StartPreparing(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		_, err = e.svc.Ready(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		_, err = e.svc.Pickup(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		for _, fn := range []func() (*models.OrderDetail, error){
			func() (*models.OrderDetail, error) { return e.svc.Accept(ctx, e.tenantID, o.ID) },
			func() (*models.OrderDetail, error) { return e.svc.Pickup(ctx, e.tenantID, o.ID) },
		} {
			_, err := fn()
			requireStatus(t, err, 400)
		}
	})

	t.Run("REQLIFE001S03_reject_transitions_out_of_declined", func(t *testing.T) {
		o := e.createOrder(t)
		_, err := e.svc.Refuse(ctx, e.tenantID, o.ID, "")
		require.NoError(t, err)
		for _, fn := range []func() (*models.OrderDetail, error){
			func() (*models.OrderDetail, error) { return e.svc.Accept(ctx, e.tenantID, o.ID) },
			func() (*models.OrderDetail, error) { return e.svc.StartPreparing(ctx, e.tenantID, o.ID) },
			func() (*models.OrderDetail, error) { return e.svc.Ready(ctx, e.tenantID, o.ID) },
			func() (*models.OrderDetail, error) { return e.svc.Pickup(ctx, e.tenantID, o.ID) },
			func() (*models.OrderDetail, error) { return e.svc.Refuse(ctx, e.tenantID, o.ID, "again") },
		} {
			_, err := fn()
			requireStatus(t, err, 400)
		}
		assert.Equal(t, models.OrderStatusDeclined, e.get(t, o.ID).Status)
	})

	t.Run("REQLIFE003S01_refuse_default_reason", func(t *testing.T) {
		o := e.createOrder(t)
		for _, reason := range []string{""} {
			d, err := e.svc.Refuse(ctx, e.tenantID, o.ID, reason)
			require.NoError(t, err)
			assert.Equal(t, models.OrderStatusDeclined, d.Status)
			require.NotNil(t, d.RefuseReason)
			assert.Equal(t, "No available slots", *d.RefuseReason)
		}
		got := e.get(t, o.ID)
		require.NotNil(t, got.RefuseReason)
		assert.Equal(t, models.DefaultRefuseReason, *got.RefuseReason)

		blank := e.createOrder(t)
		d, err := e.svc.Refuse(ctx, e.tenantID, blank.ID, "   ")
		require.NoError(t, err)
		assert.Equal(t, models.DefaultRefuseReason, *d.RefuseReason)
	})

	t.Run("REQLIFE003S02_refuse_custom_reason", func(t *testing.T) {
		o := e.createOrder(t)
		d, err := e.svc.Refuse(ctx, e.tenantID, o.ID, "  Catering queue full ")
		require.NoError(t, err)
		assert.Equal(t, models.OrderStatusDeclined, d.Status)
		require.NotNil(t, d.RefuseReason)
		assert.Equal(t, "Catering queue full", *d.RefuseReason)
		assert.Equal(t, "Catering queue full", *e.get(t, o.ID).RefuseReason)
	})

	t.Run("REQLIFE_unknown_order_404", func(t *testing.T) {
		_, err := e.svc.Accept(ctx, e.tenantID, uuid.New().String())
		requireStatus(t, err, 404)
		_, err = e.svc.Refuse(ctx, e.tenantID, uuid.New().String(), "")
		requireStatus(t, err, 404)
	})

	t.Run("REQLIFE005S01_patch_status_rejected", func(t *testing.T) {
		o := e.createOrder(t)
		for _, st := range []string{"ACCEPTED", "RECEIVED", "COMPLETE", "PICKEDUP", "BOGUS"} {
			_, err := e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{Status: ptr(st)})
			requireStatus(t, err, 400)
		}
		assert.Equal(t, models.OrderStatusReceived, e.get(t, o.ID).Status)
	})

	t.Run("REQOLINE001S05_lines_editable_after_accepted_through_ready", func(t *testing.T) {
		o := e.createOrder(t)
		_, err := e.svc.Accept(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		line := e.addRice(t, o.ID, 1)
		_, err = e.svc.StartPreparing(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		_, err = e.svc.UpdateLine(ctx, e.tenantID, o.ID, line.ID, &models.OrderLineUpdateRequest{Quantity: ptr(3)})
		require.NoError(t, err)
		_, err = e.svc.Ready(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		require.NoError(t, e.svc.RemoveLine(ctx, e.tenantID, o.ID, line.ID))
	})

	t.Run("REQORDER003S02_total_override_on_accepted", func(t *testing.T) {
		o := e.createOrder(t)
		e.addRice(t, o.ID, 2)
		_, err := e.svc.Accept(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		d, err := e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: ptr(7.0)},
		})
		require.NoError(t, err)
		assert.Equal(t, 7.0, d.ChargedTotal)
	})

	t.Run("REQORDER003S04_REQOLINE_REQPAY001S04_declined_is_read_only", func(t *testing.T) {
		o := e.createOrder(t)
		line := e.addRice(t, o.ID, 1)
		_, err := e.svc.Refuse(ctx, e.tenantID, o.ID, "")
		require.NoError(t, err)

		// REQORDER003S04
		_, err = e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: ptr(1.0)},
		})
		requireStatus(t, err, 400)
		_, err = e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{
			TotalOverride: models.NullableFloat64{Set: true, Value: nil},
		})
		requireStatus(t, err, 400)
		_, err = e.svc.Update(ctx, e.tenantID, o.ID, &models.OrderUpdateRequest{CustomerName: ptr("Bina")})
		requireStatus(t, err, 400)

		// REQOLINE001S06, REQOLINE004S03
		_, err = e.svc.AddLine(ctx, e.tenantID, o.ID, &models.OrderLineCreateRequest{MenuItemID: e.riceItemID, Quantity: 1})
		requireStatus(t, err, 400)
		_, err = e.svc.UpdateLine(ctx, e.tenantID, o.ID, line.ID, &models.OrderLineUpdateRequest{Quantity: ptr(2)})
		requireStatus(t, err, 400)
		requireStatus(t, e.svc.RemoveLine(ctx, e.tenantID, o.ID, line.ID), 400)

		// REQPAY001S04
		_, err = e.svc.RecordPayment(ctx, e.tenantID, o.ID, &models.PaymentCreateRequest{Mode: "cash", Amount: 5})
		requireStatus(t, err, 400)
		ps, err := e.svc.ListPayments(ctx, e.tenantID, o.ID)
		require.NoError(t, err)
		assert.Empty(t, ps)

		// get/list still work
		d := e.get(t, o.ID)
		assert.Equal(t, models.OrderStatusDeclined, d.Status)
		assert.Len(t, d.Lines, 1)
		_, err = e.svc.List(ctx, e.tenantID)
		require.NoError(t, err)
	})
}
