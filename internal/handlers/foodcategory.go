package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	middleware "github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/views"
)

// FoodCategoryHandler handles food category related operations
type FoodCategoryHandler struct {
	service FoodCategoryService
	Logger  *logger.Logger
}

// NewFoodCategoryHandler creates a new instance of FoodCategoryHandler with dependency injection
func NewFoodCategoryHandler(service FoodCategoryService, l *logger.Logger) *FoodCategoryHandler {
	return &FoodCategoryHandler{
		service: service,
		Logger:  l,
	}
}

type FoodCategoryService interface {
	ListCategories(ctx context.Context, tenantID string) ([]*models.FoodCategory, error)
	GetCategoryByID(ctx context.Context, tenantID string, id string) (*models.FoodCategory, error)
	CreateCategory(ctx context.Context, category *models.FoodCategory) error
	UpdateCategory(ctx context.Context, category *models.FoodCategory) (int64, error)
	DeleteCategory(ctx context.Context, tenantID string, id string) (int64, error)
}

// ListCategories retrieves all food categories for the current tenant with caching support
func (h *FoodCategoryHandler) ListCategories(c *gin.Context) {
	startTime := time.Now()
	h.Logger.Info(c.Request.Context(), "ListCategories: Fetching food categories for tenant")
	ctx := c.Request.Context()

	tenantID := c.GetString(middleware.TenantIDKey)
	categories, err := h.service.ListCategories(ctx, tenantID)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(ctx, "ListCategories: Failed to retrieve food categories", "duration_ms", duration, "error", err.Error())
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to retrieve food categories", Timestamp: time.Now().UTC()})
		return
	}

	duration := time.Since(startTime).Milliseconds()
	h.Logger.Info(ctx, "ListCategories: Successfully retrieved food categories", "duration_ms", duration, "count", len(categories))

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Food categories retrieved successfully", "data": categories})
}

// GetCategoryByID retrieves a specific food category by its ID for the current tenant
func (h *FoodCategoryHandler) GetCategoryByID(c *gin.Context) {
	startTime := time.Now()
	id := c.Param("id")
	if id == "" {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(c.Request.Context(), "GetCategoryByID: Missing ID parameter", "duration_ms", duration)
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "MISSING_ID", Message: "Category ID is required", Timestamp: time.Now().UTC()})
		return
	}

	h.Logger.Debug(c.Request.Context(), "GetCategoryByID: Getting food category by ID", "id", id)
	ctx := c.Request.Context()

	category, err := h.service.GetCategoryByID(ctx, id, c.GetString(middleware.TenantIDKey))
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(ctx, "GetCategoryByID: Failed to retrieve food category", "duration_ms", duration, "id", id, "error", err.Error())
		c.JSON(http.StatusNotFound, views.ErrorResponse{Success: false, ErrorCode: "ENTITY_NOT_FOUND", Message: "Food category not found", Timestamp: time.Now().UTC()})
		return
	}

	duration := time.Since(startTime).Milliseconds()
	h.Logger.Info(ctx, "GetCategoryByID: Successfully retrieved food category", "duration_ms", duration, "id", id)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Food category retrieved successfully", "data": category})
}

