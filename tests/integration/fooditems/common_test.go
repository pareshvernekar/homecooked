package fooditems

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	testdb "github.com/pareshvernekar/homecooked/tests/db"
)

// CreatePostgresConnection creates a new PostgreSQL connection for testing
func CreatePostgresConnection(ctx context.Context) (*sqlx.DB, func() error, error) {
	db, err := testdb.NewDatabaseHelper(ctx)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		err := db.Terminate(ctx)
		return err
	}

	return db.DB, cleanup, nil
}

// InitializeSchema initializes the test database schema
func InitializeSchema(ctx context.Context, db *sqlx.DB) error {
	schemaSQL := `
	CREATE TABLE IF NOT EXISTS food_category (
	    id VARCHAR(50) PRIMARY KEY,
	    name VARCHAR(100) NOT NULL UNIQUE,
	    description TEXT,
	    tenant_id VARCHAR(50) NOT NULL,
	    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
	    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
	);

	CREATE TABLE IF NOT EXISTS food_item (
	    id VARCHAR(50) PRIMARY KEY,
	    name VARCHAR(100) NOT NULL,
	    description TEXT,
	    price NUMERIC(10,2) NOT NULL CHECK (price >= 0),
	    availability_status VARCHAR(50) NOT NULL CHECK (availability_status IN ('available', 'low_stock', 'unavailable')),
	    category_id VARCHAR(50) REFERENCES food_category(id) ON DELETE CASCADE,
	    image_url TEXT,
	    avoidance TEXT,
	    is_vegetarian BOOLEAN DEFAULT FALSE,
	    tenant_id VARCHAR(50) NOT NULL,
	    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
	    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
	    CONSTRAINT fk_category FOREIGN KEY (category_id) MATCHES (SELECT id FROM food_category WHERE tenant_id = current_setting('app.current_tenant_id'::text, false)::TEXT),
	    CONSTRAINT valid_price CHECK (price >= 0)
	);

	CREATE INDEX idx_food_category_tenant ON food_category(tenant_id);
	CREATE INDEX idx_food_item_tenant ON food_item(tenant_id);
	CREATE INDEX idx_food_item_category ON food_item(category_id);
	`

	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return err
	}

	return nil
}

// CreateFoodCategory creates a food category in the database
func CreateFoodCategory(ctx context.Context, db *sqlx.DB, name string) (string, error) {
	id := uuid.New().String()
	desc := "Test category description"

	query := `INSERT INTO food_category (id, tenant_id, name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	var categoryID string
	now := time.Now().UTC().UnixMilli()
	_, err := db.ExecContext(ctx, query, id, "test-tenant", name, desc, now, now)
	if err != nil {
		return "", err
	}

	return categoryID, nil
}

// InsertFoodItem inserts a food item into the database
func InsertFoodItem(ctx context.Context, db *sqlx.DB, name, description string, price float64, categoryId string, tenantId string) error {
	query := `INSERT INTO food_item (id, tenant_id, name, description, price, category_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	_, err := db.ExecContext(ctx, query, uuid.New().String(), "test-tenant", name, description, price, categoryId, time.Now().UTC().UnixMilli(), time.Now().UTC().UnixMilli())
	if err != nil {
		return err
	}

	return nil
}

// SetupFoodCategory creates a test food category for all tests
func SetupFoodCategory(ctx context.Context, db *sqlx.DB) (string, error) {
	categoryID, err := CreateFoodCategory(ctx, db, "test-veg")
	if err != nil {
		return "", err
	}

	return categoryID, nil
}

// VerifyDatabaseState verifies the database state after an operation
func VerifyDatabaseState(t *testing.T, ctx context.Context, db *sqlx.DB) error {
	count := 0
	query := `SELECT COUNT(*) FROM food_item`
	err := db.Get(&count, query)
	if err != nil {
		return err
	}

	t.Logf("Database state - Food items count: %d", count)

	return nil
}
