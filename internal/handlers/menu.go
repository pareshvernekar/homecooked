package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// MenuService port for HTTP handlers.
type MenuService interface {
	Create(ctx context.Context, tenantID string, req *models.MenuCreateRequest) (*models.Menu, error)
	List(ctx context.Context, tenantID, status string) ([]*models.Menu, error)
	GetTree(ctx context.Context, tenantID, id string) (*models.MenuTree, error)
	Update(ctx context.Context, tenantID, id string, req *models.MenuUpdateRequest) error
	Deactivate(ctx context.Context, tenantID, id string) error
	Publish(ctx context.Context, tenantID, id string) error
	Unpublish(ctx context.Context, tenantID, id string) error
}

// MenuItemService port for nested item/size-option routes.
type MenuItemService interface {
	AddItem(ctx context.Context, tenantID, menuID string, req *models.MenuItemCreateRequest) (*models.MenuItem, error)
	RemoveItem(ctx context.Context, tenantID, menuID, itemID string) error
	AddSizeOption(ctx context.Context, tenantID, menuID, itemID, componentID string, req *models.SizeOptionCreateRequest) (*models.MenuItemComponentSizeOption, error)
	UpdateSizeOption(ctx context.Context, tenantID, menuID, itemID, componentID, optionID string, req *models.SizeOptionUpdateRequest) error
	RemoveSizeOption(ctx context.Context, tenantID, menuID, itemID, componentID, optionID string) error
}

// MenuHandler serves /api/v1/menus.
// REQMENU001–REQMENU009, REQITEM001–REQITEM006
type MenuHandler struct {
	menus MenuService
	items MenuItemService
	Logger *logger.Logger
}

// NewMenuHandler constructs a menu handler.
func NewMenuHandler(menus MenuService, items MenuItemService, l *logger.Logger) *MenuHandler {
	return &MenuHandler{menus: menus, items: items, Logger: l}
}

// CreateMenu POST /menus
func (h *MenuHandler) CreateMenu(c *gin.Context) {
	var req models.MenuCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "INVALID_REQUEST", "message": err.Error(), "timestamp": time.Now().UTC()})
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	m, err := h.menus.Create(c.Request.Context(), tenantID, &req)
	if err != nil {
		writeServiceError(c, err, "Failed to create menu")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Menu created successfully", "data": m})
}

// ListMenus GET /menus?status=
func (h *MenuHandler) ListMenus(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	status := c.Query("status")
	menus, err := h.menus.List(c.Request.Context(), tenantID, status)
	if err != nil {
		writeServiceError(c, err, "Failed to list menus")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Menus retrieved successfully", "data": menus})
}

// GetMenu GET /menus/:id
func (h *MenuHandler) GetMenu(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	tree, err := h.menus.GetTree(c.Request.Context(), tenantID, c.Param("id"))
	if err != nil {
		writeServiceError(c, err, "Failed to get menu")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Menu retrieved successfully", "data": tree})
}

// UpdateMenu PUT /menus/:id
func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	var req models.MenuUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "INVALID_REQUEST", "message": err.Error(), "timestamp": time.Now().UTC()})
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	if err := h.menus.Update(c.Request.Context(), tenantID, c.Param("id"), &req); err != nil {
		writeServiceError(c, err, "Failed to update menu")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Menu updated successfully"})
}

// DeleteMenu DELETE /menus/:id
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	if err := h.menus.Deactivate(c.Request.Context(), tenantID, c.Param("id")); err != nil {
		writeServiceError(c, err, "Failed to delete menu")
		return
	}
	c.Status(http.StatusNoContent)
}

// PublishMenu POST /menus/:id/publish
func (h *MenuHandler) PublishMenu(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	if err := h.menus.Publish(c.Request.Context(), tenantID, c.Param("id")); err != nil {
		writeServiceError(c, err, "Failed to publish menu")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Menu published successfully"})
}

// UnpublishMenu POST /menus/:id/unpublish
func (h *MenuHandler) UnpublishMenu(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	if err := h.menus.Unpublish(c.Request.Context(), tenantID, c.Param("id")); err != nil {
		writeServiceError(c, err, "Failed to unpublish menu")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Menu unpublished successfully"})
}

// AddMenuItem POST /menus/:id/items
func (h *MenuHandler) AddMenuItem(c *gin.Context) {
	var req models.MenuItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "INVALID_REQUEST", "message": err.Error(), "timestamp": time.Now().UTC()})
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	item, err := h.items.AddItem(c.Request.Context(), tenantID, c.Param("id"), &req)
	if err != nil {
		writeServiceError(c, err, "Failed to add menu item")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Menu item created successfully", "data": item})
}

// DeleteMenuItem DELETE /menus/:id/items/:itemId
func (h *MenuHandler) DeleteMenuItem(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	if err := h.items.RemoveItem(c.Request.Context(), tenantID, c.Param("id"), c.Param("itemId")); err != nil {
		writeServiceError(c, err, "Failed to remove menu item")
		return
	}
	c.Status(http.StatusNoContent)
}

// AddSizeOption POST /menus/:id/items/:itemId/components/:componentId/size-options
func (h *MenuHandler) AddSizeOption(c *gin.Context) {
	var req models.SizeOptionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "INVALID_REQUEST", "message": err.Error(), "timestamp": time.Now().UTC()})
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	opt, err := h.items.AddSizeOption(c.Request.Context(), tenantID, c.Param("id"), c.Param("itemId"), c.Param("componentId"), &req)
	if err != nil {
		writeServiceError(c, err, "Failed to add size option")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Size option created successfully", "data": opt})
}

// UpdateSizeOption PUT .../size-options/:optionId
func (h *MenuHandler) UpdateSizeOption(c *gin.Context) {
	var req models.SizeOptionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "INVALID_REQUEST", "message": err.Error(), "timestamp": time.Now().UTC()})
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	if err := h.items.UpdateSizeOption(c.Request.Context(), tenantID, c.Param("id"), c.Param("itemId"), c.Param("componentId"), c.Param("optionId"), &req); err != nil {
		writeServiceError(c, err, "Failed to update size option")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Size option updated successfully"})
}

// DeleteSizeOption DELETE .../size-options/:optionId
func (h *MenuHandler) DeleteSizeOption(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	if err := h.items.RemoveSizeOption(c.Request.Context(), tenantID, c.Param("id"), c.Param("itemId"), c.Param("componentId"), c.Param("optionId")); err != nil {
		writeServiceError(c, err, "Failed to remove size option")
		return
	}
	c.Status(http.StatusNoContent)
}
