package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

// ConnectionHelper provides database connection utilities with lifecycle management
type ConnectionHelper struct {
	container *PostgresContainer
	db        *sqlx.DB
	host      string
	stopFunc  func() error
}

// NewConnectionHelper creates a new database connection helper
func NewConnectionHelper(ctx context.Context) (*ConnectionHelper, error) {
	// Create PostgreSQL container using testcontainers-go
	container, err := NewPostgresContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host := container.host // Use the host address for connection

	connURL :=
		fmt.Sprintf("%s:%s@%s:5432/%s?sslmode=disable&connection_limit=10",
			os.Getenv("TEST_DB_USER")+"_test", "secretpassword123", host, "test_db")

	db, err := sqlx.Connect("postgres", connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping failed: %w", pingErr)
	}

	stopFunc := container.stopFunc

	return &ConnectionHelper{
		container: container,
		db:        db,
		host:      host,
		stopFunc:  stopFunc,
	}, nil
}

// GetDB returns the underlying database connection
func (h *ConnectionHelper) GetDB() *sqlx.DB {
	return h.db
}

// Ping validates the database connection is healthy
func (h *ConnectionHelper) Ping(ctx context.Context) error {
	return h.db.PingContext(ctx)
}

// Exec executes a query and returns the results
func (h *ConnectionHelper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return h.db.Exec(query, args...)
}

// Get retrieves a single row within the transaction
func (h *ConnectionHelper) Get(dest interface{}, query string, args ...interface{}) error {
	return h.db.Get(dest, query, args...)
}

// Close closes the database connection and terminates the container
func (h *ConnectionHelper) Close() error {
	_ = h.db.Close()
	return h.stopFunc()
}
