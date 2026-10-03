package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// NotificationService port for tenant settings and the per-order notification audit.
type NotificationService interface {
	GetSettings(ctx context.Context, tenantID string) (*models.TenantSettings, error)
	UpdateSettings(ctx context.Context, tenantID string, req *models.TenantSettingsUpdateRequest) (*models.TenantSettings, error)
	ListByOrder(ctx context.Context, tenantID, orderID string) ([]models.NotificationOutbox, error)
}

// NotificationHandler serves /api/v1/tenant/settings and /api/v1/orders/:id/notifications.
// REQNOTIF001, REQNOTIF005
type NotificationHandler struct {
	svc    NotificationService
	Logger *logger.Logger
}

// NewNotificationHandler constructs a notification handler.
func NewNotificationHandler(svc NotificationService, l *logger.Logger) *NotificationHandler {
	return &NotificationHandler{svc: svc, Logger: l}
}

// GetTenantSettings GET /tenant/settings
// REQNOTIF001
func (h *NotificationHandler) GetTenantSettings(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	settings, err := h.svc.GetSettings(c.Request.Context(), tenantID)
	if err != nil {
		writeServiceError(c, err, "Failed to get tenant settings")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Tenant settings retrieved successfully", "data": settings})
}

// UpdateTenantSettings PUT /tenant/settings {cook_admin_phone}
// REQNOTIF001
func (h *NotificationHandler) UpdateTenantSettings(c *gin.Context) {
	var req models.TenantSettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	settings, err := h.svc.UpdateSettings(c.Request.Context(), tenantID, &req)
	if err != nil {
		writeServiceError(c, err, "Failed to update tenant settings")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Tenant settings updated successfully", "data": settings})
}

// ListOrderNotifications GET /orders/:id/notifications
// REQNOTIF005
func (h *NotificationHandler) ListOrderNotifications(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	rows, err := h.svc.ListByOrder(c.Request.Context(), tenantID, c.Param("id"))
	if err != nil {
		writeServiceError(c, err, "Failed to list notifications")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Notifications retrieved successfully", "data": rows})
}
