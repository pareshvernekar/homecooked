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

// CreateTestTenant creates a new isolated tenant with given name for testing purposes
func CreateTestTenant(ctx context.Context, db *sqlx.DB) (*TenantFixture, error) {
	id := uuid.New()
	name := "test-tenant-" + id.String()[:4]
	now := time.Now().UTC().UnixMilli()

	query := `
		INSERT INTO food_category (id, name, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	_, err := db.ExecContext(ctx, query, id.String(), name, id.String(), now, now)
	if err != nil {
		return nil, err
	}

	// Create cleanup function to remove tenant data (all categories for this tenant)
	cleanup := func() error {
		deleteQuery := `DELETE FROM food_category WHERE id = $1 AND tenant_id = $2`
		_, err := db.ExecContext(ctx, deleteQuery, id.String(), id.String())
		return err
	}

	return &TenantFixture{
		ID:        id.String(),
		Name:      name,
		CreatedAt: now,
		Cleanup:   cleanup,
	}, nil
}
