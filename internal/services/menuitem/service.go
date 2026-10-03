package menuitem

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

// MenuGate checks draft/active menu for mutations.
type MenuGate interface {
	RequireDraftMenu(ctx context.Context, tenantID, id string) (*models.Menu, error)
	GetActiveMenu(ctx context.Context, tenantID, id string) (*models.Menu, error)
}

// SizeUnitGate validates size unit references.
type SizeUnitGate interface {
	EnsureUsableByTenant(ctx context.Context, tenantID, sizeUnitID string) error
}

// Repository port for menu items.
type Repository interface {
	CategoryBelongsToMenu(ctx context.Context, menuID, categoryID string) (bool, error)
	FoodItemActiveInTenant(ctx context.Context, foodItemID string) (bool, error)
	CreateItemWithComponents(ctx context.Context, item *models.MenuItem, components []models.MenuItemComponentInput) error
	SoftDeleteItem(ctx context.Context, menuID, itemID string) error
	GetActiveItem(ctx context.Context, menuID, itemID string) (*models.MenuItem, error)
	AddSizeOption(ctx context.Context, componentID string, in models.SizeOptionCreateRequest) (*models.MenuItemComponentSizeOption, error)
	UpdateSizeOption(ctx context.Context, componentID, optionID string, in models.SizeOptionUpdateRequest) error
	SoftDeleteSizeOption(ctx context.Context, componentID, optionID string) error
	GetComponentOnMenu(ctx context.Context, menuID, itemID, componentID string) (*models.MenuItemComponent, error)
}

// Service implements menu-item mutations on draft menus.
// REQITEM001–REQITEM006, REQSIZE003
type Service struct {
	repo     Repository
	menuGate MenuGate
	sizeGate SizeUnitGate
	logger   *logger.Logger
}

// NewService constructs a menu-item service.
func NewService(repo Repository, menuGate MenuGate, sizeGate SizeUnitGate, l *logger.Logger) *Service {
	return &Service{repo: repo, menuGate: menuGate, sizeGate: sizeGate, logger: l}
}

// AddItem adds a simple or combo item to a draft menu.
// REQITEM001
func (s *Service) AddItem(ctx context.Context, tenantID, menuID string, req *models.MenuItemCreateRequest) (*models.MenuItem, error) {
	if _, err := s.menuGate.RequireDraftMenu(ctx, tenantID, menuID); err != nil {
		return nil, err
	}
	kind := strings.ToLower(strings.TrimSpace(req.Kind))
	if kind != models.MenuItemKindSimple && kind != models.MenuItemKindCombo {
		return nil, apperrors.CreateValidationError("kind must be simple or combo", map[string]interface{}{"field": "kind"})
	}
	n := len(req.Components)
	if kind == models.MenuItemKindSimple && n != 1 {
		return nil, apperrors.CreateValidationError("simple item requires exactly one component", nil)
	}
	if kind == models.MenuItemKindCombo && n < 2 {
		return nil, apperrors.CreateValidationError("combo item requires at least two components", nil)
	}

	ok, err := s.repo.CategoryBelongsToMenu(ctx, menuID, req.CategoryID)
	if err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to verify category", err, tenantID)
	}
	if !ok {
		return nil, apperrors.CreateValidationError("category not found on menu", map[string]interface{}{"category_id": req.CategoryID})
	}

	for i, c := range req.Components {
		active, err := s.repo.FoodItemActiveInTenant(ctx, c.FoodItemID)
		if err != nil {
			return nil, apperrors.CreateDatabaseError("Failed to verify food item", err, tenantID)
		}
		if !active {
			return nil, apperrors.CreateValidationError(
				fmt.Sprintf("food_item %s inactive or not found", c.FoodItemID),
				map[string]interface{}{"component_index": i},
			)
		}
		if len(c.SizeOptions) < 1 {
			return nil, apperrors.CreateValidationError("each component requires size_options", map[string]interface{}{"component_index": i})
		}
		for j, so := range c.SizeOptions {
			if so.Qty <= 0 {
				return nil, apperrors.CreateValidationError("size option qty must be > 0", map[string]interface{}{"component_index": i, "size_index": j})
			}
			if so.Price < 0 {
				return nil, apperrors.CreateValidationError("size option price must be >= 0", map[string]interface{}{"component_index": i, "size_index": j})
			}
			if s.sizeGate != nil {
				if err := s.sizeGate.EnsureUsableByTenant(ctx, tenantID, so.SizeUnitID); err != nil {
					return nil, err
				}
			}
		}
	}

	item := &models.MenuItem{
		MenuID: menuID, CategoryID: req.CategoryID, Kind: kind,
		Name: strings.TrimSpace(req.Name), Description: req.Description, Sequence: req.Sequence,
	}
	if item.Name == "" {
		return nil, apperrors.CreateValidationError("name is required", map[string]interface{}{"field": "name"})
	}
	if err := s.repo.CreateItemWithComponents(ctx, item, req.Components); err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to create menu item", err, tenantID)
	}
	return item, nil
}

