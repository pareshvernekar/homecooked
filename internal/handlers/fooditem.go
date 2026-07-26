package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/validation"
	"github.com/pareshvernekar/homecooked/internal/views"
)

// FoodItemHandler handles food item related operations
type FoodItemHandler struct {
	service FoodItemService
	Logger  *logger.Logger
}
type FoodItemService interface {
	List(ctx context.Context, tenantID string, page int, limit int) ([]*models.FoodItem, error)
	GetByID(ctx context.Context, id string, tenantID string) (*models.FoodItem, error)
	Create(ctx context.Context, createReq *models.FoodItemCreateRequest, tenantID string) (*models.FoodItem, error)
	Update(ctx context.Context, id string, updateReq *models.FoodItemUpdateRequest, tenantID string) error
	Delete(ctx context.Context, id string, tenantID string) (int64, error)
}

// NewFoodItemHandler creates a new instance of FoodItemHandler with dependency injection
func NewFoodItemHandler(service FoodItemService, l *logger.Logger) *FoodItemHandler {
	return &FoodItemHandler{
		service: service,
		Logger:  l,
	}
}

// GetFoodItems paginates and retrieves food items for the current tenant with caching support
func (h *FoodItemHandler) GetFoodItems(c *gin.Context) {
	page := int64(0)
	limit := int64(20)

	if c.Query("page") != "" {
		parsedPage, err := strconv.ParseInt(c.Query("page"), 10, 64)
		if err == nil && parsedPage >= 0 && parsedPage <= 1000 {
			page = parsedPage
		}
	}

	if c.Query("limit") != "" {
		parsedLimit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	tenantID := c.GetString(middleware.TenantIDKey)

	ctx := c.Request.Context()
	h.Logger.Info(ctx, "GetFoodItems: Fetching food items for tenant", "tenant_id", tenantID)

	var foodItems []*models.FoodItem

	foodItems, err := h.service.List(ctx, tenantID, int(page), int(limit))
	if err != nil {
		h.Logger.Error(ctx, "Failed to retrieve food items", "error", err)
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to retrieve food items", Timestamp: time.Now().UTC()})
		return
	}

	h.Logger.Info(ctx, "GetFoodItems: Successfully retrieved from database", "count", len(foodItems))
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Food items retrieved successfully", "data": foodItems})
}

// CreateFoodItem creates a new food item for the current tenant
func (h *FoodItemHandler) CreateFoodItem(c *gin.Context) {
	startTime := time.Now()
	h.Logger.Info(c.Request.Context(), "CreateFoodItem: Starting food item creation flow")
	ctx := c.Request.Context()
	var createRequest *models.FoodItemCreateRequest
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(c.Request.Context(), "CreateFoodItem: Failed to parse request JSON", "duration_ms", duration, "error", err.Error())
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "INVALID_INPUT", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	h.Logger.Info(c.Request.Context(), "CreateFoodItem: Request JSON parsed successfully", "name", createRequest.Name, "price", createRequest.Price)

	if err := validation.ValidateFoodItemCreate(createRequest); err != nil {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(c.Request.Context(), "CreateFoodItem: Validation failed", "duration_ms", duration, "validation_error", err.Error())
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "VALIDATION_ERROR", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	h.Logger.Info(c.Request.Context(), "CreateFoodItem: Validation passed", "name", createRequest.Name, "category", createRequest.CategoryName, "price", createRequest.Price)

	tenantID := c.GetString(middleware.TenantIDKey)

	// Normalize category name to lowercase slug format (e.g., "Vegetarian" -> "vegetarian")
	normalizedCategory := models.NormalizeCategory(createRequest.CategoryName)
	createRequest.CategoryName = normalizedCategory

	foodItem, err := h.service.Create(ctx, createRequest, tenantID)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(c.Request.Context(), "CreateFoodItem: Failed to create food item in database", "duration_ms", duration, "tenant_id", tenantID, "error", err.Error())
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to create food item", Timestamp: time.Now().UTC()})
		return
	}

	duration := time.Since(startTime).Milliseconds()
	h.Logger.Info(c.Request.Context(), "CreateFoodItem: Food item created successfully", "duration_ms", duration, "food_item_id", foodItem.ID, "tenant_id", tenantID)

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Food item created successfully"})
}

// UpdateFoodItem updates an existing food item (full update)
func (h *FoodItemHandler) UpdateFoodItem(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "MISSING_ID", Message: "Food item ID is required", Timestamp: time.Now().UTC()})
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	h.Logger.Debug(ctx, "UpdateFoodItem: Getting food item by ID", "id", id)

	var updateRequest *models.FoodItemUpdateRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		h.Logger.Error(ctx, "UpdateFoodItem: Failed to parse update request", "error", err)
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "INVALID_INPUT", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	if err := validation.ValidateFoodItemUpdate(updateRequest); err != nil {
		h.Logger.Error(ctx, "UpdateFoodItem: Validation failed for update", "error", err)
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "VALIDATION_ERROR", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	// Normalize category name to lowercase slug format (e.g., "Vegetarian" -> "vegetarian")
	normalizedCategory := models.NormalizeCategory(updateRequest.CategoryName)
	updateRequest.CategoryName = normalizedCategory

	if err := h.service.Update(ctx, id, updateRequest, tenantID); err != nil {
		h.Logger.Error(ctx, "UpdateFoodItem: Failed to update food item", "error", err)
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to update food item", Timestamp: time.Now().UTC()})
		return
	}

	h.Logger.Info(ctx, "UpdateFoodItem: Food item updated successfully", "id", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Food item updated successfully"})
}

// DeleteFoodItem deletes a specific food item
func (h *FoodItemHandler) DeleteFoodItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "MISSING_ID", Message: "Food item ID is required", Timestamp: time.Now().UTC()})
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	ctx := c.Request.Context()

	rowsDeleted, err := h.service.Delete(ctx, id, tenantID)
	if err != nil {
		h.Logger.Error(ctx, "DeleteFoodItem: Failed to delete food item", "error", err)
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to delete food item", Timestamp: time.Now().UTC()})
		return
	}

	h.Logger.Info(ctx, "DeleteFoodItem: Food item deleted successfully", "id", id, "rowsDeleted", rowsDeleted)
	c.JSON(http.StatusNoContent, nil)
}
