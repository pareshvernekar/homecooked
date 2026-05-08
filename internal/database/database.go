package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pareshvernekar/homecooked/internal/config"
	"github.com/pareshvernekar/homecooked/internal/logger"
)

var DB *sqlx.DB
var TenantID string = "1" // Default tenant ID

// Connection pool configuration with defaults
const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 5 * time.Minute
)

func InitDB(tenantID string) error {
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
			dbURL = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)
		}
	}

	// Add connection pool parameters to DSN
	// Note: These apply to the entire connection pool - all connections will share these settings
	TenantID = tenantID

	var err error
	DB, err = sqlx.Connect("postgres", dbURL)
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		return err
	}

	// Configure connection pool parameters
	DB.SetMaxOpenConns(defaultMaxOpenConns)
	DB.SetMaxIdleConns(defaultMaxIdleConns)
	DB.SetConnMaxLifetime(defaultConnMaxLifetime)

	// Execute query to set tenant context in session - ensures all queries on this connection use correct RLS policies
	if err := SetTenantContext(tenantID); err != nil {
		logger.Logger.Error("Failed to set tenant ID", slog.Any("error", err))
		dbErr := DB.Close()
		if dbErr != nil {
			logger.Logger.Error("Failed to close database after tenant context error", slog.Any("error", dbErr))
		}
		return err
	}

	// Ping to ensure connection is healthy and session is set correctly
	if err := DB.Ping(); err != nil {
		logger.Logger.Error("Database connection failed", slog.Any("error", err))
		dbErr := DB.Close()
		if dbErr != nil {
			logger.Logger.Error("Failed to close database after tenant context error", slog.Any("error", dbErr))
		}
		return err
	}

	// 3. Ensure the connection is closed when the program exits
	defer func() {
		if closeErr := DB.Close(); closeErr != nil {
			logger.Logger.Error("Failed to close DB", slog.Any("error", closeErr))
		}

	}()
	logger.Logger.Info("Database connected successfully with tenant isolation", "tenant_id", tenantID)
	return nil
}

func SetTenantContext(tenantID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Println("DB IS ", DB)
	if _, err := DB.ExecContext(ctx, "SET app.current_tenant_id = $1", tenantID); err != nil {
		logger.Logger.Error("Failed to set tenant ID session variable", slog.Any("error", err))
		return err
	}
	return nil
}

// Close closes the database connection pool
func Close() error {
	if DB != nil {
		defer func() {
			if closeErr := DB.Close(); closeErr != nil {
				logger.Logger.Error("Failed to close DB", slog.Any("error", closeErr))
			}
		}()
		logger.Logger.Info("Database connection pool closed")
	}
	return nil
}

// GetDB returns the current database connection (with tenant context already set)
func GetDB() *sqlx.DB {
	if DB == nil {
		logger.Logger.Error("Database connection not initialized")
		return nil
	}
	return DB
}

// SetTenantID updates the global TenantID for future queries
// WARNING: This affects all subsequent queries using this connection pool
func SetTenantID(tenantID string) error {
	TenantID = tenantID
	if err := SetTenantContext(tenantID); err != nil {
		return err
	}
	logger.Logger.Info("Tenant ID updated", "tenant_id", tenantID)
	return nil
}

// UseConnection executes a query with the current tenant context
func UseConnection(fn func(ctx context.Context, db *sqlx.DB) error) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Query with implicit tenant isolation from session variable
	return fn(ctx, db)
}
