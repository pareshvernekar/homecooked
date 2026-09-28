package sizeunit_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/services/sizeunit"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
)

// REQSIZE001–REQSIZE003
func TestSizeUnitService_StandardsCustomsIsolation(t *testing.T) {
	ctx := context.Background()
	helper, err := testdb.NewDatabaseHelper(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = helper.Terminate(ctx) })
	require.NoError(t, testdb.InitializeSchema(ctx, helper.DB))

	tenantA := "tenant-a"
	tenantB := "tenant-b"
	now := time.Now().UTC().UnixMilli()
	for _, tid := range []string{tenantA, tenantB} {
		_, err := helper.DB.ExecContext(ctx, `
			INSERT INTO tenant (id, name, is_active, created_at, updated_at)
			VALUES ($1, $2, TRUE, $3, $3)
		`, tid, tid, now)
		require.NoError(t, err)
	}

	l := logger.NewLogger()
	svcA := sizeunit.NewService(repository.NewSizeUnitRepository(helper.DB, l, tenantA), l)
	svcB := sizeunit.NewService(repository.NewSizeUnitRepository(helper.DB, l, tenantB), l)

	t.Run("REQSIZE001_list_includes_standards", func(t *testing.T) {
		units, err := svcA.List(ctx, tenantA)
		require.NoError(t, err)
		codes := map[string]bool{}
		for _, u := range units {
			if u.IsSystem {
				codes[u.Code] = true
			}
		}
		for _, code := range []string{"serving", "tray", "piece", "dozen", "kg", "g", "liter"} {
			assert.True(t, codes[code], "missing standard %s", code)
		}
	})

	t.Run("REQSIZE002_custom_isolated_and_duplicate_rejected", func(t *testing.T) {
		unit, err := svcA.CreateCustom(ctx, tenantA, &models.SizeUnitCreateRequest{
			Code: "party-pan", DisplayName: "Party Pan",
		})
		require.NoError(t, err)
		require.NotEmpty(t, unit.ID)

		_, err = svcA.CreateCustom(ctx, tenantA, &models.SizeUnitCreateRequest{
			Code: "party-pan", DisplayName: "Party Pan 2",
		})
		require.Error(t, err)

		listB, err := svcB.List(ctx, tenantB)
		require.NoError(t, err)
		for _, u := range listB {
			assert.NotEqual(t, "party-pan", u.Code)
		}
	})

	t.Run("REQSIZE001_system_unit_not_deletable", func(t *testing.T) {
		err := svcA.DeactivateCustom(ctx, tenantA, "su_tray")
		require.Error(t, err)
	})

	t.Run("REQSIZE003_other_tenant_custom_not_usable", func(t *testing.T) {
		created, err := svcB.CreateCustom(ctx, tenantB, &models.SizeUnitCreateRequest{
			Code: "b-only", DisplayName: "B Only",
		})
		require.NoError(t, err)

		err = svcA.EnsureUsableByTenant(ctx, tenantA, created.ID)
		require.Error(t, err)

		err = svcA.EnsureUsableByTenant(ctx, tenantA, "su_serving")
		require.NoError(t, err)
	})
}
