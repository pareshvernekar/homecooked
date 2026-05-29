package foodcategory

import (
	"context"
	"testing"
	"time"

	cache "github.com/pareshvernekar/homecooked/internal/cache"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	repo "github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/stretchr/testify/require"
)

// MockFoodCategoryRepository is a mock implementation for testing
type MockFoodCategoryRepository struct {
	categories []models.FoodCategory
}

func (m *MockFoodCategoryRepository) ListByTenant(tenantID string) ([]models.FoodCategory, error) {
	return m.categories, nil
}

var categories = []models.FoodCategory{
	{ID: "cat1", TenantID: "tenant_1", Name: "vegetarian", Description: func() *string { s := "plant-based"; return &s }(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	{ID: "cat2", TenantID: "tenant_1", Name: "non-vegetarian", Description: func() *string { s := "meat and dairy"; return &s }(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	{ID: "cat3", TenantID: "tenant_1", Name: "vegan", Description: func() *string { s := "no animal products"; return &s }(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
}

// MockCacheClient is a mock implementation of the cache interface for testing
type MockCacheClient struct{}

func (mc *MockCacheClient) Get(ctx context.Context, key string) (any, error) {
	return nil, nil
}

func (mc *MockCacheClient) Set(ctx context.Context, key string, value any, options ...cache.SetOption) error {
	return nil
}

func (mc *MockCacheClient) Has(key string) bool {
	return false
}

func (mc *MockCacheClient) PostInitialize(ctx context.Context, tenantID string, repository repo.FoodCategoryRepository) error {
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
	require.NoError(t, err, "Expected no error for non-existent ID, got: %v", err)
	require.Nil(t, category, "Expected nil category for non-existent ID")
}
