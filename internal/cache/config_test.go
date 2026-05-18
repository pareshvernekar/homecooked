package cache

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// MockCacheClient is a mock implementation of the CacheClient interface for testing
type MockCacheClient struct {
	GetFunc   func(ctx context.Context, key string) (any, error)
	SetFunc   func(ctx context.Context, key string, value any, options ...interface{}) error
	HasFunc   func(key string) bool
	CallCount int
}

// NewMockCacheClient creates a mock with default "always miss" behavior
func NewMockCacheClient() *MockCacheClient {
	return &MockCacheClient{
		GetFunc: func(ctx context.Context, key string) (any, error) {
			return nil, nil // Default: always miss
		},
		SetFunc: func(ctx context.Context, key string, value any, options ...interface{}) error {
			return nil // Always succeeds
		},
	}
}

// Get executes the mock Get function with provided arguments
func (m *MockCacheClient) Get(ctx context.Context, key string) (any, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	m.CallCount++
	return nil, nil
}

// Set executes the mock Set function with provided arguments
func (m *MockCacheClient) Set(ctx context.Context, key string, value any, options ...interface{}) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, options...)
	}
	m.CallCount++
	return nil
}

// Has executes the mock Has function with provided arguments
func (m *MockCacheClient) Has(key string) bool {
	if m.HasFunc != nil {
		return m.HasFunc(key)
	}
	return false
}

// TestLoadConfig_WithDefaults verifies default configuration values when Viper is nil
func TestLoadConfig_WithDefaults(t *testing.T) {
	// Arrange - no Viper instance, should use defaults
	cfg := LoadConfig(nil)

	// Assert defaults match specification requirements
	assert.Equal(t, 30*time.Minute, cfg.DefaultTTL, "Default TTL should be 30 minutes")
	assert.Equal(t, 1000, cfg.MaxItems, "Max items should be 1000 for eviction trigger")

	// Verify per-entity TTL overrides are set correctly
	assert.Equal(t, time.Duration(30*time.Minute), cfg.TTLOverrides["weekly_menu.food_details"], "Weekly menu TTL should be 30m")
	assert.Equal(t, time.Duration(15*time.Minute), cfg.TTLOverrides["order.menu_items"], "Order items TTL should be 15m")
	assert.Equal(t, time.Duration(60*time.Minute), cfg.TTLOverrides["catering_menu.items"], "Catering menu TTL should be 60m")
	assert.Equal(t, time.Duration(24*time.Hour), cfg.TTLOverrides["food_catalog.categories"], "Categories TTL should be 24h")
}

// TestLoadConfig_WithViperDefaults verifies Viper config with default values
func TestLoadConfig_WithViperDefaults(t *testing.T) {
	// Arrange - create viper instance with basic cache config
	v := viper.New()
	v.Set("cache.default_ttl", "45m")
	v.Set("cache.max_items", 800)

	// Act - load with viper instance
	cfg := LoadConfig(v)

	// Assert - Viper values are applied
	assert.Equal(t, 45*time.Minute, cfg.DefaultTTL, "Viper default_ttl should override default")
	assert.Equal(t, 800, cfg.MaxItems, "Viper max_items should override default")

	// TTLOverrides should still use sensible defaults for unspecified entities
	assert.Equal(t, time.Duration(30*time.Minute), cfg.TTLOverrides["weekly_menu.food_details"])
}

// TestLoadConfig_WithViperFullConfig verifies Viper config with complete TTL overrides
func TestLoadConfig_WithViperFullConfig(t *testing.T) {
	// Arrange - create viper instance with full cache configuration
	v := viper.New()
	v.Set("cache.default_ttl", "30m")
	v.Set("cache.max_items", 1000)
	ttlOverridesRaw := `
weekly_menu.food_details: 20m
order.menu_items: 10m
catering_menu.items: 90m
food_catalog.categories: 12h
catering_menu: 45m
`
	v.Set("cache.ttl_overrides", ttlOverridesRaw)

	// Act - load with full viper config
	cfg := LoadConfig(v)

	// Assert - all Viper values are applied correctly
	assert.Equal(t, 30*time.Minute, cfg.DefaultTTL)
	assert.Equal(t, 1000, cfg.MaxItems)
	assert.Equal(t, time.Duration(20*time.Minute), cfg.TTLOverrides["weekly_menu.food_details"])
	assert.Equal(t, time.Duration(10*time.Minute), cfg.TTLOverrides["order.menu_items"])
	assert.Equal(t, time.Duration(90*time.Minute), cfg.TTLOverrides["catering_menu.items"])
	assert.Equal(t, time.Duration(12*time.Hour), cfg.TTLOverrides["food_catalog.categories"])
}
