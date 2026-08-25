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

// InitializeSchema initializes the test database schema
func InitializeSchema(ctx context.Context, db *sqlx.DB) error {
	schemaSQL := `
	CREATE TABLE IF NOT EXISTS food_category (
	    id VARCHAR(50) PRIMARY KEY,
	    name VARCHAR(100) NOT NULL UNIQUE,
	    description TEXT,
	    tenant_id VARCHAR(50) NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
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
		is_active BOOLEAN DEFAULT TRUE,
	    tenant_id VARCHAR(50) NOT NULL,
	    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
	    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
	    CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES food_category(id),
	    CONSTRAINT valid_price CHECK (price >= 0)
	);

	CREATE INDEX idx_food_category_tenant ON food_category(tenant_id);
	CREATE INDEX idx_food_item_tenant ON food_item(tenant_id);
	CREATE INDEX idx_food_item_category ON food_item(category_id);
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

// CreateTestTenant creates a test tenant for multi-tenant testing
func CreateTestTenant(ctx context.Context, db *sqlx.DB) (*fixtures.TenantFixture, error) {
	id := uuid.New().String()
	tenantID := "test-tenant"
	query := `INSERT INTO food_category (id, name, tenant_id, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	now := time.Now().UTC().UnixMilli()

	_, err := db.ExecContext(ctx, query, id, "vegetarian", tenantID, true, now, now)
	if err != nil {
		return nil, err
	}

	cleanup := func() error {
		deleteQuery := `DELETE FROM food_category WHERE id = $1 AND tenant_id = $2`
		_, err := db.ExecContext(ctx, deleteQuery, id, id)
		return err
	}

	return &fixtures.TenantFixture{
		ID:        tenantID,
		Name:      "vegetarian",
		CreatedAt: now,
		Cleanup:   cleanup,
	}, nil
}
