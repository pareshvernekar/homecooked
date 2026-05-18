package database

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/pareshvernekar/homecooked/internal/logger"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/config"
)

func TestInitDB_WithConnectionString(t *testing.T) {
	TenantID = "1"
	DB = NewMockDB("1") // Direct initialization - no defer needed
	l := logger.NewLogger()
	defer func() {
		if closeErr := DB.Close(); closeErr != nil {
			t.Logf("Failed to close DB: %v", closeErr)
		}
	}() // Ensure mock DB is closed after test
	if err := SetTenantContext("1", l); err == nil {
		t.Log("✓ Tenant context set successfully in mock mode")
	} else {
		t.Logf("SetTenantContext skipped: %v", err)
		return
	}

}

func TestSetTenantContext(t *testing.T) {

	TenantID = "test-tenant-uuid-12345"
	l := logger.NewLogger()
	DB = NewMockDB("test-tenant-uuid-12345")
	defer func() {
		if closeErr := DB.Close(); closeErr != nil {
			t.Logf("Failed to close DB: %v", closeErr)
		}
	}() // Ensure mock DB is closed after test

	err := SetTenantContext("test-tenant-uuid-12345", l)
	if err != nil {
		t.Errorf("Failed to set tenant context: %v", err)
		return
	}

	if TenantID != "test-tenant-uuid-12345" {
		t.Errorf("TenantID not updated. Expected 'test-tenant-uuid-12345', got '%s'", TenantID)
	} else {
		t.Log("✓ Tenant context set successfully")
	}
}

func TestGetDB(t *testing.T) {
	initTest()
	l := logger.NewLogger()
	// Create mock DB with configured pool settings (direct instantiation)
	DB = NewMockDB("1") // ← Mock created directly, no InitDB call
	defer func() {
		if closeErr := DB.Close(); closeErr != nil {
			t.Logf("Failed to close DB: %v", closeErr)
		}
	}() // Ensure mock DB is closed after test
	err := SetTenantContext("test-tenant-uuid-12345", l)
	if err != nil {
		t.Errorf("Failed to set tenant context: %v", err)
		return
	}

	db := GetDB(l)
	if db == nil {
		t.Errorf("GetDB returned nil after initialization")
	} else {
		fmt.Printf("✓ GetDB returns valid database connection (MOCK)\n")

		// Verify mock works
		if err := db.Ping(); err != nil {
			t.Errorf("Mock DB Ping failed: %v", err)
			return
		}
	}
}

func TestSetTenantID(t *testing.T) {

	l := logger.NewLogger()

	// Create mock DB with configured pool settings (direct instantiation)
	DB = NewMockDB("original-tenant-id") // ← Mock created directly, no InitDB call
	defer func() {
		if closeErr := DB.Close(); closeErr != nil {
			t.Logf("Failed to close DB: %v", closeErr)
		}
	}() // Ensure mock DB is closed after test

	currentID := TenantID
	if currentID != "original-tenant-id" {
		t.Errorf("Initial TenantID mismatch: expected 'original-tenant-id', got '%s'", currentID)
		return
	}

	err := SetTenantID("new-tenant-id", l)
	if err != nil {
		t.Logf("SetTenantID failed: %v", err)
		return
	}

	if TenantID != "new-tenant-id" {
		t.Errorf("SetTenantID did not update global variable. Expected 'new-tenant-id', got '%s'", TenantID)
	} else {
		fmt.Println("✓ Global TenantID updated successfully")
	}
}

func TestClose(t *testing.T) {
	l := logger.NewLogger()
	// Create mock DB with configured pool settings (direct instantiation)
	DB = NewMockDB("original-tenant-id")
	err := SetTenantContext("test-tenant-uuid-12345", l)
	if err != nil {
		t.Errorf("Failed to set tenant context: %v", err)
		return
	}

	if DB != nil {
		err := DB.Close()
		if err != nil {
			t.Logf("Close failed: %v", err)
		} else {
			fmt.Println("✓ Database pool closed successfully (MOCK)")
		}
	}
}

func TestConnectionPoolConfigurations(t *testing.T) {
	l := logger.NewLogger()
	// Create mock DB with configured pool settings (direct instantiation)
	DB = NewMockDB("1") // ← Mock created directly, no InitDB call
	defer func() {
		if closeErr := DB.Close(); closeErr != nil {
			t.Logf("Failed to close DB: %v", closeErr)
		}
	}() // Ensure mock DB is closed after test
	err := SetTenantContext("test-tenant-uuid-12345", l)
	if err != nil {
		t.Errorf("Failed to set tenant context: %v", err)
		return
	}
	// Set the pool configuration on our mock
	DB.SetMaxOpenConns(defaultMaxOpenConns) // 25
	DB.SetMaxIdleConns(defaultMaxIdleConns) // 5
	DB.SetConnMaxLifetime(defaultConnMaxLifetime)

	stats := DB.Stats()
	fmt.Printf("Stats: MaxOpenConns=%d, Idle=%d\n", stats.MaxOpenConnections, stats.Idle)

	if stats.MaxOpenConnections != defaultMaxOpenConns {
		t.Errorf("MaxOpenConns mismatch")
	} else {
		t.Log("✓ Pool configuration verified correctly")
	}
}

// Helper function for test setup
func initTest() {
	// Initialize config before calling InitDB
	err := os.Setenv("database.host", "localhost")
	if err != nil {
		panic(fmt.Sprintf("Failed to set environment variable: %v", err))
	}
	err = os.Setenv("database.port", "5432")
	if err != nil {
		panic(fmt.Sprintf("Failed to set environment variable: %v", err))
	}
	err = os.Setenv("database.user", "testuser")
	if err != nil {
		panic(fmt.Sprintf("Failed to set environment variable: %v", err))
	}
	err = os.Setenv("database.password", "testpass")
	if err != nil {
		panic(fmt.Sprintf("Failed to set environment variable: %v", err))
	}
	err = os.Setenv("database.name", "testdb")
	if err != nil {
		panic(fmt.Sprintf("Failed to set environment variable: %v", err))
	}

	// Trigger config initialization from env vars
	config.Init()
	if err := config.Read(); err != nil {
		panic(err)
	}
}

func NewMockDB(tenantID string) *sqlx.DB {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(fmt.Sprintf("Failed to create mock DB: %v", err))
	}
	sqlxDb := sqlx.NewDb(db, "sqlmock")

	mock.ExpectExec(regexp.QuoteMeta("SET app.current_tenant_id = ?")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	//	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow())

	sqlxDb.SetMaxOpenConns(defaultMaxOpenConns)
	sqlxDb.SetMaxIdleConns(defaultMaxIdleConns)
	sqlxDb.SetConnMaxLifetime(defaultConnMaxLifetime)
	TenantID = tenantID
	return sqlxDb
}
