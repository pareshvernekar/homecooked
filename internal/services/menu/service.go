package menu

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	apperrors "github.com/pareshvernekar/homecooked/internal/errors"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// MenuRepository port.
type MenuRepository interface {
	Create(ctx context.Context, m *models.Menu, categories []models.MenuCategoryInput) error
	List(ctx context.Context, statusFilter string) ([]*models.Menu, error)
	GetActiveByID(ctx context.Context, id string) (*models.Menu, error)
	GetTree(ctx context.Context, id string) (*models.MenuTree, error)
	UpdateMetadata(ctx context.Context, m *models.Menu) error
	ReplaceCategories(ctx context.Context, menuID string, categories []models.MenuCategoryInput) error
	Deactivate(ctx context.Context, id string) error
	HasPublishedWeeklyStartDate(ctx context.Context, startDate string, excludeMenuID string) (bool, error)
	SetStatus(ctx context.Context, id, status string) error
	CountActiveCategories(ctx context.Context, menuID string) (int, error)
	CountActiveItems(ctx context.Context, menuID string) (int, error)
	ListActiveItemsForPublish(ctx context.Context, menuID string) ([]models.MenuItem, error)
}

// FoodItemChecker verifies active food items for publish gate.
type FoodItemChecker interface {
	FoodItemActiveInTenant(ctx context.Context, foodItemID string) (bool, error)
}

// Service implements menu lifecycle.
// REQMENU001–REQMENU009
type Service struct {
	repo     MenuRepository
	foodCheck FoodItemChecker
	logger   *logger.Logger
}

// NewService constructs a menu service.
func NewService(repo MenuRepository, foodCheck FoodItemChecker, l *logger.Logger) *Service {
	return &Service{repo: repo, foodCheck: foodCheck, logger: l}
}

// Create creates a draft menu.
// REQMENU001
func (s *Service) Create(ctx context.Context, tenantID string, req *models.MenuCreateRequest) (*models.Menu, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.CreateValidationError("name is required", map[string]interface{}{"field": "name"})
	}
	menuType := strings.ToLower(strings.TrimSpace(req.MenuType))
	if menuType != models.MenuTypeDaily && menuType != models.MenuTypeWeekly && menuType != models.MenuTypeCatering {
		return nil, apperrors.CreateValidationError("menu_type must be daily, weekly, or catering", map[string]interface{}{"field": "menu_type"})
	}

	m := &models.Menu{
		Name: name, Description: req.Description, MenuType: menuType,
		MenuDate: req.MenuDate, StartDate: req.StartDate, EndDate: req.EndDate,
		EventDate: req.EventDate, EventLocation: req.EventLocation,
	}
	if err := s.repo.Create(ctx, m, req.Categories); err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to create menu", err, tenantID)
	}
	return m, nil
}

// List lists menus with optional status filter.
// REQMENU002
func (s *Service) List(ctx context.Context, tenantID, status string) ([]*models.Menu, error) {
	menus, err := s.repo.List(ctx, status)
	if err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to list menus", err, tenantID)
	}
	return menus, nil
}

// GetTree returns nested menu structure.
// REQMENU003
func (s *Service) GetTree(ctx context.Context, tenantID, id string) (*models.MenuTree, error) {
	tree, err := s.repo.GetTree(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.CreateNotFoundError("menu", id, tenantID)
		}
		return nil, apperrors.CreateDatabaseError("Failed to get menu", err, tenantID)
	}
	return tree, nil
}

