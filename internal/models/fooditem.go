package models

import (
	"time"

	"github.com/google/uuid"
)

// FoodItem represents a food menu item in the database
type FoodItem struct {
	ID                  string        `json:"id"`
	TenantID            string        `json:"tenant_id"`
	Name                string        `json:"name"`
	Description         *string       `json:"description,omitempty"`
	Category            string        `json:"category"`
	Price               float64       `json:"price"`
	ImageURL            *string       `json:"image_url,omitempty"`
	Avoidance          *string       `json:"avoidance,omitempty"`
	IsVegetarian        bool          `json:"is_vegetarian"`
	AvailabilityStatus  string        `json:"availability_status"`
	CreatedAt           *time.Time    `json:"created_at,omitempty"`
	UpdatedAt           *time.Time    `json:"updated_at,omitempty"`
	DeliveredAt         *time.Time    `json:"delivered_at,omitempty"`
}

// FoodItemCreateRequest represents the request payload for creating a food item
type FoodItemCreateRequest struct {
	Name               string    `json:"name" binding:"required"`
	Description        *string   `json:"description,omitempty"`
	Category           string    `json:"category" binding:"required"`
	Price              float64   `json:"price" binding:"required,gte=0"`
	ImageURL           *string   `json:"image_url,omitempty"`
	Avoidance         *string    `json:"avoidance,omitempty"`
	IsVegetarian      *bool      `json:"is_vegetarian,omitempty"`
	AvailabilityStatus string     `json:"availability_status" binding:"required"`
}

// FoodItemUpdateRequest represents the request payload for updating a food item
type FoodItemUpdateRequest struct {
	Name               string    `json:"name,omitempty"`
	Description        *string   `json:"description,omitempty"`
	Category           string    `json:"category,omitempty"`
	Price              float64   `json:"price,omitempty"`
	ImageURL           *string   `json:"image_url,omitempty"`
	Avoidance         *string    `json:"avoidance,omitempty"`
	IsVegetarian      *bool      `json:"is_vegetarian,omitempty"`
	AvailabilityStatus string     `json:"availability_status,omitempty"`
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
