package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pareshvernekar/homecooked/internal/config"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
)

var DB *sqlx.DB
var TenantID string = "1" // Default tenant ID

// Connection pool configuration with defaults
const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 5 * time.Minute
)

func InitDB(tenantID string, l *logger.Logger) error {
	// Retrieve database configuration from Viper instead of os.Getenv
	host := config.Config.GetString("database.host")
	port := config.Config.GetInt("database.port")
	user := config.Config.GetString("database.user")
	password := config.Config.GetString("database.password")
	dbname := config.Config.GetString("database.name")
	var dbURL string

	// If all empty, try connection string format
	if host == "" || port == 0 || user == "" {
		connectionString := os.Getenv("DB_CONNECTION_STRING")
		if connectionString != "" {
			dbURL = connectionString
		} else {
			return fmt.Errorf("database not configured")
		}
	} else {
		dbURL = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	}

	fmt.Println("Connecting to database with URL:", dbURL)
	var err error

	DB, err = sqlx.Connect("postgres", dbURL)
	if err != nil {
		l.Error(context.Background(), "Failed to connect to database", slog.Any("error", err))
		return err
	}

	// Configure connection pool parameters
	DB.SetMaxOpenConns(defaultMaxOpenConns)
	DB.SetMaxIdleConns(defaultMaxIdleConns)
	DB.SetConnMaxLifetime(defaultConnMaxLifetime)

	// Execute query to set tenant context in session - ensures all queries on this connection use correct RLS policies
	if err := SetTenantContext(tenantID, l); err != nil {
		l.Error(context.Background(), "Failed to set tenant ID", slog.Any("error", err))
		return err
	}

	// Ping to ensure connection is healthy and session is set correctly
	if err := DB.Ping(); err != nil {
		l.Error(context.Background(), "Database connection failed", slog.Any("error", err))
		return err
	}

	l.Info(context.Background(), "Database connected successfully with tenant isolation", "tenant_id", tenantID)

	// Register signal handler for graceful shutdown
	// 1. Create the channel
	errChan := make(chan error)

	go func(ch chan error) {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		<-sigChan
		err := DB.Close()
		if err != nil {
			l.Error(context.Background(), "Failed to close DB", slog.Any("error", err))
			ch <- err
		} else {
			l.Info(context.Background(), "Received shutdown signal, closing database connection")
		}

	}(errChan)
	if err := <-errChan; err != nil {
		return fmt.Errorf("error during shutdown: %w", err)
	}
	return nil
}

func SetTenantContext(tenantID string, l *logger.Logger) error {
	if DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := DB.NamedExecContext(ctx, "SET app.current_tenant_id = :tenant_id", map[string]interface{}{"tenant_id": tenantID}); err != nil {
			l.Error(context.Background(), "Failed to set tenant ID session variable", slog.Any("error", err))
			return err
		}
	}
	return nil
}

// Close closes the database connection pool
func Close(l *logger.Logger) error {
	if DB != nil {
		defer func() {
			if closeErr := DB.Close(); closeErr != nil {
				l.Error(context.Background(), "Failed to close DB", slog.Any("error", closeErr))
			} else {
				l.Info(context.Background(), "Database connection pool closed")
			}
		}()
		err := DB.Close()
		if err != nil {
			l.Error(context.Background(), "Failed to close DB", slog.Any("error", err))
			return err
		}
	}
	return nil
}

// GetDB returns the current database connection (with tenant context already set)
func GetDB(l *logger.Logger) *sqlx.DB {
	if DB == nil {
		l.Error(context.Background(), "Database connection not initialized")
		return nil
	}
	return DB
}

// SetTenantID updates the global TenantID for future queries
// WARNING: This affects all subsequent queries using this connection pool
func SetTenantID(tenantID string, l *logger.Logger) error {
	TenantID = tenantID
	if err := SetTenantContext(tenantID, l); err != nil {
		return err
	}
	l.Info(context.Background(), "Tenant ID updated", "tenant_id", tenantID)
	return nil
}

// UseConnection executes a query with the current tenant context
func UseConnection(l *logger.Logger, fn func(ctx context.Context, db *sqlx.DB) error) error {
	db := GetDB(l)
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Query with implicit tenant isolation from session variable
	return fn(ctx, db)
}
