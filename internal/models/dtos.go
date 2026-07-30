package models

import "time"

// =============================================================================
// Data Transfer Objects (DTOs) - API Request/Response Structures
// These are separate from entity models to ensure clean separation between
// database representation and API contracts.
// =============================================================================

// -----------------------------------------------------------------------------
// Food Item DTOs
// -----------------------------------------------------------------------------

// FoodItemResponse represents the food item in JSON responses
type FoodItemResponse struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	CategoryID   string    `json:"category_id"`
	Price        float64   `json:"price"`
	ImageURL     string    `json:"image_url,omitempty"`
	Avoidance    string    `json:"avoidance,omitempty"`
	IsVegetarian bool      `json:"is_vegetarian"`
	Status       string    `json:"availability_status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FoodItemPartialResponse represents a food item in list responses (minimal)
type FoodItemPartialResponse struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	CategoryID string    `json:"category_id"`
	Price      float64   `json:"price"`
	Status     string    `json:"availability_status"`
	CreatedAt  time.Time `json:"created_at"`
}

// -----------------------------------------------------------------------------
// Category DTOs
// -----------------------------------------------------------------------------

// CategoryResponse represents the food category in JSON responses
type CategoryResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// -----------------------------------------------------------------------------
// User DTOs
// -----------------------------------------------------------------------------

// UserResponse represents the user in JSON responses
type UserResponse struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// -----------------------------------------------------------------------------
// Weekly Menu DTOs
// -----------------------------------------------------------------------------

// WeeklyMenuResponse represents the weekly menu in JSON responses
type WeeklyMenuResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MenuItemWeeklyResponse represents a menu item in weekly menu response
type MenuItemWeeklyResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	MenuID      string    `json:"menu_id"`
	FoodItemID  string    `json:"food_item_id"`
	Description string    `json:"description,omitempty"`
	Size        string    `json:"size"`
	Price       float64   `json:"price"`
	Sequence    int       `json:"sequence"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WeeklyMenuWithItemsResponse represents weekly menu with its items
type WeeklyMenuWithItemsResponse struct {
	ID          string                   `json:"id"`
	TenantID    string                   `json:"tenant_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	StartDate   time.Time                `json:"start_date"`
	EndDate     time.Time                `json:"end_date"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	MenuItems   []MenuItemWeeklyResponse `json:"menu_items"`
}

// -----------------------------------------------------------------------------
// Catering Menu DTOs
// -----------------------------------------------------------------------------

// CateringMenuResponse represents the catering menu in JSON responses
type CateringMenuResponse struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	EventDate     time.Time `json:"event_date"`
	EventLocation string    `json:"event_location"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// MenuItemCateringResponse represents a menu item in catering menu response
type MenuItemCateringResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	MenuID      string    `json:"menu_id"`
	FoodItemID  string    `json:"food_item_id"`
	Description string    `json:"description,omitempty"`
	Size        string    `json:"size"`
	Price       float64   `json:"price"`
	Sequence    int       `json:"sequence"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CateringMenuWithItemsResponse represents catering menu with its items
type CateringMenuWithItemsResponse struct {
	ID            string                     `json:"id"`
	TenantID      string                     `json:"tenant_id"`
	Name          string                     `json:"name"`
	Description   string                     `json:"description,omitempty"`
	EventDate     time.Time                  `json:"event_date"`
	EventLocation string                     `json:"event_location"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
	MenuItems     []MenuItemCateringResponse `json:"menu_items"`
}

// -----------------------------------------------------------------------------
// Menu Item DTOs (Generic)
// -----------------------------------------------------------------------------

// MenuItemResponse represents a menu item in JSON responses
type MenuItemResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	MenuID      string    `json:"menu_id"`
	MenuType    string    `json:"menu_type"` // "weekly" or "catering"
	FoodItemID  string    `json:"food_item_id"`
	Description string    `json:"description,omitempty"`
	Size        string    `json:"size"`
	Price       float64   `json:"price"`
	Sequence    int       `json:"sequence"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// -----------------------------------------------------------------------------
// Order DTOs
// -----------------------------------------------------------------------------

// OrderResponse represents the order in JSON responses
type OrderResponse struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	UserID       string     `json:"user_id"`
	MenuType     string     `json:"menu_type"`
	MenuID       string     `json:"menu_id"`
	Status       string     `json:"status"`
	OrderDate    time.Time  `json:"order_date"`
	DeliveryDate *time.Time `json:"delivery_date,omitempty"`
	TotalPrice   float64    `json:"total_price"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// OrderItemResponse represents an order item in JSON responses
type OrderItemResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	OrderId     string    `json:"order_id"`
	MenuItemId  string    `json:"menu_item_id"`
	FoodItemID  string    `json:"food_item_id"`
	Quantity    int       `json:"quantity"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
}

// -----------------------------------------------------------------------------
// Notification DTOs
// -----------------------------------------------------------------------------

// NotificationResponse represents the notification in JSON responses
type NotificationResponse struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	OrderId   *string   `json:"order_id,omitempty"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// -----------------------------------------------------------------------------
// Pagination DTOs
// -----------------------------------------------------------------------------

// PaginatedResponse represents paginated API responses
type PaginatedResponse struct {
	Data       interface{} `json:"data,omitempty"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
	Message    string      `json:"message"`
}

// -----------------------------------------------------------------------------
// Success Response DTOs
// -----------------------------------------------------------------------------

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Page       int         `json:"page,omitempty"`
	TotalItems int64       `json:"total_items,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

// -----------------------------------------------------------------------------
// Error Response DTOs
// -----------------------------------------------------------------------------

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success   bool        `json:"success"`
	ErrorCode string      `json:"error_code"`
	Message   string      `json:"message"`
	Detail    interface{} `json:"detail,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ValidationError represents validation error response
type ValidationError struct {
	Success   bool        `json:"success"`
	Field     string      `json:"field"`
	Message   string      `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
	Detail    interface{} `json:"detail,omitempty"`
}
