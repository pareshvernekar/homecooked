package foodcategory

import (
	"context"
	"testing"
	"time"

	cache "github.com/pareshvernekar/homecooked/internal/cache"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	"github.com/stretchr/testify/require"
)

// MockFoodCategoryRepository is a mock implementation for testing
type MockFoodCategoryRepository struct {
	categories []models.FoodCategory
}

func (m *MockFoodCategoryRepository) ListByTenant(ctx context.Context, tenantID string) ([]models.FoodCategory, error) {
	return m.categories, nil
}

func (m *MockFoodCategoryRepository) GetByID(ctx context.Context, tenantID string, id string) (*models.FoodCategory, error) {
	for _, cat := range m.categories {
		if cat.ID == id && cat.TenantID == tenantID {
			return &cat, nil
		}
	}
	return nil, nil
}

func (m *MockFoodCategoryRepository) Create(ctx context.Context, category *models.FoodCategory) error {
	m.categories = append(m.categories, *category)
	return nil
}

func (m *MockFoodCategoryRepository) Update(ctx context.Context, category *models.FoodCategory) (int64, error) {
	for i, cat := range m.categories {
		if cat.ID == category.ID && cat.TenantID == category.TenantID {
			m.categories[i] = *category
			return 1, nil
		}
	}
	return 0, nil
}

func (m *MockFoodCategoryRepository) Delete(ctx context.Context, tenantID string, id string) (int64, error) {
	for i, cat := range m.categories {
		if cat.ID == id && cat.TenantID == tenantID {
			m.categories = append(m.categories[:i], m.categories[i+1:]...)
			return 1, nil
		}
	}
	return 0, nil
}

var categories = []models.FoodCategory{
	{ID: "cat1", TenantID: "tenant_1", Name: "vegetarian", Description: "plant-based", CreatedAt: time.Now().UTC().UnixMilli(), UpdatedAt: time.Now().UTC().UnixMilli()},
	{ID: "cat2", TenantID: "tenant_1", Name: "non-vegetarian", Description: "meat and dairy", CreatedAt: time.Now().UTC().UnixMilli(), UpdatedAt: time.Now().UTC().UnixMilli()},
	{ID: "cat3", TenantID: "tenant_1", Name: "vegan", Description: "no animal products", CreatedAt: time.Now().UTC().UnixMilli(), UpdatedAt: time.Now().UTC().UnixMilli()},
}

// MockCacheClient is a mock implementation of the cache interface for testing
type MockCacheClient struct{}

func (mc *MockCacheClient) Get(ctx context.Context, key string) (models.FoodCategory, error) {
	return models.FoodCategory{}, nil
}

func (mc *MockCacheClient) Set(ctx context.Context, key string, value models.FoodCategory, options ...cache.SetOption) error {
	return nil
}

func (mc *MockCacheClient) Has(key string) bool {
	return false
}

func (mc *MockCacheClient) PostInitialize(ctx context.Context, tenantID string, repos map[string]any) error {
	return nil
}

// TestNewFoodCategoryService tests service initialization with dependency injection
func TestNewFoodCategoryService(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := NewFoodCategoryService(mockRepo, l, &MockCacheClient{})
	require.NotNil(t, service, "Expected NewFoodCategoryService to return a valid service")
}

// TestListCategories_Success tests successful retrieval of all food categories
func TestListCategories_Success(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	mockRepo.categories = []models.FoodCategory{categories[0]}
	l := logger.NewLogger()

	service := NewFoodCategoryService(mockRepo, l, &MockCacheClient{})

	categoriesResult, err := service.ListCategories("tenant_1")

	require.NoError(t, err, "Expected no error, got: %v", err)
	require.Len(t, categoriesResult, 1, "Expected 1 category, got: %d", len(categoriesResult))
	require.Equal(t, "vegetarian", categoriesResult[0].Name, "Expected name 'vegetarian', got: %s", categoriesResult[0].Name)
}

// TestListCategories_EmptyResult tests successful retrieval with empty category list
func TestListCategories_EmptyResult(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	mockRepo.categories = []models.FoodCategory{}
	l := logger.NewLogger()

	service := NewFoodCategoryService(mockRepo, l, &MockCacheClient{})

	categoriesResult, err := service.ListCategories("tenant_1")
	require.NoError(t, err, "Expected no error with empty result, got: %v", err)
	require.Len(t, categoriesResult, 0, "Expected empty slice, got: %d items", len(categoriesResult))
}

// TestGetCategoryByID_Success tests successful retrieval of a single category by ID
func TestGetCategoryByID_Success(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	mockRepo.categories = []models.FoodCategory{categories[0]}
	l := logger.NewLogger()

	service := NewFoodCategoryService(mockRepo, l, &MockCacheClient{})

	category, err := service.GetCategoryByID("cat1", "tenant_1")
	require.NoError(t, err, "Expected no error for valid ID, got: %v", err)
	require.NotNil(t, category, "Expected non-nil category for valid ID")
	require.Equal(t, "vegetarian", category.Name, "Expected name 'vegetarian', got: %s", category.Name)
}

// TestGetCategoryByID_NotFound tests retrieval when category doesn't exist
func TestGetCategoryByID_NotFound(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	mockRepo.categories = []models.FoodCategory{categories[0]}
	l := logger.NewLogger()

	service := NewFoodCategoryService(mockRepo, l, &MockCacheClient{})

	category, err := service.GetCategoryByID("non-existent-id", "tenant_1")
	require.Error(t, err, "Expected error for non-existent ID")
	require.Nil(t, category, "Expected nil category for non-existent ID")
}