// CreateCategory creates a new food category for the current tenant
func (h *FoodCategoryHandler) CreateCategory(c *gin.Context) {
	startTime := time.Now()
	h.Logger.Info(c.Request.Context(), "CreateCategory: Starting food category creation flow")
	ctx := c.Request.Context()

	var createRequest map[string]interface{}
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(ctx, "CreateCategory: Failed to parse request JSON", "duration_ms", duration, "error", err.Error())
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "INVALID_INPUT", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	h.Logger.Info(ctx, "CreateCategory: Request JSON parsed successfully")

	categoryName := ""
	if val, ok := createRequest["category_name"]; ok && val != nil {
		categoryName = val.(string)
	}

	description := ""
	if desc, ok := createRequest["description"]; ok && desc != nil {
		d := desc.(string)
		description = d
	}

	category := &models.FoodCategory{
		ID:          "",
		TenantID:    c.GetString(middleware.TenantIDKey),
		Name:        categoryName,
		Description: description,
		IsActive:    true,
		CreatedAt:   time.Now().UTC().UnixMilli(),
		UpdatedAt:   time.Now().UTC().UnixMilli(),
	}

	if err := h.service.CreateCategory(ctx, category); err != nil {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(ctx, "CreateCategory: Failed to create food category in database", "duration_ms", duration, "error", err.Error())
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to create food category", Timestamp: time.Now().UTC()})
		return
	}

	duration := time.Since(startTime).Milliseconds()
	h.Logger.Info(ctx, "CreateCategory: Food category created successfully", "duration_ms", duration)

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Food category created successfully"})
}

// UpdateCategory updates an existing food category (full update)
func (h *FoodCategoryHandler) UpdateCategory(c *gin.Context) {
	startTime := time.Now()
	id := c.Param("id")
	if id == "" {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(c.Request.Context(), "UpdateCategory: Missing ID parameter", "duration_ms", duration)
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "MISSING_ID", Message: "Category ID is required", Timestamp: time.Now().UTC()})
		return
	}

	h.Logger.Debug(c.Request.Context(), "UpdateCategory: Getting food category by ID", "id", id)
	ctx := c.Request.Context()

	var updateRequest map[string]interface{}
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(ctx, "UpdateCategory: Failed to parse update request", "duration_ms", duration, "error", err.Error())
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "INVALID_INPUT", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	name := ""
	if val, ok := updateRequest["name"]; ok && val != nil {
		name = val.(string)
	}

	description := ""
	if desc, ok := updateRequest["description"]; ok && desc != nil {
		d := desc.(string)
		description = d
	}

	isActive := true
	if isActiveVal, ok := updateRequest["is_active"]; ok && isActiveVal != nil {
		switch v := isActiveVal.(type) {
		case bool:
			isActive = v
		}
	}

	category := &models.FoodCategory{
		ID:          id,
		TenantID:    c.GetString(middleware.TenantIDKey),
		Name:        name,
		Description: description,
		IsActive:    isActive,
		CreatedAt:   time.Now().UTC().UnixMilli(),
		UpdatedAt:   time.Now().UTC().UnixMilli(),
	}

	if _, err := h.service.UpdateCategory(ctx, category); err != nil {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(ctx, "UpdateCategory: Failed to update food category", "duration_ms", duration, "error", err.Error())
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to update food category", Timestamp: time.Now().UTC()})
		return
	}

	duration := time.Since(startTime).Milliseconds()
	h.Logger.Info(ctx, "UpdateCategory: Food category updated successfully", "duration_ms", duration)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Food category updated successfully"})
}

// DeleteCategory deletes a specific food category (soft delete - sets is_active = false)
func (h *FoodCategoryHandler) DeleteCategory(c *gin.Context) {
	startTime := time.Now()
	id := c.Param("id")
	if id == "" {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(c.Request.Context(), "DeleteCategory: Missing ID parameter", "duration_ms", duration)
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "MISSING_ID", Message: "Category ID is required", Timestamp: time.Now().UTC()})
		return
	}

	h.Logger.Debug(c.Request.Context(), "DeleteCategory: Deleting food category", "id", id)
	ctx := c.Request.Context()

	rowsDeleted, err := h.service.DeleteCategory(ctx, c.GetString(middleware.TenantIDKey), id)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		h.Logger.Error(ctx, "DeleteCategory: Failed to delete food category", "duration_ms", duration, "error", err.Error())
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to delete food category", Timestamp: time.Now().UTC()})
		return
	}

	duration := time.Since(startTime).Milliseconds()
	h.Logger.Info(ctx, "DeleteCategory: Food category deleted successfully", "duration_ms", duration, "rowsDeleted", rowsDeleted)

	c.JSON(http.StatusNoContent, nil)
}
