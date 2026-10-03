package notification

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	apperrors "github.com/pareshvernekar/homecooked/internal/errors"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// Store is the persistence needed for tenant settings and the per-order audit list.
type Store interface {
	GetCookAdminPhone(ctx context.Context, tenantID string) (string, error)
	UpdateCookAdminPhone(ctx context.Context, tenantID, phone string) error
	ListByOrder(ctx context.Context, tenantID, orderID string) ([]models.NotificationOutbox, error)
	OrderExists(ctx context.Context, tenantID, orderID string) (bool, error)
}

const maxPhoneLen = 50

const cookPhoneUnsetWarning = "cook_admin_phone is not set; order.created alerts to the cook are skipped"

// Service serves tenant notification settings and the notification audit list.
// REQNOTIF001, REQNOTIF005
type Service struct {
	store Store
}

// NewService constructs the settings/audit service.
func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) settings(phone string) *models.TenantSettings {
	out := &models.TenantSettings{}
	if phone == "" {
		out.Warnings = []string{cookPhoneUnsetWarning}
		return out
	}
	out.CookAdminPhone = &phone
	return out
}

// GetSettings returns the current tenant's settings.
// REQNOTIF001
func (s *Service) GetSettings(ctx context.Context, tenantID string) (*models.TenantSettings, error) {
	phone, err := s.store.GetCookAdminPhone(ctx, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.CreateNotFoundError("tenant", tenantID, tenantID)
		}
		return nil, apperrors.CreateDatabaseError("Failed to load tenant settings", err, tenantID)
	}
	return s.settings(phone), nil
}

// UpdateSettings sets or clears cook_admin_phone (null or empty clears).
// REQNOTIF001, REQNOTIF001S01, REQNOTIF001S02
func (s *Service) UpdateSettings(ctx context.Context, tenantID string, req *models.TenantSettingsUpdateRequest) (*models.TenantSettings, error) {
	if !req.CookAdminPhone.Set {
		return nil, apperrors.CreateValidationError("cook_admin_phone is required (use null or an empty string to clear)",
			map[string]interface{}{"field": "cook_admin_phone"})
	}
	phone := ""
	if req.CookAdminPhone.Value != nil {
		phone = strings.TrimSpace(*req.CookAdminPhone.Value)
	}
	if len(phone) > maxPhoneLen {
		return nil, apperrors.CreateValidationError("cook_admin_phone must be at most 50 characters",
			map[string]interface{}{"field": "cook_admin_phone"})
	}
	if err := s.store.UpdateCookAdminPhone(ctx, tenantID, phone); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.CreateNotFoundError("tenant", tenantID, tenantID)
		}
		return nil, apperrors.CreateDatabaseError("Failed to update tenant settings", err, tenantID)
	}
	return s.settings(phone), nil
}

// ListByOrder lists notification outbox rows for an order of the current tenant.
// An order that does not belong to the tenant is NotFound (no cross-tenant exposure).
// REQNOTIF005, REQNOTIF005S01
func (s *Service) ListByOrder(ctx context.Context, tenantID, orderID string) ([]models.NotificationOutbox, error) {
	ok, err := s.store.OrderExists(ctx, tenantID, orderID)
	if err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to load order", err, tenantID)
	}
	if !ok {
		return nil, apperrors.CreateNotFoundError("order", orderID, tenantID)
	}
	rows, err := s.store.ListByOrder(ctx, tenantID, orderID)
	if err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to list notifications", err, tenantID)
	}
	return rows, nil
}