// RemoveItem soft-deletes a menu-item from a draft menu.
// REQITEM002, REQITEM006
func (s *Service) RemoveItem(ctx context.Context, tenantID, menuID, itemID string) error {
	if _, err := s.menuGate.RequireDraftMenu(ctx, tenantID, menuID); err != nil {
		return err
	}
	if err := s.repo.SoftDeleteItem(ctx, menuID, itemID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.CreateNotFoundError("menu_item", itemID, tenantID)
		}
		return apperrors.CreateDatabaseError("Failed to remove menu item", err, tenantID)
	}
	return nil
}

// AddSizeOption adds a size option on a draft menu component.
func (s *Service) AddSizeOption(ctx context.Context, tenantID, menuID, itemID, componentID string, req *models.SizeOptionCreateRequest) (*models.MenuItemComponentSizeOption, error) {
	if _, err := s.menuGate.RequireDraftMenu(ctx, tenantID, menuID); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetComponentOnMenu(ctx, menuID, itemID, componentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.CreateNotFoundError("component", componentID, tenantID)
		}
		return nil, apperrors.CreateDatabaseError("Failed to load component", err, tenantID)
	}
	if req.Qty <= 0 {
		return nil, apperrors.CreateValidationError("qty must be > 0", map[string]interface{}{"field": "qty"})
	}
	if req.Price < 0 {
		return nil, apperrors.CreateValidationError("price must be >= 0", map[string]interface{}{"field": "price"})
	}
	if s.sizeGate != nil {
		if err := s.sizeGate.EnsureUsableByTenant(ctx, tenantID, req.SizeUnitID); err != nil {
			return nil, err
		}
	}
	opt, err := s.repo.AddSizeOption(ctx, componentID, *req)
	if err != nil {
		return nil, apperrors.CreateDatabaseError("Failed to add size option", err, tenantID)
	}
	return opt, nil
}

// UpdateSizeOption updates a size option on a draft menu. On a published menu only a
// price-only update is allowed so unfulfilled orders can reflect live prices.
// REQITEM006, REQITEM006S03
func (s *Service) UpdateSizeOption(ctx context.Context, tenantID, menuID, itemID, componentID, optionID string, req *models.SizeOptionUpdateRequest) error {
	menu, err := s.menuGate.GetActiveMenu(ctx, tenantID, menuID)
	if err != nil {
		return err
	}
	if menu.Status == models.MenuStatusPublished {
		if req.Price == nil || req.SizeUnitID != nil || req.Qty != nil || req.IsDefault != nil {
			return apperrors.CreateValidationError(
				"menu is not editable while published; only size option price may be updated",
				map[string]interface{}{"id": menuID},
			)
		}
	}
	if _, err := s.repo.GetComponentOnMenu(ctx, menuID, itemID, componentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.CreateNotFoundError("component", componentID, tenantID)
		}
		return apperrors.CreateDatabaseError("Failed to load component", err, tenantID)
	}
	if req.Qty != nil && *req.Qty <= 0 {
		return apperrors.CreateValidationError("qty must be > 0", map[string]interface{}{"field": "qty"})
	}
	if req.Price != nil && *req.Price < 0 {
		return apperrors.CreateValidationError("price must be >= 0", map[string]interface{}{"field": "price"})
	}
	if req.SizeUnitID != nil && s.sizeGate != nil {
		if err := s.sizeGate.EnsureUsableByTenant(ctx, tenantID, *req.SizeUnitID); err != nil {
			return err
		}
	}
	if err := s.repo.UpdateSizeOption(ctx, componentID, optionID, *req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.CreateNotFoundError("size_option", optionID, tenantID)
		}
		return apperrors.CreateDatabaseError("Failed to update size option", err, tenantID)
	}
	return nil
}

// RemoveSizeOption soft-deletes a size option.
func (s *Service) RemoveSizeOption(ctx context.Context, tenantID, menuID, itemID, componentID, optionID string) error {
	if _, err := s.menuGate.RequireDraftMenu(ctx, tenantID, menuID); err != nil {
		return err
	}
	if _, err := s.repo.GetComponentOnMenu(ctx, menuID, itemID, componentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.CreateNotFoundError("component", componentID, tenantID)
		}
		return apperrors.CreateDatabaseError("Failed to load component", err, tenantID)
	}
	if err := s.repo.SoftDeleteSizeOption(ctx, componentID, optionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.CreateNotFoundError("size_option", optionID, tenantID)
		}
		return apperrors.CreateDatabaseError("Failed to remove size option", err, tenantID)
	}
	return nil
}
