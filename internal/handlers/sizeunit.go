package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	apperrors "github.com/pareshvernekar/homecooked/internal/errors"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/views"
)

// SizeUnitService is the handler port for size units.
type SizeUnitService interface {
	List(ctx context.Context, tenantID string) ([]*models.SizeUnit, error)
	CreateCustom(ctx context.Context, tenantID string, req *models.SizeUnitCreateRequest) (*models.SizeUnit, error)
	DeactivateCustom(ctx context.Context, tenantID, id string) error
}

// SizeUnitHandler handles /api/v1/size-units.
// REQSIZE001, REQSIZE002
type SizeUnitHandler struct {
	service SizeUnitService
	Logger  *logger.Logger
}

// NewSizeUnitHandler constructs a size unit handler.
func NewSizeUnitHandler(service SizeUnitService, l *logger.Logger) *SizeUnitHandler {
	return &SizeUnitHandler{service: service, Logger: l}
}

func writeServiceError(c *gin.Context, err error, fallbackMsg string) {
	var se *apperrors.ServiceError
	if errors.As(err, &se) {
		c.JSON(se.GetStatusCode(), views.ErrorResponse{
			Success: false, ErrorCode: string(se.Code), Message: se.Message, Timestamp: time.Now().UTC(),
		})
		return
	}
	c.JSON(http.StatusInternalServerError, views.ErrorResponse{
		Success: false, ErrorCode: "DATABASE_ERROR", Message: fallbackMsg, Timestamp: time.Now().UTC(),
	})
}

// ListSizeUnits returns system standards and tenant customs.
func (h *SizeUnitHandler) ListSizeUnits(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := c.GetString(middleware.TenantIDKey)
	units, err := h.service.List(ctx, tenantID)
	if err != nil {
		writeServiceError(c, err, "Failed to list size units")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Size units retrieved successfully", "data": units})
}

// CreateSizeUnit creates a tenant custom size unit.
func (h *SizeUnitHandler) CreateSizeUnit(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := c.GetString(middleware.TenantIDKey)

	var req models.SizeUnitCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, views.ErrorResponse{
			Success: false, ErrorCode: "INVALID_REQUEST", Message: err.Error(), Timestamp: time.Now().UTC(),
		})
		return
	}

	unit, err := h.service.CreateCustom(ctx, tenantID, &req)
	if err != nil {
		writeServiceError(c, err, "Failed to create size unit")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Size unit created successfully", "data": unit})
}

// DeleteSizeUnit soft-deactivates a tenant custom size unit.
func (h *SizeUnitHandler) DeleteSizeUnit(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := c.GetString(middleware.TenantIDKey)
	id := c.Param("id")

	if err := h.service.DeactivateCustom(ctx, tenantID, id); err != nil {
		writeServiceError(c, err, "Failed to delete size unit")
		return
	}
	c.Status(http.StatusNoContent)
}
