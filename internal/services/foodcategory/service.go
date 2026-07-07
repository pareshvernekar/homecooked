package foodcategory

import (
	"context"
	"fmt"
	"strings"

	cache "github.com/pareshvernekar/homecooked/internal/cache"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	models "github.com/pareshvernekar/homecooked/internal/models"
)

type FoodCategoryRepository interface {
	ListByTenant(ctx context.Context, tenantID string) ([]models.FoodCategory, error)
	GetByID(ctx context.Context, tenantID string, id string) (*models.FoodCategory, error)
	Create(ctx context.Context, category *models.FoodCategory) error
	Update(ctx context.Context, category *models.FoodCategory) (int64, error)
	Delete(ctx context.Context, tenantID string, id string) (int64, error)
}

// FoodCategoryService handles food category business logic
type FoodCategoryService struct {
	repository  FoodCategoryRepository
	logger      *logger.Logger
	cacheClient cache.TypedClient[models.FoodCategory]
}

// NewFoodCategoryService creates a new food category service instance with dependency injection
func NewFoodCategoryService(repo FoodCategoryRepository, l *logger.Logger, c cache.TypedClient[models.FoodCategory]) *FoodCategoryService {
	return &FoodCategoryService{
		repository:  repo,
		logger:      l,
		cacheClient: c,
	}
}

// ListCategories retrieves all food categories for a specific tenant
func (s *FoodCategoryService) ListCategories(tenantID string) ([]models.FoodCategory, error) {
	ctx := context.Background()
	s.logger.Info(ctx, "ListCategories: Fetching food categories for tenant", "tenant_id", tenantID)

	categories, err := s.repository.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error(ctx, "ListCategories: Failed to retrieve food categories from database", "tenant_id", tenantID, "error", err)
		return nil, err
	}

	s.logger.Info(ctx, "ListCategories: Successfully retrieved food categories", "tenant_id", tenantID, "count", len(categories))

	if s.cacheClient != nil && len(categories) > 0 {
		s.logger.Debug(ctx, "ListCategories: Updating cache with food categories", "keys", "all categories")

		for _, catItem := range categories {
			cacheErr := s.cacheClient.Set(ctx, catItem.GetCacheKey(), catItem)
			if cacheErr != nil {
				s.logger.Error(ctx, "ListCategories: Failed to cache food category", "key", catItem.GetCacheKey(), "error", cacheErr)
				return nil, cacheErr
			}
		}
	}

	return categories, nil
}

// GetCategoryByID retrieves a specific food category by its ID for the current tenant
func (s *FoodCategoryService) GetCategoryByID(categoryID string, tenantID string) (*models.FoodCategory, error) {
	ctx := context.Background()
	s.logger.Info(ctx, "GetCategoryByID: Fetching food category by ID", "category_id", categoryID, "tenant_id", tenantID)

	categories, err := s.repository.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error(ctx, "GetCategoryByID: Failed to retrieve food categories from database", "category_id", categoryID, "error", err)
		return nil, err
	}

	for _, catItem := range categories {
		if catItem.ID == categoryID && catItem.TenantID == tenantID {
			s.logger.Info(ctx, "GetCategoryByID: Successfully retrieved food category", "category_id", categoryID)

			if s.cacheClient != nil {
				cacheKey := catItem.GetCacheKey()
				cacheErr := s.cacheClient.Set(ctx, cacheKey, catItem)
				if cacheErr != nil {
					s.logger.Error(ctx, "GetCategoryByID: Failed to cache food category", "key", cacheKey, "error", cacheErr)
					return nil, cacheErr
				}
			}

			return &catItem, nil
		}
	}

	s.logger.Error(ctx, "GetCategoryByID: Food category not found", "category_id", categoryID)
	return nil, fmt.Errorf("food category not found with id %s", categoryID)
}

// GetCategoryByName retrieves a food category by its name (case-insensitive) and returns its ID for database persistence.
// This method resolves user-friendly category names to database-required UUIDs.
func (s *FoodCategoryService) GetCategoryByName(categoryName, tenantID string) (string, error) {
	ctx := context.Background()
	s.logger.Info(ctx, "GetCategoryByName: Resolving category name to ID", "category_name", categoryName, "tenant_id", tenantID)

	// Fetch all categories for the tenant
	categories, err := s.repository.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error(ctx, "GetCategoryByName: Failed to retrieve food categories", "error", err)
		return "", fmt.Errorf("failed to fetch categories for tenant %s: %w", tenantID, err)
	}

	// Search for matching category (case-insensitive comparison)
	for _, cat := range categories {
		if strings.EqualFold(cat.Name, categoryName) {
			s.logger.Info(ctx, "GetCategoryByName: Category name resolved to ID",
				"category_name", cat.Name,
				"category_id", cat.ID)

			// Cache the category if cache client is available
			if s.cacheClient != nil {
				cacheErr := s.cacheClient.Set(ctx, cat.GetCacheKey(), cat)
				if cacheErr != nil {
					s.logger.Error(ctx, "GetCategoryByName: Failed to cache food category after resolution",
						"key", cat.GetCacheKey(), "error", cacheErr)
					// Don't fail - caching should not block the operation
				}
			}

			return cat.ID, nil
		}
	}
	s.logger.Error(ctx, "GetCategoryByName: Category name not found in tenant's catalog",
		"category_name", categoryName, "tenant_id", tenantID)
	return "", fmt.Errorf("category '%s' not found in tenant %s's catalog", categoryName, tenantID)
}
