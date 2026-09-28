package menu_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/services/menu"
	"github.com/pareshvernekar/homecooked/internal/services/menuitem"
	"github.com/pareshvernekar/homecooked/internal/services/sizeunit"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
)

// REQMENU001, REQMENU007, REQMENU009, REQITEM001, REQITEM005, REQITEM006
func TestMenuLifecycle_PublishAndImmutability(t *testing.T) {
	ctx := context.Background()
	helper, err := testdb.NewDatabaseHelper(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = helper.Terminate(ctx) })
	require.NoError(t, testdb.InitializeSchema(ctx, helper.DB))

	tenantID := "menu-tenant"
	now := time.Now().UTC().UnixMilli()
	_, err = helper.DB.ExecContext(ctx, `
		INSERT INTO tenant (id, name, is_active, created_at, updated_at) VALUES ($1,'Menu',TRUE,$2,$2)`, tenantID, now)
	require.NoError(t, err)

	catID := uuid.New().String()
	foodA := uuid.New().String()
	foodB := uuid.New().String()
	_, err = helper.DB.ExecContext(ctx, `
		INSERT INTO food_category (id, tenant_id, name, is_active, created_at, updated_at)
		VALUES ($1,$2,'mains',TRUE,$3,$3)`, catID, tenantID, now)
	require.NoError(t, err)
	for _, fid := range []struct{ id, name string }{{foodA, "Rice"}, {foodB, "Dal"}} {
		_, err = helper.DB.ExecContext(ctx, `
			INSERT INTO food_item (id, tenant_id, name, availability_status, category_id, is_active, created_at, updated_at)
			VALUES ($1,$2,$3,'available',$4,TRUE,$5,$5)`, fid.id, tenantID, fid.name, catID, now)
		require.NoError(t, err)
	}

	l := logger.NewLogger()
	menuRepo := repository.NewMenuRepository(helper.DB, l, tenantID)
	itemRepo := repository.NewMenuItemRepository(helper.DB, l, tenantID)
	sizeRepo := repository.NewSizeUnitRepository(helper.DB, l, tenantID)
	sizeSvc := sizeunit.NewService(sizeRepo, l)
	menuSvc := menu.NewService(menuRepo, itemRepo, l)
	itemSvc := menuitem.NewService(itemRepo, menuSvc, sizeSvc, l)

	start := "2026-11-02"
	end := "2026-11-08"
	created, err := menuSvc.Create(ctx, tenantID, &models.MenuCreateRequest{
		Name: "Week 1", MenuType: models.MenuTypeWeekly,
		StartDate: &start, EndDate: &end,
		Categories: []models.MenuCategoryInput{{Name: "Mains", Sequence: 1}},
	})
	require.NoError(t, err)
	require.Equal(t, models.MenuStatusDraft, created.Status)

	tree, err := menuSvc.GetTree(ctx, tenantID, created.ID)
	require.NoError(t, err)
	require.Len(t, tree.Categories, 1)
	menuCatID := tree.Categories[0].ID

	idx0 := 0
	_, err = itemSvc.AddItem(ctx, tenantID, created.ID, &models.MenuItemCreateRequest{
		CategoryID: menuCatID, Kind: models.MenuItemKindCombo, Name: "Thali",
		Components: []models.MenuItemComponentInput{
			{FoodItemID: foodA, DefaultSizeIndex: &idx0, SizeOptions: []models.MenuItemSizeOptionInput{
				{SizeUnitID: "su_serving", Qty: 1, Price: 5},
			}},
			{FoodItemID: foodB, DefaultSizeIndex: &idx0, SizeOptions: []models.MenuItemSizeOptionInput{
				{SizeUnitID: "su_serving", Qty: 1, Price: 7},
			}},
		},
	})
	require.NoError(t, err)

	require.NoError(t, menuSvc.Publish(ctx, tenantID, created.ID))

	t.Run("REQITEM006_reject_add_on_published", func(t *testing.T) {
		_, err := itemSvc.AddItem(ctx, tenantID, created.ID, &models.MenuItemCreateRequest{
			CategoryID: menuCatID, Kind: models.MenuItemKindSimple, Name: "Extra",
			Components: []models.MenuItemComponentInput{
				{FoodItemID: foodA, DefaultSizeIndex: &idx0, SizeOptions: []models.MenuItemSizeOptionInput{
					{SizeUnitID: "su_tray", Qty: 1, Price: 20},
				}},
			},
		})
		require.Error(t, err)
	})

	t.Run("REQMENU009_second_weekly_same_start_rejected", func(t *testing.T) {
		other, err := menuSvc.Create(ctx, tenantID, &models.MenuCreateRequest{
			Name: "Week 1b", MenuType: models.MenuTypeWeekly,
			StartDate: &start, EndDate: &end,
			Categories: []models.MenuCategoryInput{{Name: "Mains", Sequence: 1}},
		})
		require.NoError(t, err)
		tree2, err := menuSvc.GetTree(ctx, tenantID, other.ID)
		require.NoError(t, err)
		_, err = itemSvc.AddItem(ctx, tenantID, other.ID, &models.MenuItemCreateRequest{
			CategoryID: tree2.Categories[0].ID, Kind: models.MenuItemKindSimple, Name: "Rice",
			Components: []models.MenuItemComponentInput{
				{FoodItemID: foodA, DefaultSizeIndex: &idx0, SizeOptions: []models.MenuItemSizeOptionInput{
					{SizeUnitID: "su_serving", Qty: 1, Price: 5},
				}},
			},
		})
		require.NoError(t, err)
		err = menuSvc.Publish(ctx, tenantID, other.ID)
		require.Error(t, err)
	})

	t.Run("REQITEM005_default_total_sum", func(t *testing.T) {
		require.NoError(t, menuSvc.Unpublish(ctx, tenantID, created.ID))
		got, err := menuSvc.GetTree(ctx, tenantID, created.ID)
		require.NoError(t, err)
		require.NotEmpty(t, got.Categories[0].Items)
		item := got.Categories[0].Items[0]
		require.NotNil(t, item.DefaultTotal)
		assert.Equal(t, 12.0, *item.DefaultTotal)
		assert.NotNil(t, item.Components[0].FoodItem)
		assert.Equal(t, "Rice", item.Components[0].FoodItem.Name)
	})
}
