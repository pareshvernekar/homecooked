package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TransactionHelper provides transaction management utilities for test database operations
type TransactionHelper struct {
	Tx       *sqlx.Tx
	Ctx      context.Context
	Rollback func() error // Rollback function to be called on transaction cleanup
}

// NewTransactionHelper creates a new transaction helper
func NewTransactionHelper(ctx context.Context, db *sqlx.DB) (*TransactionHelper, error) {
	tx, err := db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &TransactionHelper{
		Tx:       tx,
		Ctx:      ctx,
		Rollback: func() error { return tx.Rollback() },
	}, nil
}

// Commit completes the transaction
func (th *TransactionHelper) Commit() error {
	if err := th.Tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Exec executes a statement within the transaction
func (th *TransactionHelper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return th.Tx.ExecContext(th.Ctx, query, args...)
}

// Get retrieves a single row within the transaction
func (th *TransactionHelper) Get(dest interface{}, query string, args ...interface{}) error {
	return th.Tx.GetContext(th.Ctx, dest, query, args...)
}

// Queryx queries the database and stores results in a slice within the transaction
func (th *TransactionHelper) Queryx(dest interface{}, query string, args ...interface{}) error {
	return th.Tx.SelectContext(th.Ctx, dest, query, args...)
}
