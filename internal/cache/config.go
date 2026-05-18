package cache

import (
	"time"

	"github.com/goccy/go-yaml"
	"github.com/spf13/viper"
)

// CacheConfig holds configuration options for the cache system.
// All values are sourced from Viper configuration - no file-based loading.
type CacheConfig struct {
	DefaultTTL   time.Duration            // Global default TTL (e.g., 30 minutes)
	MaxItems     int                      // Eviction trigger threshold (1000 entries)
	TTLOverrides map[string]time.Duration // Per-entity TTL configuration
}

// LoadConfig loads cache configuration exclusively from Viper.
// No YAML file reading - all config must come through the application's Viper instance.
// Returns cached config with sensible defaults applied for any missing fields.
func LoadConfig(v *viper.Viper) CacheConfig {
	// Initialize config with sensible defaults (fallback if Viper is nil or empty)
	config := CacheConfig{
		DefaultTTL: 30 * time.Minute,
		MaxItems:   1000,
		TTLOverrides: map[string]time.Duration{
			"weekly_menu.food_details": 30 * time.Minute,
			"order.menu_items":         15 * time.Minute,
			"catering_menu.items":      60 * time.Minute,
			"food_catalog.categories":  24 * time.Hour,
			"catering_menu":            60 * time.Minute,
		},
	}

	// Apply Viper configuration overrides (highest priority)
	if v == nil {
		return config // Return defaults if no Viper instance provided
	}

	if defaultTTL := v.GetString("cache.default_ttl"); defaultTTL != "" {
		duration, err := time.ParseDuration(defaultTTL)
		if err == nil {
			config.DefaultTTL = duration
		}
	}

	if maxItems := v.GetInt("cache.max_items"); maxItems > 0 {
		config.MaxItems = maxItems
	}

	if ttlOverridesRaw := v.GetString("cache.ttl_overrides"); ttlOverridesRaw != "" {
		var overrides map[string]time.Duration
		if err := yaml.Unmarshal([]byte(ttlOverridesRaw), &overrides); err == nil && len(overrides) > 0 {
			config.TTLOverrides = overrides
		}
	}

	return config
}
