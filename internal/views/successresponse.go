package views

import (
	"net/http"
	"time"
)

// SuccessResponse represents a standardized success response format
type SuccessResponse struct {
	Success    bool            `json:"success"`
	Message    string          `json:"message"`
	Data       interface{}     `json:"data,omitempty"`
	Pagination *PaginationInfo `json:"pagination,omitempty"` // For list responses only
	RequestID  string          `json:"request_id,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasMore    bool `json:"has_more,omitempty"`
}

// SuccessResponse is a helper function that creates a success response with single item
func CreateSuccessResponse(data interface{}, message string) SuccessResponse {
	return SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

// CreatePaginatedResponse is a helper function that creates a paginated success response
func CreatePaginatedResponse(data interface{}, page, limit, total int) SuccessResponse {
	totalPages := (total + limit - 1) / limit

	return SuccessResponse{
		Success: true,
		Message: "Items retrieved successfully",
		Data:    data,
		Pagination: &PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasMore:    (total % limit) != 0,
		},
		Timestamp: time.Now().UTC(),
	}
}

// GetStatusCode returns the HTTP status code for the given SuccessResponse
func (s *SuccessResponse) GetStatusCode() int {
	if s == nil {
		return http.StatusOK
	}

	// Typically success responses are 200, but we can customize based on data type
	if s.Data == nil {
		return http.StatusOK // Just confirmation, no data returned
	}

	return http.StatusCreated // For created resources (201)
}

// StatusCode returns the appropriate HTTP status code for this response
func (s *SuccessResponse) StatusCode() int {
	if s.Pagination != nil {
		return http.StatusOK // 200 for list operations
	}

	return http.StatusCreated // 201 for single resource creation
}

// Example usage:
//
// func (h *FoodItemHandler) CreateFoodItem(c *gin.Context) {
//     // ... create food item logic ...
//
//     // Create success response
//     response := utils.CreateSuccessResponse(foodItem, "Food item created successfully")
//
//     // Return the appropriate status code
//     statusCode := response.GetStatusCode() // returns 201 for created items
//     c.JSON(statusCode, response)
// }
