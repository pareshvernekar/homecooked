package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pareshvernekar/homecooked/internal/models"
)

var validate = validator.New()

// Register validators for food items
func init() {
	err := validate.RegisterValidation("validCategory", validCategory)
	if err != nil {
		panic(err)
	}
}

// validCategory validates category field against allowed values
func validCategory(fl validator.FieldLevel) bool {
	return models.IsValidCategory(fl.Field().String())
}

// ValidateFoodItemCreate validates a food item creation request
func ValidateFoodItemCreate(data *models.FoodItemCreateRequest) error {
	// Field: Name (required string)
	if err := validate.Var(data.Name, "required"); err != nil {
		return fmt.Errorf("name %v", err)
	}

	// Field: CategoryName (required + custom validation)
	if !models.IsValidCategory(data.CategoryName) {
		return fmt.Errorf("category_name %v is invalid", data.CategoryName)
	}

	// IsVegetarian is optional but required when provided
	if data.IsVegetarian == nil {
		return nil
	}
	if err := validate.Var(data.IsVegetarian, "required"); err != nil {
		return fmt.Errorf("is_vegetarian %v", err)
	}

	// Field: AvailabilityStatus (required string)
	if err := validate.Var(data.AvailabilityStatus, "required"); err != nil {
		return fmt.Errorf("availability_status %v", err)
	}

	return nil
}

// ValidateFoodItemUpdate validates a food item update request using partial update semantics.
// Only validates fields that are present (non-empty for strings, non-zero for numbers).
// This differs from full update semantics where ALL core fields must be provided.
func ValidateFoodItemUpdate(data *models.FoodItemUpdateRequest) error {
	// Name validation - required if provided, validated only when not empty
	// Partial update: name can be omitted to keep existing value
	if data.Name != "" {
		if err := validate.Var(data.Name, "required"); err != nil {
			return fmt.Errorf("name is required but cannot be empty: %v", err)
		}
	}

	// CategoryName validation - required if provided, validated against enum list
	// Partial update: category_name can be omitted to keep existing value
	if data.CategoryName != "" {
		if !models.IsValidCategory(data.CategoryName) {
			return fmt.Errorf("category_name must be one of: vegetarian, non-vegetarian, vegan, dessert, beverage, appetizer, main_course, sides")
		}
	}

	// Description - optional field, no validation (omitting keeps existing description)

	// ImageURL - optional field, no validation (omitting keeps existing image URL)

	// Avoidance - optional field, no validation (omitting keeps existing avoidance/restriction)

	// IsVegetarian - boolean pointer, required if provided
	// Partial update: can be nil to use existing value
	if data.IsVegetarian != nil {
		if err := validate.Var(data.IsVegetarian, "required"); err != nil {
			return fmt.Errorf("is_vegetarian must be either true or false when provided: %v", err)
		}
	}

	// AvailabilityStatus validation - required if provided (not empty), validated against enum
	// Partial update: can be omitted to keep existing status
	if data.AvailabilityStatus != "" {
		validStatuses := map[string]bool{
			"available":   true,
			"unavailable": true,
			"low_stock":   true,
		}
		if !validStatuses[data.AvailabilityStatus] {
			return fmt.Errorf("availability_status must be one of: available, unavailable, low_stock")
		}
	}

	return nil
}

// IsValidAvailabilityStatus validates if a status is one of the allowed values
func IsValidAvailabilityStatus(status string) bool {
	validStatuses := map[string]bool{
		"available":   true,
		"unavailable": true,
		"low_stock":   true,
	}
	return validStatuses[status]
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	return validate
}
