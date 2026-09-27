package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Registers the "postgres" driver
	"github.com/pareshvernekar/homecooked/tests/fixtures"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// DatabaseHelper provides common database operations for testing
type DatabaseHelper struct {
	DB *sqlx.DB
}

// NewDatabaseHelper creates a new database helper with proper connection management
func NewDatabaseHelper(ctx context.Context) (*DatabaseHelper, error) {

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}
	// Pass sslmode=disable directly to ConnectionString
	connURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}
	db, err := sqlx.Connect("postgres", connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &DatabaseHelper{DB: db}, nil
}

// Terminate terminates the database connection and container
func (h *DatabaseHelper) Terminate(ctx context.Context) error {
	return h.DB.Close()
}

// CreatePostgresConnection creates a new PostgreSQL connection for testing
func CreatePostgresConnection(ctx context.Context) (*sqlx.DB, func() error, error) {
	db, err := NewDatabaseHelper(ctx)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		err := db.Terminate(ctx)
		return err
	}

	return db.DB, cleanup, nil
}

// InitializeSchema initializes the test database schema.
// Keep in sync with scripts/init/01_food_catalog.sql.
func InitializeSchema(ctx context.Context, db *sqlx.DB) error {
	schemaSQL := `
	CREATE TABLE IF NOT EXISTS tenant (
	    id VARCHAR(50) PRIMARY KEY,
	    name VARCHAR(100) NOT NULL,
	    description TEXT,
	    is_active BOOLEAN NOT NULL DEFAULT TRUE,
	    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
	    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT
	);

	CREATE TABLE IF NOT EXISTS food_category (
	    id VARCHAR(50) NOT NULL,
	    tenant_id VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
	    name VARCHAR(100) NOT NULL,
	    description TEXT,
	    is_active BOOLEAN NOT NULL DEFAULT TRUE,
	    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
	    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
	    PRIMARY KEY (tenant_id, id),
	    CONSTRAINT food_category_tenant_name_unique UNIQUE (tenant_id, name)
	);

	CREATE TABLE IF NOT EXISTS food_item (
	    id VARCHAR(50) NOT NULL,
	    tenant_id VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
	    name VARCHAR(100) NOT NULL,
	    description TEXT,
	    price NUMERIC(10,2) NOT NULL CHECK (price >= 0),
	    availability_status VARCHAR(50) NOT NULL CHECK (availability_status IN ('available', 'low_stock', 'unavailable')),
	    category_id VARCHAR(50),
	    image_url TEXT,
	    avoidance TEXT,
	    is_vegetarian BOOLEAN NOT NULL DEFAULT FALSE,
	    is_active BOOLEAN NOT NULL DEFAULT TRUE,
	    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
	    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
	    delivered_at BIGINT,
	    PRIMARY KEY (tenant_id, id),
	    CONSTRAINT food_item_category_fk
	        FOREIGN KEY (tenant_id, category_id)
	        REFERENCES food_category (tenant_id, id)
	        ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_food_category_tenant ON food_category(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_food_item_tenant ON food_item(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_food_item_category ON food_item(tenant_id, category_id);
	`
	if err := SetUserContext(ctx, db, "test-tenant"); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return err
	}

	return nil
}

// SetUserContext sets the current_user_id for the current database session
func SetUserContext(ctx context.Context, db *sqlx.DB, tenantID string) error {
	// Method 1: Using SET command
	query := fmt.Sprintf("SET app.current_tenant_id = '%s'", tenantID)
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to set user context: %w", err)
	}
	return nil
}

// CreateTestTenant ensures tenant "test-tenant" exists and seeds a "vegetarian" category.
func CreateTestTenant(ctx context.Context, db *sqlx.DB) (*fixtures.TenantFixture, error) {
	tenantID := "test-tenant"
	categoryID := uuid.New().String()
	now := time.Now().UTC().UnixMilli()

	_, err := db.ExecContext(ctx, `
		INSERT INTO tenant (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, TRUE, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`, tenantID, "Test Tenant", "Integration test tenant", now, now)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO food_category (id, name, tenant_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, TRUE, $4, $5)
		ON CONFLICT (tenant_id, name) DO UPDATE SET updated_at = EXCLUDED.updated_at
	`, categoryID, "vegetarian", tenantID, now, now)
	if err != nil {
		return nil, err
	}

	// Prefer the persisted category id when a prior vegetarian row already existed.
	var persistedID string
	if err := db.GetContext(ctx, &persistedID, `
		SELECT id FROM food_category WHERE tenant_id = $1 AND name = $2
	`, tenantID, "vegetarian"); err != nil {
		return nil, err
	}
	categoryID = persistedID

	cleanup := func() error {
		_, err := db.ExecContext(ctx, `DELETE FROM food_item WHERE tenant_id = $1`, tenantID)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, `DELETE FROM food_category WHERE tenant_id = $1`, tenantID)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, `DELETE FROM tenant WHERE id = $1`, tenantID)
		return err
	}

	return &fixtures.TenantFixture{
		ID:        categoryID,
		Name:      "vegetarian",
		CreatedAt: now,
		Cleanup:   cleanup,
	}, nil
}
