package models

import (
	"time"
)

// FoodCategory represents a food category in the database (e.g., vegetarian, non-vegetarian, vegan)
type FoodCategory struct {
	ID           string          `db:"id"`
	TenantID     string          `db:"tenant_id"`
	Name         string          `db:"name"`
	Description  *string         `db:"description"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
}

// IsValid validates that the FoodCategory has required fields
func (f *FoodCategory) IsValid() error {
	if f.Name == "" {
		return nil // name is nullable based on spec
	}
	if len(f.Name) > 100 {
		return nil // name is VARCHAR(100)
	}
	if f.ID == "" {
		return nil // id is primary key, required
	}
	return nil
}

// GetName returns the category name in lowercase kebab-case format for cache keys
func (f *FoodCategory) GetNameForCacheKey() string {
	name := f.Name
	for _, c := range name {
		if !isLowerKebabRune(c) {
			name += "-"
		}
	}
	return name
}

// GetCacheKey returns the cache key in entity_type:name format (e.g., "food-category:vegetarian")
func (f *FoodCategory) GetCacheKey() string {
	entityType := "food-category"
	keyName := f.GetNameForCacheKey()
	if !isValidCacheKeyName(keyName) {
		keyName = "unknown" // fallback for invalid names
	}
	return entityType + ":" + keyName
}

// isLowerKebabRune checks if a character is valid in kebab-case identifiers (lowercase letters, numbers, hyphens)
func isLowerKebabRune(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-'
}

// isValidCacheKeyName validates that the name follows kebab-case naming conventions
func isValidCacheKeyName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !isLowerKebabRune(c) && c != '-' {
			return false
		}
	}
	return true
}

// ConvertToSnakeCase converts a string to snake_case format (for database column mapping)
func (f *FoodCategory) ConvertToSnakeCase() string {
	s := ""
	for i, c := range f.Name {
		if i == 0 {
			s += toLower(c)
		} else if isUpper(c) {
			s += "_" + toLower(c)
		} else {
			s += string(c)
		}
	}
	return s
}

// toLower converts a rune to lowercase
func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

// isUpper checks if a rune is an uppercase letter
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}
