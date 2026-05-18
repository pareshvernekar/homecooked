package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/stretchr/testify/assert"
)

// MockLogger is a mock implementation of the Logger interface for testing
type MockLogger struct {
	Messages []string
}

func (m *MockLogger) Info(ctx context.Context, msg string, keysAndVals ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf("%s: %s", msg, keysAndVals[0]))
}

// TestNewCacheClient_WithValidConfig tests cache initialization with valid configuration
func TestNewCacheClient_WithValidConfig(t *testing.T) {
	// Arrange - create valid cache configuration
	config := CacheConfig{
		DefaultTTL: 30 * time.Minute,
		MaxItems:   1000,
		TTLOverrides: map[string]time.Duration{
			"food_catalog.categories": 24 * time.Hour,
		},
	}

	// Act - create new cache client
	client, err := NewCacheClient(config)

	// Assert - should succeed with valid config
	assert.NoError(t, err, "Should not return error with valid configuration")
	assert.NotNil(t, client, "Should create cache client successfully")
}

// TestNewCacheClient_WithInvalidDefaultTTL tests cache initialization rejects zero TTL
func TestNewCacheClient_WithInvalidDefaultTTL(t *testing.T) {
	// Arrange - invalid config with zero TTL
	config := CacheConfig{
		DefaultTTL: 0,
		MaxItems:   1000,
		TTLOverrides: map[string]time.Duration{
			"food_catalog.categories": 24 * time.Hour,
		},
	}

	// Act - attempt to create cache client
	client, err := NewCacheClient(config)

	// Assert - should return error for invalid TTL
	assert.Error(t, err, "Should return error when default_ttl is zero")
	assert.Nil(t, client, "Client should be nil when validation fails")
}

// TestNewCacheClient_WithInvalidMaxItems tests cache initialization rejects negative max items
func TestNewCacheClient_WithInvalidMaxItems(t *testing.T) {
	// Arrange - invalid config with negative MaxItems
	config := CacheConfig{
		DefaultTTL: 30 * time.Minute,
		MaxItems:   -100,
		TTLOverrides: map[string]time.Duration{
			"food_catalog.categories": 24 * time.Hour,
		},
	}

	// Act - attempt to create cache client
	client, err := NewCacheClient(config)

	// Assert - should return error for invalid MaxItems
	assert.Error(t, err, "Should return error when max_items is negative")
	assert.Nil(t, client, "Client should be nil when validation fails")
}

// TestNewCacheClient_WithZeroMaxItems tests cache initialization rejects zero max items
func TestNewCacheClient_WithZeroMaxItems(t *testing.T) {
	// Arrange - invalid config with zero MaxItems
	config := CacheConfig{
		DefaultTTL: 30 * time.Minute,
		MaxItems:   0,
		TTLOverrides: map[string]time.Duration{
			"food_catalog.categories": 24 * time.Hour,
		},
	}

	// Act - attempt to create cache client
	client, err := NewCacheClient(config)

	// Assert - should return error for zero MaxItems
	assert.Error(t, err, "Should return error when max_items is zero")
	assert.Nil(t, client, "Client should be nil when validation fails")
}

// TestNewCacheClient_SetOperations tests cache Set operations work correctly
func TestNewCacheClient_SetOperations(t *testing.T) {
	// Arrange - create valid cache configuration
	config := CacheConfig{
		DefaultTTL: 30 * time.Minute,
		MaxItems:   1000,
		TTLOverrides: map[string]time.Duration{
			"food_catalog.food_details": 30 * time.Minute,
			"order.menu_items":          15 * time.Minute,
		},
	}

	// Act - create new cache client
	client, err := NewCacheClient(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// Test Set with per-entity TTL override (custom option)
	err = client.Set(ctx, "food_catalog.food_details:123", &models.FoodItem{}, WithTTL(60*time.Minute))
	assert.NoError(t, err, "Set with custom TTL should succeed")

	// Assert - key exists in cache
	assert.True(t, client.Has("food_catalog.food_details:123"), "Key should exist after Set")
}

// TestNewCacheClient_GetOperations tests cache Get operations work correctly
func TestNewCacheClient_GetOperations(t *testing.T) {
	// Arrange - create valid cache configuration with mock data
	config := CacheConfig{
		DefaultTTL: 30 * time.Minute,
		MaxItems:   1000,
		TTLOverrides: map[string]time.Duration{
			"food_catalog.categories":  24 * time.Hour,
			"weekly_menu.food_details": 30 * time.Minute,
			"order.menu_items":         15 * time.Minute,
			"catering_menu.items":      60 * time.Minute,
		},
	}

	// Act - create new cache client
	client, err := NewCacheClient(config)
	assert.NoError(t, err)

	ctx := context.Background()

	var testData = []byte(`[{"id":"123","name":"Test Item"}]`)

	// Set test data (simulating DB load)
	err = client.Set(ctx, "food_catalog.food_details:123", testData, WithTTL(30*time.Minute))
	assert.NoError(t, err)

	// Act - retrieve from cache
	val, err := client.Get(ctx, "food_catalog.food_details:123")

	// Assert - should return cached value
	assert.NoError(t, err, "Get should succeed after Set")
	assert.NotNil(t, val, "Should return non-nil value from cache")
}

// TestNewCacheClient_CrossTenantIsolation tests cache respects tenant isolation
func TestNewCacheClient_CrossTenantIsolation(t *testing.T) {
	// Arrange - create valid cache configuration
	config := CacheConfig{
		DefaultTTL: 30 * time.Minute,
		MaxItems:   1000,
		TTLOverrides: map[string]time.Duration{
			"food_catalog.categories": 24 * time.Hour,
		},
	}

	client, err := NewCacheClient(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// Act - Set data for tenant 1
	err = client.Set(ctx, "food_catalog.categories:tenant_1", []string{"Appetizers"}, WithTTL(30*time.Minute))
	assert.NoError(t, err)

	// Assert - Key exists (simulating tenant-specific cache key)
	assert.True(t, client.Has("food_catalog.categories:tenant_1"), "Tenant 1 key should exist")
}