// Update updates draft menu metadata.
// REQMENU004
func (s *Service) Update(ctx context.Context, tenantID, id string, req *models.MenuUpdateRequest) error {
	m, err := s.requireDraft(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if req.Name != nil {
		n := strings.TrimSpace(*req.Name)
		if n == "" {
			return apperrors.CreateValidationError("name cannot be empty", map[string]interface{}{"field": "name"})
		}
		m.Name = n
	}
	if req.Description != nil {
		m.Description = *req.Description
	}
	if req.MenuDate != nil {
		m.MenuDate = req.MenuDate
	}
	if req.StartDate != nil {
		m.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		m.EndDate = req.EndDate
	}
	if req.EventDate != nil {
		m.EventDate = req.EventDate
	}
	if req.EventLocation != nil {
		m.EventLocation = req.EventLocation
	}
	if err := s.repo.UpdateMetadata(ctx, m); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.CreateValidationError("menu is not editable while published or inactive", map[string]interface{}{"id": id})
		}
		return apperrors.CreateDatabaseError("Failed to update menu", err, tenantID)
	}
	if req.Categories != nil {
		if err := s.repo.ReplaceCategories(ctx, id, req.Categories); err != nil {
			return apperrors.CreateDatabaseError("Failed to update categories", err, tenantID)
		}
	}
	return nil
}

// Deactivate soft-deletes a menu.
// REQMENU004
func (s *Service) Deactivate(ctx context.Context, tenantID, id string) error {
	if err := s.repo.Deactivate(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.CreateNotFoundError("menu", id, tenantID)
		}
		return apperrors.CreateDatabaseError("Failed to deactivate menu", err, tenantID)
	}
	return nil
}

// Publish validates structural completeness and weekly uniqueness then publishes.
// REQMENU007, REQMENU009
func (s *Service) Publish(ctx context.Context, tenantID, id string) error {
	m, err := s.repo.GetActiveByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.CreateNotFoundError("menu", id, tenantID)
		}
		return apperrors.CreateDatabaseError("Failed to load menu", err, tenantID)
	}
	if m.Status == models.MenuStatusPublished {
		return apperrors.CreateValidationError("menu is already published", map[string]interface{}{"id": id})
	}
	if err := s.validatePublishGate(ctx, m); err != nil {
		return err
	}
	if m.MenuType == models.MenuTypeWeekly && m.StartDate != nil && *m.StartDate != "" {
		taken, err := s.repo.HasPublishedWeeklyStartDate(ctx, *m.StartDate, id)
		if err != nil {
			return apperrors.CreateDatabaseError("Failed to check weekly start_date uniqueness", err, tenantID)
		}
		if taken {
			return apperrors.CreateValidationError(
				fmt.Sprintf("an active published weekly menu already uses start_date %s", *m.StartDate),
				map[string]interface{}{"field": "start_date"},
			)
		}
	}
	if err := s.repo.SetStatus(ctx, id, models.MenuStatusPublished); err != nil {
		return apperrors.CreateDatabaseError("Failed to publish menu", err, tenantID)
	}
	return nil
}

// Unpublish moves published → draft.
// REQMENU008
func (s *Service) Unpublish(ctx context.Context, tenantID, id string) error {
	m, err := s.repo.GetActiveByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.CreateNotFoundError("menu", id, tenantID)
		}
		return apperrors.CreateDatabaseError("Failed to load menu", err, tenantID)
	}
	if m.Status != models.MenuStatusPublished {
		return apperrors.CreateValidationError("only published menus can be unpublished", map[string]interface{}{"id": id})
	}
	if err := s.repo.SetStatus(ctx, id, models.MenuStatusDraft); err != nil {
		return apperrors.CreateDatabaseError("Failed to unpublish menu", err, tenantID)
	}
	return nil
}

func (s *Service) requireDraft(ctx context.Context, tenantID, id string) (*models.Menu, error) {
	m, err := s.repo.GetActiveByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.CreateNotFoundError("menu", id, tenantID)
		}
		return nil, apperrors.CreateDatabaseError("Failed to load menu", err, tenantID)
	}
	if m.Status != models.MenuStatusDraft {
		return nil, apperrors.CreateValidationError("menu is not editable while published", map[string]interface{}{"id": id})
	}
	return m, nil
}

