package sizeunit

import (
	"context"
	"fmt"
	"strings"

	"github.com/pareshvernekar/homecooked/internal/errors"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// Repository is the size-unit persistence port.
type Repository interface {
	ListActive(ctx context.Context) ([]*models.SizeUnit, error)
	GetByID(ctx context.Context, id string) (*models.SizeUnit, error)
	CreateCustom(ctx context.Context, code, displayName string) (*models.SizeUnit, error)
	SoftDeactivateCustom(ctx context.Context, id string) error
}

// Service handles size unit business logic.
// REQSIZE001–REQSIZE003
type Service struct {
	repo   Repository
	logger *logger.Logger
}

// NewService constructs a size unit service.
func NewService(repo Repository, l *logger.Logger) *Service {
	return &Service{repo: repo, logger: l}
}

// List returns system + tenant custom active units.
func (s *Service) List(ctx context.Context, tenantID string) ([]*models.SizeUnit, error) {
	units, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, errors.CreateDatabaseError("Failed to list size units", err, tenantID)
	}
	return units, nil
}

// CreateCustom creates a tenant-defined size unit.
// REQSIZE002
func (s *Service) CreateCustom(ctx context.Context, tenantID string, req *models.SizeUnitCreateRequest) (*models.SizeUnit, error) {
	code := strings.TrimSpace(req.Code)
	displayName := strings.TrimSpace(req.DisplayName)
	if code == "" || displayName == "" {
		return nil, errors.CreateValidationError("code and display_name are required", map[string]interface{}{
			"field": "code",
		})
	}

	unit, err := s.repo.CreateCustom(ctx, code, displayName)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") ||
			strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, errors.CreateValidationError(
				fmt.Sprintf("custom size unit code %q already exists for tenant", code),
				map[string]interface{}{"field": "code"},
			)
		}
		return nil, errors.CreateDatabaseError("Failed to create custom size unit", err, tenantID)
	}
	return unit, nil
}

// DeactivateCustom soft-deletes a tenant custom unit; system units are rejected.
// REQSIZE001S02, REQSIZE002
func (s *Service) DeactivateCustom(ctx context.Context, tenantID, id string) error {
	if err := s.repo.SoftDeactivateCustom(ctx, id); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not found") || strings.Contains(msg, "not deletable") {
			return errors.CreateValidationError(
				"size unit cannot be deleted (system units are immutable; customs must belong to tenant)",
				map[string]interface{}{"id": id},
			)
		}
		return errors.CreateDatabaseError("Failed to deactivate size unit", err, tenantID)
	}
	return nil
}

// EnsureUsableByTenant verifies a size unit id is a system unit or the tenant's custom.
// REQSIZE003
func (s *Service) EnsureUsableByTenant(ctx context.Context, tenantID, sizeUnitID string) error {
	_, err := s.repo.GetByID(ctx, sizeUnitID)
	if err != nil {
		return errors.CreateValidationError(
			fmt.Sprintf("size unit %q is unknown or not available to tenant", sizeUnitID),
			map[string]interface{}{"size_unit_id": sizeUnitID},
		)
	}
	return nil
}
