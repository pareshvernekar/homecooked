package foodcategory

import (
	"context"

	cache "github.com/pareshvernekar/homecooked/internal/cache"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	repo "github.com/pareshvernekar/homecooked/internal/repository"
)

// FoodCategoryService handles food category business logic
type FoodCategoryService struct {
	repository     repo.FoodCategoryRepository
	logger          *logger.Logger
	cacheClient    cache.Client
}

// NewFoodCategoryService creates a new food category service instance with dependency injection
func NewFoodCategoryService(repo repo.FoodCategoryRepository, l *logger.Logger, c cache.Client) *FoodCategoryService {
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

	categories, err := s.repository.ListByTenant(tenantID)
	if err != nil {
		s.logger.Error(ctx, "ListCategories: Failed to retrieve food categories from database", "tenant_id", tenantID, "error", err)
		return nil, err
	}

	s.logger.Info(ctx, "ListCategories: Successfully retrieved food categories", "tenant_id", tenantID, "count", len(categories))

	if s.cacheClient != nil {
		cacheKeys := make([]string, 0)
		for _, catItem := range categories {
			cacheKeys = append(cacheKeys, catItem.GetCacheKey())
			}

		s.logger.Debug(ctx, "ListCategories: Updating cache with food categories", "keys", cacheKeys)

			 // Update cache with all categories
		for idx, key := range cacheKeys {
			val := &models.FoodCategory{
				ID:          categories[idx].ID,
				TenantID:    categories[idx].TenantID,
				Name:        categories[idx].Name,
				Description: categories[idx].Description,
				CreatedAt:   categories[idx].CreatedAt,
				UpdatedAt:   categories[idx].UpdatedAt,
			}
			s.cacheClient.Set(ctx, key, val)
		}
	}

	return categories, nil
}

// GetCategoryByID retrieves a specific food category by its ID for the current tenant
func (s *FoodCategoryService) GetCategoryByID(categoryID string, tenantID string) (*models.FoodCategory, error) {
	ctx := context.Background()
	s.logger.Info(ctx, "GetCategoryByID: Fetching food category by ID", "category_id", categoryID, "tenant_id", tenantID)

	category, err := s.repository.ListByTenant(tenantID)
	if err != nil {
		s.logger.Error(ctx, "GetCategoryByID: Failed to retrieve food categories from database", "category_id", categoryID, "error", err)
		return nil, err
	}

	for _, catItem := range category {
		if catItem.ID == categoryID && catItem.TenantID == tenantID {
			s.logger.Info(ctx, "GetCategoryByID: Successfully retrieved food category", "category_id", categoryID)

			if s.cacheClient != nil {
				cacheKey := catItem.GetCacheKey()
				val := &models.FoodCategory{
					ID:          catItem.ID,
					TenantID:    catItem.TenantID,
					Name:        catItem.Name,
					Description: catItem.Description,
					CreatedAt:   catItem.CreatedAt,
					UpdatedAt:   catItem.UpdatedAt,
				}
				s.cacheClient.Set(ctx, cacheKey, val)
			}

			return &catItem, nil
		}
	}

	s.logger.Error(ctx, "GetCategoryByID: Food category not found", "category_id", categoryID)
	return nil, nil // Using nil to indicate not found (caller should check for error condition)
}
