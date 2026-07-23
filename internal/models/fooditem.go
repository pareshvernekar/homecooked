package models

import (
	"strings"

	"github.com/google/uuid"
)

// FoodItem represents a food menu item in the database
type FoodItem struct {
	ID                 string  `json:"id"`
	TenantID           string  `json:"tenant_id"`
	Name               string  `json:"name"`
	Description        string  `json:"description,omitempty"`
	CategoryID         string  `json:"category_id"`
	Price              float64 `json:"price"`
	ImageURL           string  `json:"image_url,omitempty"`
	Avoidance          string  `json:"avoidance,omitempty"`
	IsVegetarian       bool    `json:"is_vegetarian"`
	AvailabilityStatus string  `json:"availability_status"`
	CreatedAt          int64   `json:"created_at,omitempty"`
	UpdatedAt          int64   `json:"updated_at,omitempty"`
	DeliveredAt        int64   `json:"delivered_at,omitempty"`
}

// FoodItemCreateRequest represents the request payload for creating a food item
type FoodItemCreateRequest struct {
	Name               string  `json:"name" binding:"required"`
	Description        string  `json:"description,omitempty"`
	CategoryName       string  `json:"category_name" binding:"required"`
	Price              float64 `json:"price" binding:"required,gte=0"`
	ImageURL           string  `json:"image_url,omitempty"`
	Avoidance          string  `json:"avoidance,omitempty"`
	IsVegetarian       *bool   `json:"is_vegetarian,omitempty"`
	AvailabilityStatus string  `json:"availability_status" binding:"required"`
}

// FoodItemUpdateRequest represents the request payload for updating a food item
type FoodItemUpdateRequest struct {
	Name               string  `json:"name,omitempty"`
	Description        string  `json:"description,omitempty"`
	CategoryName       string  `json:"category_name,omitempty"`
	Price              float64 `json:"price,omitempty"`
	ImageURL           string  `json:"image_url,omitempty"`
	Avoidance          string  `json:"avoidance,omitempty"`
	IsVegetarian       *bool   `json:"is_vegetarian,omitempty"`
	AvailabilityStatus string  `json:"availability_status,omitempty"`
}

// Helper function to validate category enum
func IsValidCategory(category string) bool {
	validCategories := map[string]bool{
		"vegetarian":     true,
		"non-vegetarian": true,
		"vegan":          true,
		"dessert":        true,
		"beverage":       true,
		"appetizer":      true,
		"main_course":    true,
		"sides":          true,
	}
	return validCategories[category]
}

// NormalizeCategory converts category name to lowercase slug format
// e.g., "Vegetarian", "VEGAN", "Non-Vegetarian" -> "vegetarian", "vegan", "non-vegetarian"
func NormalizeCategory(category string) string {
	normalized := strings.ToLower(strings.TrimSpace(category))
	// Replace spaces with hyphens for slug format
	normalized = strings.ReplaceAll(normalized, " ", "-")
	return normalized
}

// NormalizeAvailabilityStatus converts status to lowercase
// e.g., "Available", "LOW_STOCK" -> "available", "low_stock"
func NormalizeAvailabilityStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

// Helper function to validate availability status enum
func IsValidAvailabilityStatus(status string) bool {
	validStatuses := map[string]bool{
		"available":   true,
		"unavailable": true,
		"low_stock":   true,
	}
	return validStatuses[status]
}

// Helper function to validate UUID format
func IsValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// Helper function to generate default availability status
func GetDefaultAvailabilityStatus() string {
	return "available"
}
