package foodcategory

import (
	"context"
	"fmt"
	"strings"

	logger "github.com/pareshvernekar/homecooked/internal/logger"
	models "github.com/pareshvernekar/homecooked/internal/models"
)

type FoodCategoryRepository interface {
	ListByTenant(ctx context.Context, tenantID string) ([]*models.FoodCategory, error)
	GetByID(ctx context.Context, tenantID string, id string) (*models.FoodCategory, error)
	Create(ctx context.Context, category *models.FoodCategory) error
	Update(ctx context.Context, category *models.FoodCategory) (int64, error)
	Delete(ctx context.Context, tenantID string, id string) (int64, error)
}

// FoodCategoryService handles food category business logic
type FoodCategoryService struct {
	repository FoodCategoryRepository
	logger     *logger.Logger
}

// NewFoodCategoryService creates a new food category service instance with dependency injection
func NewFoodCategoryService(repo FoodCategoryRepository, l *logger.Logger) *FoodCategoryService {
	return &FoodCategoryService{
		repository: repo,
		logger:     l,
	}
}

// ListCategories retrieves all food categories for a specific tenant
func (s *FoodCategoryService) ListCategories(ctx context.Context, tenantID string) ([]*models.FoodCategory, error) {
	s.logger.Info(ctx, "ListCategories: Fetching food categories for tenant", "tenant_id", tenantID)

	categories, err := s.repository.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error(ctx, "ListCategories: Failed to retrieve food categories from database", "tenant_id", tenantID, "error", err)
		return nil, err
	}

	s.logger.Info(ctx, "ListCategories: Successfully retrieved food categories", "tenant_id", tenantID, "count", len(categories))
	return categories, nil
}

// GetCategoryByID retrieves a specific food category by its ID for the current tenant
func (s *FoodCategoryService) GetCategoryByID(ctx context.Context, categoryID string, tenantID string) (*models.FoodCategory, error) {
	s.logger.Info(ctx, "GetCategoryByID: Fetching food category by ID", "category_id", categoryID, "tenant_id", tenantID)

	categories, err := s.repository.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error(ctx, "GetCategoryByID: Failed to retrieve food categories from database", "category_id", categoryID, "error", err)
		return nil, err
	}

	for _, catItem := range categories {
		if catItem.ID == categoryID && catItem.TenantID == tenantID {
			s.logger.Info(ctx, "GetCategoryByID: Successfully retrieved food category", "category_id", categoryID)
			return catItem, nil
		}
	}

	s.logger.Error(ctx, "GetCategoryByID: Food category not found", "category_id", categoryID)
	return nil, fmt.Errorf("food category not found with id %s", categoryID)
}

// GetCategoryByName retrieves a food category by its name (case-insensitive) and returns its ID for database persistence.
// This method resolves user-friendly category names to database-required UUIDs.
func (s *FoodCategoryService) GetCategoryByName(ctx context.Context, categoryName, tenantID string) (*models.FoodCategory, error) {
	s.logger.Info(ctx, "GetCategoryByName: Resolving category name to ID", "category_name", categoryName, "tenant_id", tenantID)

	// Fetch all categories for the tenant
	categories, err := s.repository.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error(ctx, "GetCategoryByName: Failed to retrieve food categories", "error", err)
		return nil, fmt.Errorf("failed to fetch categories for tenant %s: %w", tenantID, err)
	}

	// Search for matching category (case-insensitive comparison)
	for _, cat := range categories {
		if strings.EqualFold(cat.Name, categoryName) {
			s.logger.Info(ctx, "GetCategoryByName: Category name resolved to ID",
				"category_name", cat.Name,
				"category_id", cat.ID)
			return cat, nil
		}
	}

	s.logger.Error(ctx, "GetCategoryByName: Category name not found in tenant's catalog",
		"category_name", categoryName, "tenant_id", tenantID)
	return nil, fmt.Errorf("category '%s' not found in tenant %s's catalog", categoryName, tenantID)
}

// CreateCategory creates a new food category for the current tenant
func (s *FoodCategoryService) CreateCategory(ctx context.Context, category *models.FoodCategory) error {
	s.logger.Info(ctx, "CreateCategory: Creating new food category", "category_id", category.ID, "name", category.Name)

	err := s.repository.Create(ctx, category)
	if err != nil {
		s.logger.Error(ctx, "CreateCategory: Failed to create food category in database", "category_id", category.ID, "error", err)
		return err
	}

	s.logger.Info(ctx, "CreateCategory: Successfully created food category", "category_id", category.ID)
	return nil
}

// UpdateCategory updates an existing food category by its ID
func (s *FoodCategoryService) UpdateCategory(ctx context.Context, category *models.FoodCategory) (int64, error) {
	s.logger.Info(ctx, "UpdateCategory: Updating food category", "category_id", category.ID)

	rowsAffected, err := s.repository.Update(ctx, category)
	if err != nil {
		s.logger.Error(ctx, "UpdateCategory: Failed to update food category", "category_id", category.ID, "error", err)
		return 0, err
	}

	if rowsAffected == 0 {
		s.logger.Error(ctx, "UpdateCategory: Category not found for update", "category_id", category.ID)
		return 0, fmt.Errorf("food category not found with id %s", category.ID)
	}

	s.logger.Info(ctx, "UpdateCategory: Successfully updated food category", "category_id", category.ID, "rows_updated", rowsAffected)
	return rowsAffected, nil
}

// DeleteCategory deletes a food category by its ID (soft delete - sets is_active = false)
func (s *FoodCategoryService) DeleteCategory(ctx context.Context, tenantID string, id string) (int64, error) {
	s.logger.Info(ctx, "DeleteCategory: Deleting food category", "tenant_id", tenantID, "id", id)

	rowsAffected, err := s.repository.Delete(ctx, tenantID, id)
	if err != nil {
		s.logger.Error(ctx, "DeleteCategory: Failed to delete food category", "tenant_id", tenantID, "category_id", id, "error", err)
		return 0, err
	}

	if rowsAffected == 0 {
		s.logger.Error(ctx, "DeleteCategory: Category not found for deletion", "tenant_id", tenantID, "category_id", id)
		return 0, fmt.Errorf("food category not found with id %s", id)
	}

	s.logger.Info(ctx, "DeleteCategory: Successfully deleted food category", "tenant_id", tenantID, "id", id, "rows_deleted", rowsAffected)
	return rowsAffected, nil
}
