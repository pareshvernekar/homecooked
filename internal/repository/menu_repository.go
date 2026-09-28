package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// PostgreSQLMenuRepository implements menu persistence.
// REQMENU001–REQMENU004, REQMENU007–REQMENU009
type PostgreSQLMenuRepository struct {
	DB       *sqlx.DB
	TenantID string
	Logger   *logger.Logger
}

// NewMenuRepository constructs a tenant-scoped menu repository.
func NewMenuRepository(db *sqlx.DB, l *logger.Logger, tenantID string) *PostgreSQLMenuRepository {
	return &PostgreSQLMenuRepository{DB: db, TenantID: tenantID, Logger: l}
}

const menuSelectCols = `
	id, tenant_id, name, COALESCE(description, '') as description, menu_type, status, is_active,
	menu_date::text as menu_date, start_date::text as start_date, end_date::text as end_date,
	event_date::text as event_date, event_location, created_at, updated_at`

// Create inserts a draft menu and optional categories.
// REQMENU001, REQMENU005
func (r *PostgreSQLMenuRepository) Create(ctx context.Context, m *models.Menu, categories []models.MenuCategoryInput) error {
	if err := requireTenantExists(ctx, r.DB, r.TenantID); err != nil {
		return err
	}
	now := time.Now().UTC().UnixMilli()
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	m.TenantID = r.TenantID
	m.Status = models.MenuStatusDraft
	m.IsActive = true
	m.CreatedAt = now
	m.UpdatedAt = now

	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO menu (
			id, tenant_id, name, description, menu_type, status, is_active,
			menu_date, start_date, end_date, event_date, event_location, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,
			NULLIF($8,'')::date, NULLIF($9,'')::date, NULLIF($10,'')::date, NULLIF($11,'')::date, $12, $13, $14
		)`,
		m.ID, r.TenantID, m.Name, nullStr(m.Description), m.MenuType, m.Status, m.IsActive,
		deref(m.MenuDate), deref(m.StartDate), deref(m.EndDate), deref(m.EventDate), nullStr(deref(m.EventLocation)),
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert menu: %w", err)
	}

	for _, c := range categories {
		catID := c.ID
		if catID == "" {
			catID = uuid.New().String()
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO menu_category (id, tenant_id, menu_id, name, sequence, is_active, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,TRUE,$6,$6)`,
			catID, r.TenantID, m.ID, c.Name, c.Sequence, now,
		)
		if err != nil {
			return fmt.Errorf("insert menu_category: %w", err)
		}
	}

	return tx.Commit()
}

// List returns active menus filtered by status (published default, draft, or all).
// REQMENU002
func (r *PostgreSQLMenuRepository) List(ctx context.Context, statusFilter string) ([]*models.Menu, error) {
	query := `SELECT ` + menuSelectCols + ` FROM menu WHERE tenant_id = $1 AND is_active = TRUE`
	args := []interface{}{r.TenantID}
	switch strings.ToLower(statusFilter) {
	case "", models.MenuStatusPublished:
		query += ` AND status = $2`
		args = append(args, models.MenuStatusPublished)
	case models.MenuStatusDraft:
		query += ` AND status = $2`
		args = append(args, models.MenuStatusDraft)
	case "all":
		// no status filter
	default:
		query += ` AND status = $2`
		args = append(args, statusFilter)
	}
	query += ` ORDER BY created_at DESC`

	var menus []*models.Menu
	if err := r.DB.SelectContext(ctx, &menus, query, args...); err != nil {
		return nil, err
	}
	return menus, nil
}

