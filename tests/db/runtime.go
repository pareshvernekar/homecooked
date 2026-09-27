package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

// DatabaseRuntime is the runtime configuration for test database operations
type DatabaseRuntime struct {
	Container *PostgresContainer
	URL       string
	DB        *sqlx.DB
	Cleanup   func() error
}

// NewDatabaseRuntime initializes a complete database runtime environment with testcontainers-go
func NewDatabaseRuntime(ctx context.Context) (*DatabaseRuntime, error) {
	container, err := NewPostgresContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	connURL :=
		fmt.Sprintf("%s:%s@%s:5432/test_db?sslmode=disable&connection_limit=10",
			os.Getenv("TEST_DB_USER")+"_test", "secretpassword123", container.host)

	db, err := sqlx.Connect("postgres", connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	cleanup := container.stopFunc

	return &DatabaseRuntime{
		Container: container,
		URL:       connURL,
		DB:        db,
		Cleanup:   cleanup,
	}, nil
}

// GetConnection returns the database connection and cleanup function
func GetConnection(ctx context.Context) (*sqlx.DB, func() error, error) {
	container, err := NewPostgresContainer(ctx)
	if err != nil {
		return nil, nil, err
	}

	connURL :=
		fmt.Sprintf("%s:%s@%s:5432/test_db?sslmode=disable&connection_limit=10",
			os.Getenv("TEST_DB_USER")+"_test", "secretpassword123", container.host)

	db, err := sqlx.Connect("postgres", connURL)
	if err != nil {
		return nil, func() error { return container.Terminate(ctx) }, err
	}

	cleanup := func() error {
		return container.Terminate(ctx)
	}

	return db, cleanup, nil
}
