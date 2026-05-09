package handlers

import (
	"net/http"
	"time"

	"github.com/homecooked/api/internal/config"
	"github.com/homecooked/api/internal/database"
	"github.com/homecooked/api/internal/middleware"
	"github.com/homecooked/api/internal/models"
	"github.com/homecooked/api/internal/utils"

	"github.com/gin-gonic/gin"
)

// FoodItemHandler handles food item related operations
type FoodItemHandler struct {
	DB        *database.DB
	Config     *config.Config
	RequestID string
}

// NewFoodItemHandler creates a new instance of FoodItemHandler
func NewFoodItemHandler(db *database.DB, cfg *config.Config) *FoodItemHandler {
	return &FoodItemHandler{
		DB:    db,
		Config: cfg,
		RequestID: utils.GenerateUUID(),
	}
}

// GetFoodItems paginates and retrieves food items for the current tenant
func (h *FoodItemHandler) GetFoodItems(c *gin.Context) {
	var page int64 = 1
	var limit int64 = 20

	// Parse query parameters
	if c.Query("page") != "" {
		parsedPage, err := strconv.ParseInt(c.Query("page"), 10, 64)
		if err == nil && parsedPage > 0 && parsedPage <= 1000 {
			page = parsedPage
		}
	}

	if c.Query("limit") != "" {
		parsedLimit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Apply tenant-based pagination with RLS
	query := h.DB.Model(&models.FoodItem{}).
		Joins("tenant").
		Where("tenants.tenant_id = ?", middleware.GetTenantID(c)).
		Offset((page - 1) * int(limit)).
		Limit(int(limit)).
		Order("created_at DESC")

	var foodItems []models.FoodItem
	if err := query.Find(&foodItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve food items",
			RequestID: h.RequestID,
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data:   foodItems,
		Page:   int(page),
		Limit:  int(limit),
		Total:  len(foodItems),
	})
}

// CreateFoodItem creates a new food item for the current tenant
func (h *FoodItemHandler) CreateFoodItem(c *gin.Context) {
	var foodItem models.FoodItemCreateRequest
	if err := c.ShouldBindJSON(&foodItem); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Code:    "INVALID_INPUT",
			Message: "Invalid request body",
			RequestID: h.RequestID,
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if foodItem.Name == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Code:    "MISSING_FIELD",
			Message: "Field 'name' is required",
			RequestID: h.RequestID,
		})
		return
	}

	if foodItem.Price <= 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Code:    "INVALID_PRICE",
			Message: "Price must be greater than zero",
			RequestID: h.RequestID,
		})
		return
	}

	if foodItem.Category == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Code:    "MISSING_FIELD",
			Message: "Field 'category' is required",
			RequestID: h.RequestID,
		})
		return
	}

	// Auto-generate ID if not provided (UUID v4)
	if foodItem.ID == "" {
		foodItem.ID = uuid.New().String()
	} else {
		// Validate UUID format
		if _, err := uuid.Parse(foodItem.ID); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Code:    "INVALID_ID",
				Message: "Invalid ID format (must be UUID)",
				RequestID: h.RequestID,
			})
			return
		}
	}

	// Auto-generate timestamps
	now := time.Now().UTC()
(foodItem.CreatedAt = &now)
(foodItem.UpdatedAt = &now)

	// Check if food item with same ID already exists (for update scenarios via POST)
	var existing models.FoodItem
	if err := h.DB.First(&existing, foodItem.ID).Error; err == nil && !h.DB.Error {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Code:    "DUPLICATE_ID",
			Message: "Food item with this ID already exists",
			RequestID: h.RequestID,
		})
		return
	}

	// Set tenant context for RLS (implicit via database connection)
	tenantID := middleware.GetTenantID(c)
	foodItem.TenantID = tenantID

	// Create the food item
	result := h.DB.Create(&foodItem)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to create food item",
			RequestID: h.RequestID,
		})
		return
	}

	// Set updated timestamp after creation
	now = time.Now().UTC()
	foodItem.UpdatedAt = &now

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Data:       foodItem,
		Message:    "Food item created successfully",
		RequestID: h.RequestID,
	})
}

// UpdateFoodItem updates an existing food item (full update)
func (h *FoodItemHandler) UpdateFoodItem(c *gin.Context) {
	var foodItem models.FoodItemUpdateRequest
	if err := c.ShouldBindJSON(&foodItem); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Code:    "INVALID_INPUT",
			Message: "Invalid request body",
			RequestID: h.RequestID,
			Error:   err.Error(),
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Code:    "MISSING_ID",
			Message: "Food item ID is required",
			RequestID: h.RequestID,
		})
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid ID format (must be UUID)",
			RequestID: h.RequestID,
		})
		return
	}

	// Check if food item exists
	var existing models.FoodItem
	if err := h.DB.First(&existing, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Food item not found",
			RequestID: h.RequestID,
		})
		return
	}

	// Copy updateable fields (excluding ID)
	if foodItem.Name != "" {
		existing.Name = foodItem.Name
	}
	if foodItem.Description != "" {
		existing.Description = foodItem.Description
	}
	if foodItem.Category != "" {
		existing.Category = foodItem.Category
	}
	if foodItem.Price > 0 {
		existing.Price = foodItem.Price
	}

	// Update timestamp
	now := time.Now().UTC()
	existing.UpdatedAt = &now

	result := h.DB.Save(&existing)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to update food item",
			RequestID: h.RequestID,
		})
		return
	}

	// Reload updated entity
	h.DB.First(&existing, id)

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Data:       existing,
		Message:    "Food item updated successfully",
		RequestID: h.RequestID,
	})
}

// DeleteFoodItem deletes a specific food item
func (h *FoodItemHandler) DeleteFoodItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Code:    "MISSING_ID",
			Message: "Food item ID is required",
			RequestID: h.RequestID,
		})
		return
	}

	// Check if food item exists
	var existing models.FoodItem
	if err := h.DB.First(&existing, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Food item not found",
			RequestID: h.RequestID,
		})
		return
	}

	result := h.DB.Delete(&existing)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "Failed to delete food item",
			RequestID: h.RequestID,
		})
		return
	}

	c.Status(http.StatusNoContent)
}