func (s *Service) validatePublishGate(ctx context.Context, m *models.Menu) error {
	switch m.MenuType {
	case models.MenuTypeDaily:
		if m.MenuDate == nil || strings.TrimSpace(*m.MenuDate) == "" {
			return apperrors.CreateValidationError("daily menu requires menu_date to publish", map[string]interface{}{"field": "menu_date"})
		}
	case models.MenuTypeWeekly:
		if m.StartDate == nil || *m.StartDate == "" || m.EndDate == nil || *m.EndDate == "" {
			return apperrors.CreateValidationError("weekly menu requires start_date and end_date to publish", map[string]interface{}{"field": "start_date"})
		}
	case models.MenuTypeCatering:
		if m.EventDate == nil || *m.EventDate == "" || m.EventLocation == nil || strings.TrimSpace(*m.EventLocation) == "" {
			return apperrors.CreateValidationError("catering menu requires event_date and event_location to publish", map[string]interface{}{"field": "event_date"})
		}
	}

	cats, err := s.repo.CountActiveCategories(ctx, m.ID)
	if err != nil {
		return apperrors.CreateDatabaseError("Failed to count categories", err, m.TenantID)
	}
	if cats < 1 {
		return apperrors.CreateValidationError("menu requires at least one category to publish", nil)
	}
	itemCount, err := s.repo.CountActiveItems(ctx, m.ID)
	if err != nil {
		return apperrors.CreateDatabaseError("Failed to count items", err, m.TenantID)
	}
	if itemCount < 1 {
		return apperrors.CreateValidationError("menu requires at least one menu-item to publish", nil)
	}

	items, err := s.repo.ListActiveItemsForPublish(ctx, m.ID)
	if err != nil {
		return apperrors.CreateDatabaseError("Failed to load items for publish", err, m.TenantID)
	}
	for _, item := range items {
		if err := validateItemStructure(item); err != nil {
			return err
		}
		for _, c := range item.Components {
			if s.foodCheck != nil {
				ok, err := s.foodCheck.FoodItemActiveInTenant(ctx, c.FoodItemID)
				if err != nil {
					return apperrors.CreateDatabaseError("Failed to verify food item", err, m.TenantID)
				}
				if !ok {
					return apperrors.CreateValidationError(
						fmt.Sprintf("component food_item %s is inactive or missing", c.FoodItemID),
						map[string]interface{}{"food_item_id": c.FoodItemID},
					)
				}
			}
		}
	}
	return nil
}

func validateItemStructure(item models.MenuItem) error {
	n := len(item.Components)
	switch item.Kind {
	case models.MenuItemKindSimple:
		if n != 1 {
			return apperrors.CreateValidationError("simple menu-item must have exactly one component", map[string]interface{}{"item_id": item.ID})
		}
	case models.MenuItemKindCombo:
		if n < 2 {
			return apperrors.CreateValidationError("combo menu-item must have at least two components", map[string]interface{}{"item_id": item.ID})
		}
	default:
		return apperrors.CreateValidationError("invalid menu-item kind", map[string]interface{}{"kind": item.Kind})
	}
	for _, c := range item.Components {
		if len(c.SizeOptions) < 1 {
			return apperrors.CreateValidationError("each component requires at least one size option", map[string]interface{}{"component_id": c.ID})
		}
		if c.DefaultSizeOptionID == nil || *c.DefaultSizeOptionID == "" {
			return apperrors.CreateValidationError("each component requires a default size option", map[string]interface{}{"component_id": c.ID})
		}
		found := false
		for _, o := range c.SizeOptions {
			if o.ID == *c.DefaultSizeOptionID {
				found = true
				break
			}
		}
		if !found {
			return apperrors.CreateValidationError("default size option must belong to the component", map[string]interface{}{"component_id": c.ID})
		}
	}
	return nil
}

// GetActiveMenu returns an active menu (draft or published) for the tenant.
// REQITEM006, REQORDER001
func (s *Service) GetActiveMenu(ctx context.Context, tenantID, id string) (*models.Menu, error) {
	m, err := s.repo.GetActiveByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.CreateNotFoundError("menu", id, tenantID)
		}
		return nil, apperrors.CreateDatabaseError("Failed to load menu", err, tenantID)
	}
	return m, nil
}

// RequireDraftMenu exposes draft check for menu-item service.
func (s *Service) RequireDraftMenu(ctx context.Context, tenantID, id string) (*models.Menu, error) {
	return s.requireDraft(ctx, tenantID, id)
}
