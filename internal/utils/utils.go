package utils

import "github.com/google/uuid"

// GenerateUUID generates a new RFC 4122 UUID version 4 (randomly generated)
// This is the standard UUID format used in databases for primary keys
func GenerateUUID() string {
	return uuid.New().String()
}

// Example usage:
//
// foodItem := models.FoodItem{
//     ID:      utils.GenerateUUID(), // Auto-generate ID when creating new item
//     Name:     "Classic Burger",
//     Price:    9.99,
//     Category: "burgers",
// }

// GenerateToken generates a UUID for use as JWT token or session identifier
func GenerateToken() string {
	return uuid.New().String()
}

// GenerateOrderID generates an order-specific ID
func GenerateOrderID() string {
	return uuid.New().String()
}
