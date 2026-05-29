package cache

import (
	"time"

	"github.com/goccy/go-yaml"
	"github.com/spf13/viper"
)

// =============================================================================
// CACHE CONFIGURATION (Go 1.21+ Slices/Maps Support)
// =============================================================================
// CacheConfig holds configuration options for the cache system.
// Uses Go 1.21+ maps for type-safe duration values (no string parsing needed).
// TTL overrides use pattern: "entity_type:id" e.g., "food_category:vegetarian" = 1h
type CacheConfig struct {
	DefaultTTL       time.Duration            // Global default TTL (e.g., 30 minutes)
	MaxItems         int                      // Eviction trigger threshold (1000 entries)
	TTLOverrides     map[string]time.Duration // Per-entity TTL config: "food_category:vegetarian" -> 1h
}

// =============================================================================
// DEFAULT TTL VALUES (Go 1.21+ slices/maps patterns)
// =============================================================================
// Default TTL configuration for various entity types - using Go 1.21+ map initialization
var defaultTTLOverrides = map[string]time.Duration{
	"weekly_menu.food_details":     30 * time.Minute,      // Daily menu updates - refresh every 30m
	"order.menu_items":             15 * time.Minute,      // Frequent order changes - refresh every 15m
	"catering_menu.items":          60 * time.Minute,      // Same-day catering - refresh every hour
	"food_catalog.categories":       24 * time.Hour,        // Rare category changes - daily refresh
	"catering_menu":                60 * time.Minute,      // General catering menu TTL
}

// =============================================================================
// CONFIG LOADING FROM VIPER (No file-based loading)
// =============================================================================
// LoadConfig loads cache configuration exclusively from Viper.
// All config must come through the application's Viper instance.
// Returns cached config with sensible defaults applied for any missing fields.
func LoadConfig(v *viper.Viper) CacheConfig {
	// Initialize config with sensible defaults (fallback if Viper is nil or empty)
	config := CacheConfig{
		DefaultTTL:    30 * time.Minute,
		MaxItems:      1000,
		TTLOverrides:  make(map[string]time.Duration), // Empty map - will populate with defaults or Viper overrides
	}

	// Copy default TTL overrides (highest priority if not overridden)
	for k, v := range defaultTTLOverrides {
		config.TTLOverrides[k] = v
	}

	// Apply Viper configuration overrides (highest priority)
	if v == nil {
		return config // Return defaults if no Viper instance provided
	}

	// Override DefaultTTL from environment: CACHE_DEFAULT_TTL=30m
	if defaultTTL := v.GetString("cache.default_ttl"); defaultTTL != "" {
		duration, err := time.ParseDuration(defaultTTL)
		if err == nil && duration > 0 {
			config.DefaultTTL = duration
		}
	}

	// Override MaxItems from environment: CACHE_MAX_ITEMS=800
	if maxItems := v.GetInt("cache.max_items"); maxItems > 0 {
		config.MaxItems = maxItems
	}

	// Override TTLOverrides from environment: CACHE_TTL_OVERRIDES='{"weekly_menu.food_details": "45m", ...}'
	if ttlOverridesRaw := v.GetString("cache.ttl_overrides"); ttlOverridesRaw != "" {
		var overrides map[string]time.Duration
		if err := yaml.Unmarshal([]byte(ttlOverridesRaw), &overrides); err == nil && len(overrides) > 0 {
			// Merge with defaults - Viper values override defaults
			for k, v := range overrides {
				config.TTLOverrides[k] = v
			}
		}
	}

	return config
}