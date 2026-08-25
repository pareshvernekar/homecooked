package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

// TestPostgresDB represents a PostgreSQL database instance for testing
type TestPostgresDB struct {
	DB      *sqlx.DB
	Cleanup func() error
}

// NewTestPostgresDB creates a new test database connection with schema initialization
func NewTestPostgresDB(ctx context.Context) (*TestPostgresDB, func() error, error) {
	// Create PostgreSQL container using testcontainers-go
	container, err := NewPostgresContainer(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create container: %w", err)
	}

	connURL :=
		fmt.Sprintf("%s:%s@%s:5432/%s?sslmode=disable&connection_limit=10",
			os.Getenv("TEST_DB_USER")+"_test", "secretpassword123", container.host, "test_db")

	db, err := sqlx.Connect("postgres", connURL)
	if err != nil {
		return nil, func() error { return container.Terminate(ctx) }, fmt.Errorf("failed to connect: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, func() error { return container.Terminate(ctx) }, fmt.Errorf("ping failed: %w", err)
	}

	if err := initializeSchema(ctx, db); err != nil {
		dbErr := db.Close()
		if dbErr != nil {
			return nil, func() error { return container.Terminate(ctx) }, fmt.Errorf("error closing DB: %w", dbErr)
		}
		return nil, func() error { return container.Terminate(ctx) }, fmt.Errorf("schema init failed: %w", err)
	}

	cleanup := container.stopFunc
	return &TestPostgresDB{DB: db, Cleanup: cleanup}, cleanup, nil
}

// Initialize database schema with SQL statements
func initializeSchema(ctx context.Context, db *sqlx.DB) error {
	schemaSQL := `
-- Schema for integration testing with TestContainers PostgreSQL instance
CREATE TABLE IF NOT EXISTS food_category (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS food_item (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    availability_status VARCHAR(50) NOT NULL
        CHECK (availability_status IN ('available', 'low_stock', 'unavailable')),
    category_id UUID REFERENCES food_category(id) ON DELETE CASCADE,
    image_url TEXT,
    avoidance TEXT,
    is_vegetarian BOOLEAN DEFAULT FALSE,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_category FOREIGN KEY (category_id)
        MATCHES (SELECT id FROM food_category WHERE tenant_id = current_setting('app.current_tenant_id'::text, false)::uuid),
    CONSTRAINT valid_price CHECK (price >= 0)
);

CREATE INDEX idx_food_category_tenant ON food_category(tenant_id);
CREATE INDEX idx_food_item_tenant ON food_item(tenant_id);
CREATE INDEX idx_food_item_category ON food_item(category_id);
`

	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}
