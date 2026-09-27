package fixtures

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TenantFixture represents a test tenant for multi-tenant testing
type TenantFixture struct {
	ID        string
	Name      string
	CreatedAt int64
	Cleanup   func() error
}

// CreateTestTenant creates a tenant row and returns a fixture for cleanup.
func CreateTestTenant(ctx context.Context, db *sqlx.DB) (*TenantFixture, error) {
	id := uuid.New().String()
	name := "test-tenant-" + id[:8]
	now := time.Now().UTC().UnixMilli()

	_, err := db.ExecContext(ctx, `
		INSERT INTO tenant (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, TRUE, $4, $5)
	`, id, name, "fixture tenant", now, now)
	if err != nil {
		return nil, err
	}

	cleanup := func() error {
		_, err := db.ExecContext(ctx, `DELETE FROM food_item WHERE tenant_id = $1`, id)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, `DELETE FROM food_category WHERE tenant_id = $1`, id)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, `DELETE FROM tenant WHERE id = $1`, id)
		return err
	}

	return &TenantFixture{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		Cleanup:   cleanup,
	}, nil
}
