package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/validation"
	"github.com/pareshvernekar/homecooked/internal/views"
)

// FoodItemHandler handles food item related operations
type FoodItemHandler struct {
	Repo repository.FoodItemRepository
}

// NewFoodItemHandler creates a new instance of FoodItemHandler with dependency injection
func NewFoodItemHandler(repo repository.FoodItemRepository) *FoodItemHandler {
	return &FoodItemHandler{
		Repo: repo,
	}
}

// GetFoodItems paginates and retrieves food items for the current tenant
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

	foodItems, total, err := h.Repo.ListByTenant(tenantID, int(page), int(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to retrieve food items", Timestamp: time.Now().UTC()})
		return
	}
	fmt.Println("Total food items for tenant:", total)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Food items retrieved successfully", "data": foodItems})
}

// CreateFoodItem creates a new food item for the current tenant
func (h *FoodItemHandler) CreateFoodItem(c *gin.Context) {
	var createRequest models.FoodItemCreateRequest
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "INVALID_INPUT", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	// Use custom validation to validate enum values like AvailabilityStatus
	if err := validation.ValidateFoodItemCreate(createRequest); err != nil {
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "VALIDATION_ERROR", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

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
		UpdatedAt:   &createdTime,
	}

	if err := h.Repo.Create(&foodItem); err != nil {
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to create food item", Timestamp: time.Now().UTC()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Food item created successfully"})
}

// UpdateFoodItem updates an existing food item (full update)
func (h *FoodItemHandler) UpdateFoodItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "MISSING_ID", Message: "Food item ID is required", Timestamp: time.Now().UTC()})
		return
	}

	existingFoodItem, err := h.Repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, views.ErrorResponse{Success: false, ErrorCode: "NOT_FOUND", Message: "Food item not found", Timestamp: time.Now().UTC()})
		return
	}

	var updateRequest models.FoodItemUpdateRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "INVALID_INPUT", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	if err := validation.ValidateFoodItemUpdate(updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "VALIDATION_ERROR", Message: "Invalid request body", Timestamp: time.Now().UTC(), Detail: err.Error()})
		return
	}

	existingFoodItem.Name = updateRequest.Name
	existingFoodItem.Description = updateRequest.Description
	existingFoodItem.Price = updateRequest.Price
	existingFoodItem.Category = updateRequest.Category
	existingFoodItem.AvailabilityStatus = updateRequest.AvailabilityStatus
	updatedTime := time.Now().UTC()
	existingFoodItem.UpdatedAt = &updatedTime

	if err := h.Repo.Update(existingFoodItem); err != nil {
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to update food item", Timestamp: time.Now().UTC()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Food item updated successfully"})
}

// DeleteFoodItem deletes a specific food item
func (h *FoodItemHandler) DeleteFoodItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, views.ErrorResponse{Success: false, ErrorCode: "MISSING_ID", Message: "Food item ID is required", Timestamp: time.Now().UTC()})
		return
	}

	existingFoodItem, err := h.Repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, views.ErrorResponse{Success: false, ErrorCode: "NOT_FOUND", Message: "Food item not found", Timestamp: time.Now().UTC()})
		return
	}

	if err := h.Repo.Delete(existingFoodItem.ID); err != nil {
		c.JSON(http.StatusInternalServerError, views.ErrorResponse{Success: false, ErrorCode: "DATABASE_ERROR", Message: "Failed to delete food item", Timestamp: time.Now().UTC()})
		return
	}

	c.Status(http.StatusNoContent)
}
