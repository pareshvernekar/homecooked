package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// PostgreSQLMenuItemRepository manages menu items, components, and size options.
// REQITEM001–REQITEM006
type PostgreSQLMenuItemRepository struct {
	DB       *sqlx.DB
	TenantID string
	Logger   *logger.Logger
}

// NewMenuItemRepository constructs a menu-item repository.
func NewMenuItemRepository(db *sqlx.DB, l *logger.Logger, tenantID string) *PostgreSQLMenuItemRepository {
	return &PostgreSQLMenuItemRepository{DB: db, TenantID: tenantID, Logger: l}
}

// GetActiveDraftMenu returns the menu if it is active and draft.
func (r *PostgreSQLMenuItemRepository) GetActiveDraftMenu(ctx context.Context, menuID string) (*models.Menu, error) {
	var m models.Menu
	err := r.DB.GetContext(ctx, &m, `
		SELECT `+menuSelectCols+`
		FROM menu WHERE tenant_id = $1 AND id = $2 AND is_active = TRUE`, r.TenantID, menuID)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// CategoryBelongsToMenu checks an active category on the menu.
func (r *PostgreSQLMenuItemRepository) CategoryBelongsToMenu(ctx context.Context, menuID, categoryID string) (bool, error) {
	var ok bool
	err := r.DB.GetContext(ctx, &ok, `
		SELECT EXISTS(
			SELECT 1 FROM menu_category
			WHERE tenant_id = $1 AND id = $2 AND menu_id = $3 AND is_active = TRUE
		)`, r.TenantID, categoryID, menuID)
	return ok, err
}

// FoodItemActiveInTenant checks food item usability.
func (r *PostgreSQLMenuItemRepository) FoodItemActiveInTenant(ctx context.Context, foodItemID string) (bool, error) {
	var ok bool
	err := r.DB.GetContext(ctx, &ok, `
		SELECT EXISTS(
			SELECT 1 FROM food_item
			WHERE tenant_id = $1 AND id = $2 AND is_active = TRUE
		)`, r.TenantID, foodItemID)
	return ok, err
}

// CreateItemWithComponents inserts a menu-item with nested components and size options in one transaction.
// REQITEM001, REQITEM003
func (r *PostgreSQLMenuItemRepository) CreateItemWithComponents(
	ctx context.Context,
	item *models.MenuItem,
	components []models.MenuItemComponentInput,
) error {
	now := time.Now().UTC().UnixMilli()
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	item.TenantID = r.TenantID
	item.IsActive = true
	item.CreatedAt = now
	item.UpdatedAt = now

	tx, err := r.DB.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SET CONSTRAINTS menu_item_component_default_size_fk DEFERRED`); err != nil {
		// constraint may not be deferrable on some envs; continue — we set default after options
		r.Logger.Debug(ctx, "defer constraints skipped", "error", err.Error())
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO menu_item (id, tenant_id, menu_id, category_id, kind, name, description, sequence, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,TRUE,$9,$9)`,
		item.ID, r.TenantID, item.MenuID, item.CategoryID, item.Kind, item.Name, nullStr(item.Description), item.Sequence, now,
	); err != nil {
		return fmt.Errorf("insert menu_item: %w", err)
	}

	for _, cin := range components {
		compID := uuid.New().String()
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO menu_item_component (id, tenant_id, menu_item_id, food_item_id, is_active, created_at, updated_at)
			VALUES ($1,$2,$3,$4,TRUE,$5,$5)`,
			compID, r.TenantID, item.ID, cin.FoodItemID, now,
		); err != nil {
			return fmt.Errorf("insert component: %w", err)
		}

		optIDs := make([]string, 0, len(cin.SizeOptions))
		for _, so := range cin.SizeOptions {
			optID := uuid.New().String()
			if so.ID != "" {
				optID = so.ID
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO menu_item_component_size_option
				(id, tenant_id, component_id, size_unit_id, qty, price, is_active, created_at, updated_at)
				VALUES ($1,$2,$3,$4,$5,$6,TRUE,$7,$7)`,
				optID, r.TenantID, compID, so.SizeUnitID, so.Qty, so.Price, now,
			); err != nil {
				return fmt.Errorf("insert size option: %w", err)
			}
			optIDs = append(optIDs, optID)
		}

		defaultID, err := resolveDefaultOptionID(cin, optIDs)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE menu_item_component SET default_size_option_id = $1, updated_at = $2
			WHERE tenant_id = $3 AND id = $4`, defaultID, now, r.TenantID, compID); err != nil {
			return fmt.Errorf("set default size: %w", err)
		}
	}

	return tx.Commit()
}

func resolveDefaultOptionID(cin models.MenuItemComponentInput, optIDs []string) (string, error) {
	if len(optIDs) == 0 {
		return "", fmt.Errorf("component requires at least one size option")
	}
	if cin.DefaultSizeOptionID != nil && *cin.DefaultSizeOptionID != "" {
		for _, id := range optIDs {
			if id == *cin.DefaultSizeOptionID {
				return id, nil
			}
		}
		return "", fmt.Errorf("default_size_option_id not among component size options")
	}
	if cin.DefaultSizeIndex != nil {
		i := *cin.DefaultSizeIndex
		if i < 0 || i >= len(optIDs) {
			return "", fmt.Errorf("default_size_index out of range")
		}
		return optIDs[i], nil
	}
	// default to first size option
	return optIDs[0], nil
}

// SoftDeleteItem deactivates a menu-item.
// REQITEM002
func (r *PostgreSQLMenuItemRepository) SoftDeleteItem(ctx context.Context, menuID, itemID string) error {
	now := time.Now().UTC().UnixMilli()
	result, err := r.DB.ExecContext(ctx, `
		UPDATE menu_item SET is_active = FALSE, updated_at = $1
		WHERE tenant_id = $2 AND menu_id = $3 AND id = $4 AND is_active = TRUE`,
		now, r.TenantID, menuID, itemID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetActiveItem returns an active item on a menu.
func (r *PostgreSQLMenuItemRepository) GetActiveItem(ctx context.Context, menuID, itemID string) (*models.MenuItem, error) {
	var item models.MenuItem
	err := r.DB.GetContext(ctx, &item, `
		SELECT id, tenant_id, menu_id, category_id, kind, name, COALESCE(description,'') as description,
		       sequence, is_active, created_at, updated_at
		FROM menu_item
		WHERE tenant_id = $1 AND menu_id = $2 AND id = $3 AND is_active = TRUE`,
		r.TenantID, menuID, itemID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// AddSizeOption inserts a size option; optionally sets as default.
// REQITEM003
func (r *PostgreSQLMenuItemRepository) AddSizeOption(
	ctx context.Context,
	componentID string,
	in models.SizeOptionCreateRequest,
) (*models.MenuItemComponentSizeOption, error) {
	now := time.Now().UTC().UnixMilli()
	id := uuid.New().String()
	if _, err := r.DB.ExecContext(ctx, `
		INSERT INTO menu_item_component_size_option
		(id, tenant_id, component_id, size_unit_id, qty, price, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,TRUE,$7,$7)`,
		id, r.TenantID, componentID, in.SizeUnitID, in.Qty, in.Price, now,
	); err != nil {
		return nil, err
	}
	if in.IsDefault {
		if _, err := r.DB.ExecContext(ctx, `
			UPDATE menu_item_component SET default_size_option_id = $1, updated_at = $2
			WHERE tenant_id = $3 AND id = $4`, id, now, r.TenantID, componentID); err != nil {
			return nil, err
		}
	}
	return &models.MenuItemComponentSizeOption{
		ID: id, TenantID: r.TenantID, ComponentID: componentID,
		SizeUnitID: in.SizeUnitID, Qty: in.Qty, Price: in.Price, IsActive: true,
		IsDefault: in.IsDefault, CreatedAt: now, UpdatedAt: now,
	}, nil
}

// UpdateSizeOption updates fields on a size option.
func (r *PostgreSQLMenuItemRepository) UpdateSizeOption(
	ctx context.Context,
	componentID, optionID string,
	in models.SizeOptionUpdateRequest,
) error {
	var cur models.MenuItemComponentSizeOption
	if err := r.DB.GetContext(ctx, &cur, `
		SELECT id, tenant_id, component_id, size_unit_id, qty, price, is_active, created_at, updated_at
		FROM menu_item_component_size_option
		WHERE tenant_id = $1 AND component_id = $2 AND id = $3 AND is_active = TRUE`,
		r.TenantID, componentID, optionID); err != nil {
		return err
	}
	if in.SizeUnitID != nil {
		cur.SizeUnitID = *in.SizeUnitID
	}
	if in.Qty != nil {
		cur.Qty = *in.Qty
	}
	if in.Price != nil {
		cur.Price = *in.Price
	}
	now := time.Now().UTC().UnixMilli()
	if _, err := r.DB.ExecContext(ctx, `
		UPDATE menu_item_component_size_option
		SET size_unit_id = $1, qty = $2, price = $3, updated_at = $4
		WHERE tenant_id = $5 AND id = $6`,
		cur.SizeUnitID, cur.Qty, cur.Price, now, r.TenantID, optionID); err != nil {
		return err
	}
	if in.IsDefault != nil && *in.IsDefault {
		if _, err := r.DB.ExecContext(ctx, `
			UPDATE menu_item_component SET default_size_option_id = $1, updated_at = $2
			WHERE tenant_id = $3 AND id = $4`, optionID, now, r.TenantID, componentID); err != nil {
			return err
		}
	}
	return nil
}

// SoftDeleteSizeOption deactivates a size option.
func (r *PostgreSQLMenuItemRepository) SoftDeleteSizeOption(ctx context.Context, componentID, optionID string) error {
	now := time.Now().UTC().UnixMilli()
	result, err := r.DB.ExecContext(ctx, `
		UPDATE menu_item_component_size_option SET is_active = FALSE, updated_at = $1
		WHERE tenant_id = $2 AND component_id = $3 AND id = $4 AND is_active = TRUE`,
		now, r.TenantID, componentID, optionID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetComponentOnMenu verifies component belongs to an item on the menu.
func (r *PostgreSQLMenuItemRepository) GetComponentOnMenu(ctx context.Context, menuID, itemID, componentID string) (*models.MenuItemComponent, error) {
	var c models.MenuItemComponent
	err := r.DB.GetContext(ctx, &c, `
		SELECT c.id, c.tenant_id, c.menu_item_id, c.food_item_id, c.default_size_option_id, c.is_active, c.created_at, c.updated_at
		FROM menu_item_component c
		JOIN menu_item i ON i.tenant_id = c.tenant_id AND i.id = c.menu_item_id
		WHERE c.tenant_id = $1 AND c.id = $2 AND c.menu_item_id = $3 AND i.menu_id = $4
		  AND c.is_active = TRUE AND i.is_active = TRUE`,
		r.TenantID, componentID, itemID, menuID)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
