package models
package models

import (
	"time"

	"gorm.io/gorm"
)

// FoodItem represents a food menu item in the database
type FoodItem struct {
	ID             string    `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID       string    `gorm:"type:uuid;not null" json:"tenant_id"`
	Name           string    `gorm:"size:200;not null;unique:tenant_id" json:"name"`
	Description    *string   `gorm:"type:text" json:"description,omitempty"`
	Category       string    `gorm:"type:varchar(100);not null;index:idx_category_tenant" json:"category"`
	Price          float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	ImageURL       *string    `gorm:"type:text" json:"image_url,omitempty"`
	Avoidance      string    `gorm:"type:jsonb;default:'[]'" json:"avoidance,omitempty"`
	IsVegetarian   bool       `gorm:"default:false" json:"is_vegetarian"`
	AvailabilityStatus string  `gorm:"type:enum('available','unavailable','low_stock');not null" json:"availability_status"`
	CreatedAt      *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not null" json:"created_at,omitempty"`
	UpdatedAt      *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at,omitempty"`
	DeliveredAt    *time.Time `gorm:"type:timestamp" json:"delivered_at,omitempty"`
}

// FoodItemCreateRequest represents the request payload for creating a food item
type FoodItemCreateRequest struct {
	Name            string  `json:"name" binding:"required"`
	Description     *string `json:"description,omitempty"`
	Category        string   `json:"category" binding:"required"`
	Price           float64  `json:"price" binding:"required,gte=0"`
	ImageURL        *string  `json:"image_url,omitempty"`
	Avoidance       *string  `json:"avoidance,omitempty"`
	IsVegetarian     *bool    `json:"is_vegetarian,omitempty"`
	AvailabilityStatus string `json:"availability_status" binding:"required,in:available|unavailable|low_stock"`
}

// FoodItemUpdateRequest represents the request payload for updating a food item
type FoodItemUpdateRequest struct {
	Name            string   `json:"name,omitempty"`
	Description     *string   `json:"description,omitempty"`
	Category        string    `json:"category,omitempty"`
	Price           float64   `json:"price,omitempty"`
	ImageURL        *string    `json:"image_url,omitempty"`
	Avoidance       *string    `json:"avoidance,omitempty"`
	IsVegetarian     *bool      `json:"is_vegetarian,omitempty"`
	AvailabilityStatus string `json:"availability_status,omitempty"`
}

// Helper function to validate category enum
func IsValidCategory(category string) bool {
	validCategories := map[string]bool{
		"vegetarian":  true,
		"non-vegetarian": true,
		"vegan":       true,
		"dessert":     true,
		"beverage":    true,
		"appetizer":   true,
		"main_course": true,
		"sides":       true,
	}
	return validCategories[category]
}

// Helper function to validate availability status enum
func IsValidAvailabilityStatus(status string) bool {
	validStatuses := map[string]bool{
		"available":         true,
		"unavailable":       true,
		"low_stock":         true,
	}
	return validStatuses[status]
}

// Helper function to validate UUID format
func IsValidUUID(uuid string) bool {
	_, err := uuid.Parse(uuid)
	return err == nil
}

// Helper function to generate default availability status
func GetDefaultAvailabilityStatus() string {
	return "available"
}

// Helper function to get available food items for a tenant
func (m *FoodItem) GetAvailableItems(db *gorm.DB, tenantID string) error {
	return db.Model(&FoodItem{}).
		Where("tenant_id = ? AND availability_status = 'available'", tenantID).
		Find(this).Error
}

// Helper function to get food items by category for a tenant
func (m *FoodItem) GetByCategory(db *gorm.DB, tenantID string, category string) error {
	return db.Model(&FoodItem{}).
		Where("tenant_id = ? AND category = ? AND availability_status = 'available'", 
			tenantID, category).
		Find(this).Error
}