// GetActiveByID returns an active menu (draft or published) for the tenant.
func (r *PostgreSQLMenuRepository) GetActiveByID(ctx context.Context, id string) (*models.Menu, error) {
	var m models.Menu
	err := r.DB.GetContext(ctx, &m, `
		SELECT `+menuSelectCols+` FROM menu
		WHERE tenant_id = $1 AND id = $2 AND is_active = TRUE`, r.TenantID, id)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetTree loads menu with categories, items, components, size options, and live-joined food display.
// REQMENU003
func (r *PostgreSQLMenuRepository) GetTree(ctx context.Context, id string) (*models.MenuTree, error) {
	m, err := r.GetActiveByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var categories []models.MenuCategory
	if err := r.DB.SelectContext(ctx, &categories, `
		SELECT id, tenant_id, menu_id, name, sequence, is_active, created_at, updated_at
		FROM menu_category
		WHERE tenant_id = $1 AND menu_id = $2 AND is_active = TRUE
		ORDER BY sequence ASC, created_at ASC`, r.TenantID, id); err != nil {
		return nil, err
	}

	tree := &models.MenuTree{Menu: *m, Categories: make([]models.MenuCategoryTree, 0, len(categories))}
	for _, cat := range categories {
		items, err := r.loadItemsForCategory(ctx, id, cat.ID)
		if err != nil {
			return nil, err
		}
		tree.Categories = append(tree.Categories, models.MenuCategoryTree{
			MenuCategory: cat,
			Items:        items,
		})
	}
	return tree, nil
}

func (r *PostgreSQLMenuRepository) loadItemsForCategory(ctx context.Context, menuID, categoryID string) ([]models.MenuItem, error) {
	var items []models.MenuItem
	if err := r.DB.SelectContext(ctx, &items, `
		SELECT id, tenant_id, menu_id, category_id, kind, name, COALESCE(description,'') as description,
		       sequence, is_active, created_at, updated_at
		FROM menu_item
		WHERE tenant_id = $1 AND menu_id = $2 AND category_id = $3 AND is_active = TRUE
		ORDER BY sequence ASC, created_at ASC`, r.TenantID, menuID, categoryID); err != nil {
		return nil, err
	}

	for i := range items {
		comps, err := r.loadComponents(ctx, items[i].ID)
		if err != nil {
			return nil, err
		}
		items[i].Components = comps
		total := sumDefaultTotal(comps)
		items[i].DefaultTotal = &total
	}
	return items, nil
}

func (r *PostgreSQLMenuRepository) loadComponents(ctx context.Context, menuItemID string) ([]models.MenuItemComponent, error) {
	var comps []models.MenuItemComponent
	if err := r.DB.SelectContext(ctx, &comps, `
		SELECT id, tenant_id, menu_item_id, food_item_id, default_size_option_id, is_active, created_at, updated_at
		FROM menu_item_component
		WHERE tenant_id = $1 AND menu_item_id = $2 AND is_active = TRUE
		ORDER BY created_at ASC`, r.TenantID, menuItemID); err != nil {
		return nil, err
	}

	for i := range comps {
		opts, err := r.loadSizeOptions(ctx, comps[i].ID)
		if err != nil {
			return nil, err
		}
		if comps[i].DefaultSizeOptionID != nil {
			for j := range opts {
				if opts[j].ID == *comps[i].DefaultSizeOptionID {
					opts[j].IsDefault = true
				}
			}
		}
		comps[i].SizeOptions = opts

		var fi models.FoodItemDisplay
		err = r.DB.GetContext(ctx, &fi, `
			SELECT id, name, COALESCE(is_vegetarian,false) as is_vegetarian,
			       COALESCE(avoidance,'') as avoidance, COALESCE(image_url,'') as image_url
			FROM food_item
			WHERE tenant_id = $1 AND id = $2`, r.TenantID, comps[i].FoodItemID)
		if err == nil {
			comps[i].FoodItem = &fi
		}
	}
	return comps, nil
}

func (r *PostgreSQLMenuRepository) loadSizeOptions(ctx context.Context, componentID string) ([]models.MenuItemComponentSizeOption, error) {
	var opts []models.MenuItemComponentSizeOption
	err := r.DB.SelectContext(ctx, &opts, `
		SELECT id, tenant_id, component_id, size_unit_id, qty, price, is_active, created_at, updated_at
		FROM menu_item_component_size_option
		WHERE tenant_id = $1 AND component_id = $2 AND is_active = TRUE
		ORDER BY created_at ASC`, r.TenantID, componentID)
	return opts, err
}

func sumDefaultTotal(comps []models.MenuItemComponent) float64 {
	var total float64
	for _, c := range comps {
		defID := ""
		if c.DefaultSizeOptionID != nil {
			defID = *c.DefaultSizeOptionID
		}
		for _, o := range c.SizeOptions {
			if o.ID == defID || (defID == "" && o.IsDefault) {
				total += o.Price
				break
			}
		}
	}
	return total
}

// UpdateMetadata updates draft menu fields.
// REQMENU004
func (r *PostgreSQLMenuRepository) UpdateMetadata(ctx context.Context, m *models.Menu) error {
	now := time.Now().UTC().UnixMilli()
	result, err := r.DB.ExecContext(ctx, `
		UPDATE menu SET
			name = $1, description = $2,
			menu_date = NULLIF($3,'')::date, start_date = NULLIF($4,'')::date,
			end_date = NULLIF($5,'')::date, event_date = NULLIF($6,'')::date,
			event_location = $7, updated_at = $8
		WHERE tenant_id = $9 AND id = $10 AND is_active = TRUE AND status = $11`,
		m.Name, nullStr(m.Description),
		deref(m.MenuDate), deref(m.StartDate), deref(m.EndDate), deref(m.EventDate),
		nullStr(deref(m.EventLocation)), now, r.TenantID, m.ID, models.MenuStatusDraft,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	m.UpdatedAt = now
	return nil
}

// ReplaceCategories soft-deactivates existing categories and inserts the provided set (draft only).
func (r *PostgreSQLMenuRepository) ReplaceCategories(ctx context.Context, menuID string, categories []models.MenuCategoryInput) error {
	now := time.Now().UTC().UnixMilli()
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		UPDATE menu_category SET is_active = FALSE, updated_at = $1
		WHERE tenant_id = $2 AND menu_id = $3 AND is_active = TRUE`, now, r.TenantID, menuID); err != nil {
		return err
	}
	for _, c := range categories {
		catID := c.ID
		if catID == "" {
			catID = uuid.New().String()
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO menu_category (id, tenant_id, menu_id, name, sequence, is_active, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,TRUE,$6,$6)`,
			catID, r.TenantID, menuID, c.Name, c.Sequence, now,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Deactivate soft-deletes a menu.
// REQMENU004
func (r *PostgreSQLMenuRepository) Deactivate(ctx context.Context, id string) error {
	now := time.Now().UTC().UnixMilli()
	result, err := r.DB.ExecContext(ctx, `
		UPDATE menu SET is_active = FALSE, updated_at = $1
		WHERE tenant_id = $2 AND id = $3 AND is_active = TRUE`, now, r.TenantID, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// HasPublishedWeeklyStartDate reports whether another active published weekly uses startDate.
// REQMENU009
func (r *PostgreSQLMenuRepository) HasPublishedWeeklyStartDate(ctx context.Context, startDate string, excludeMenuID string) (bool, error) {
	var exists bool
	err := r.DB.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM menu
			WHERE tenant_id = $1
			  AND menu_type = 'weekly'
			  AND status = 'published'
			  AND is_active = TRUE
			  AND start_date = $2::date
			  AND id <> $3
		)`, r.TenantID, startDate, excludeMenuID)
	return exists, err
}

// SetStatus updates menu status (publish/unpublish).
// REQMENU007, REQMENU008
func (r *PostgreSQLMenuRepository) SetStatus(ctx context.Context, id, status string) error {
	now := time.Now().UTC().UnixMilli()
	result, err := r.DB.ExecContext(ctx, `
		UPDATE menu SET status = $1, updated_at = $2
		WHERE tenant_id = $3 AND id = $4 AND is_active = TRUE`, status, now, r.TenantID, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CountActiveCategories returns active category count for a menu.
func (r *PostgreSQLMenuRepository) CountActiveCategories(ctx context.Context, menuID string) (int, error) {
	var n int
	err := r.DB.GetContext(ctx, &n, `
		SELECT COUNT(*) FROM menu_category
		WHERE tenant_id = $1 AND menu_id = $2 AND is_active = TRUE`, r.TenantID, menuID)
	return n, err
}

// CountActiveItems returns active menu-item count for a menu.
func (r *PostgreSQLMenuRepository) CountActiveItems(ctx context.Context, menuID string) (int, error) {
	var n int
	err := r.DB.GetContext(ctx, &n, `
		SELECT COUNT(*) FROM menu_item
		WHERE tenant_id = $1 AND menu_id = $2 AND is_active = TRUE`, r.TenantID, menuID)
	return n, err
}

// ListActiveItemsForPublish returns items with components/options for structural validation.
func (r *PostgreSQLMenuRepository) ListActiveItemsForPublish(ctx context.Context, menuID string) ([]models.MenuItem, error) {
	var items []models.MenuItem
	if err := r.DB.SelectContext(ctx, &items, `
		SELECT id, tenant_id, menu_id, category_id, kind, name, COALESCE(description,'') as description,
		       sequence, is_active, created_at, updated_at
		FROM menu_item
		WHERE tenant_id = $1 AND menu_id = $2 AND is_active = TRUE`, r.TenantID, menuID); err != nil {
		return nil, err
	}
	for i := range items {
		comps, err := r.loadComponents(ctx, items[i].ID)
		if err != nil {
			return nil, err
		}
		items[i].Components = comps
	}
	return items, nil
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
