package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	cache "github.com/pareshvernekar/homecooked/internal/cache"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/validation"
	"github.com/pareshvernekar/homecooked/internal/views"
)

// FoodItemHandler handles food item related operations
type FoodItemHandler struct {
	Repo       repository.FoodItemRepository
	Logger      *logger.Logger
	CacheClient cache.TypedClient[models.FoodItem]
}

// NewFoodItemHandler creates a new instance of FoodItemHandler with dependency injection
func NewFoodItemHandler(repo repository.FoodItemRepository, l *logger.Logger, c cache.TypedClient[models.FoodItem]) *FoodItemHandler {
	return &FoodItemHandler{
		Repo:       repo,
		Logger:     l,
		CacheClient: c,
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

	var foodItems []models.FoodItem

	// Attempt to load from cache if cache client is available
	if h.CacheClient != nil {
		cacheKey := fmt.Sprintf("food_item:%s", tenantID)
		_, err := h.CacheClient.Get(ctx, cacheKey)

		if err == nil {
			// Cache hit - return cached value
			// Since we can't iterate a single FoodItem in Get(), we fall back to DB for lists
			h.Logger.Info(ctx, "GetFoodItems: Single item retrieved from cache", "cache_key", cacheKey)
			return
		}

		h.Logger.Debug(ctx, "GetFoodItems: Cache miss - loading from database")
	}

	// Cache miss or no cache - load from database
	foodItems, total, err := h.Repo.ListByTenant(tenantID, int(page), int(limit))
	if err != nil {
		h.Logger.Error(ctx, "Failed to retrieve food items", "error", err)
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to retrieve food items", Timestamp: time.Now().UTC()})
		return
	}

	h.Logger.Info(ctx, "GetFoodItems: Successfully retrieved from database", "count", len(foodItems), "total", total)

	// Update cache with new data if available
	for i := range foodItems {
		cacheKey := fmt.Sprintf("food_item:%s", strings.ReplaceAll(foodItems[i].ID, "-", "_"))
		err = h.CacheClient.Set(ctx, cacheKey, foodItems[i], cache.WithTTL(30*time.Minute))
		if err != nil {
			h.Logger.Error(ctx, "Failed to update cache", "error", err, "item_id", foodItems[i].ID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Food items retrieved successfully", "data": foodItems})
}

// CreateFoodItem creates a new food item for the current tenant
func (h *FoodItemHandler) CreateFoodItem(c *gin.Context) {
	startTime := time.Now()
	h.Logger.Info(c.Request.Context(), "CreateFoodItem: Starting food item creation flow")

	var createRequest models.FoodItemCreateRequest
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

	h.Logger.Info(c.Request.Context(), "CreateFoodItem: Validation passed", "name", createRequest.Name, "category", createRequest.Category)

	tenantID := c.GetString(middleware.TenantIDKey)
	createdTime := time.Now().UTC()

	foodItem := models.FoodItem{
		ID:          uuid.NewString(),
		Name:        createRequest.Name,
		Description: createRequest.Description,
		Price:       createRequest.Price,
		Category:    createRequest.Category,
		TenantID:    tenantID,
		CreatedAt:   &createdTime,
		UpdatedAt:    &createdTime,
	}

	h.Logger.Info(c.Request.Context(), "CreateFoodItem: Generated UUID for food item", "uuid", foodItem.ID)

	if err := h.Repo.Create(&foodItem); err != nil {
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
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "MISSING_ID", Message: "Food item ID is required", Timestamp: time.Now().UTC()})
		return
	}

	h.Logger.Debug(c.Request.Context(), "UpdateFoodItem: Getting food item by ID", "id", id)

	foodItem, err := h.Repo.GetByID(id)
	if err != nil {
		h.Logger.Error(c.Request.Context(), "UpdateFoodItem: Food item not found", "id", id, "error", err)
		c.JSON(http.StatusNotFound, views.ErrorResponse{Success: false, ErrorCode: "NOT_FOUND", Message: "Food item not found", Timestamp: time.Now().UTC()})
		return
	}

	var updateRequest models.FoodItemUpdateRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		h.Logger.Error(c.Request.Context(), "UpdateFoodItem: Failed to parse update request", "error", err)
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "INVALID_INPUT", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	if err := validation.ValidateFoodItemUpdate(updateRequest); err != nil {
		h.Logger.Error(c.Request.Context(), "UpdateFoodItem: Validation failed for update", "error", err)
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "VALIDATION_ERROR", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	ctx := c.Request.Context()

	foodItem.Name = updateRequest.Name
	foodItem.Description = updateRequest.Description
	foodItem.Price = updateRequest.Price
	foodItem.Category = updateRequest.Category
	foodItem.AvailabilityStatus = updateRequest.AvailabilityStatus
	updatedTime := time.Now().UTC()
	foodItem.UpdatedAt = &updatedTime

	if err := h.Repo.Update(foodItem); err != nil {
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

	ctx := c.Request.Context()
	h.Logger.Debug(ctx, "DeleteFoodItem: Getting food item by ID", "id", id)

	foodItem, err := h.Repo.GetByID(id)
	if err != nil {
		h.Logger.Error(ctx, "DeleteFoodItem: Food item not found", "id", id, "error", err)
		c.JSON(http.StatusNotFound, views.ErrorResponse{Success: false, ErrorCode: "NOT_FOUND", Message: "Food item not found", Timestamp: time.Now().UTC()})
		return
	}

	if foodItem.TenantID != c.GetString(middleware.TenantIDKey) {
		h.Logger.Warn(ctx, "DeleteFoodItem: Attempted to delete food item outside tenant scope")
		c.JSON(http.StatusForbidden, views.ErrorResponse{Success: false, ErrorCode: "FORBIDDEN", Message: "Cannot delete food item outside tenant scope", Timestamp: time.Now().UTC()})
		return
	}

	if err := h.Repo.Delete(foodItem.ID); err != nil {
		h.Logger.Error(ctx, "DeleteFoodItem: Failed to delete food item", "error", err)
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to delete food item", Timestamp: time.Now().UTC()})
		return
	}

	h.Logger.Info(ctx, "DeleteFoodItem: Food item deleted successfully", "id", id)
	c.JSON(http.StatusNoContent, nil)
}
