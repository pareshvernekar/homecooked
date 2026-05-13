package validation

import (
	"errors"

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
func ValidateFoodItemCreate(data models.FoodItemCreateRequest) error {
	// Field: Name (required string)
	if err := validate.Var(data.Name, "required"); err != nil {
		return err
	}

	// Field: Category (required + custom validation)
	if !models.IsValidCategory(data.Category) {
		return errors.New("category must be one of: vegetarian, non-vegetarian, vegan, dessert, beverage, appetizer, main_course, sides")
	}

	// Field: Price (required, >= 0)
	if err := validate.Var(data.Price, "required"); err != nil {
		return err
	}
	if data.Price < 0 {
		return errors.New("price must be greater than or equal to 0")
	}

	// Field: ImageURL (optional pointer)
	_ = validate.Var(data.ImageURL, "omitempty")

	// Field: Avoidance (optional pointer)
	if data.Avoidance != nil {
		_ = validate.Var(*data.Avoidance, "required")
	}

	// Field: IsVegetarian (optional bool)
	_ = validate.Var(data.IsVegetarian, "omitempty")

	// Field: AvailabilityStatus (required string)
	if err := validate.Var(data.AvailabilityStatus, "required"); err != nil {
		return err
	}

	return nil

}

// ValidateFoodItemUpdate validates a food item update request
func ValidateFoodItemUpdate(data models.FoodItemUpdateRequest) error {
	rules := `omitempty,dive,name|required,categories|required,price|gte=0,image_url|nullable,avoidance|nullable,is_vegetarian|nullable,availability_status|`
	errs := validate.Var(data, rules)
	return errs
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
