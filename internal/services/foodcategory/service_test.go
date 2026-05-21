package foodcategory

import (
	"context"
	"testing"
	"time"

	cache "github.com/pareshvernekar/homecooked/internal/cache"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/models"
	repo "github.com/pareshvernekar/homecooked/internal/repository"
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

	if service == nil {
		t.Error("Expected NewFoodCategoryService to return a valid service")
	}
}

// TestListCategories_Success tests successful retrieval of all food categories
func TestListCategories_Success(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	mockRepo.categories = []models.FoodCategory{categories[0]}
	l := logger.NewLogger()

	service := NewFoodCategoryService(mockRepo, l, &MockCacheClient{})

	categoriesResult, err := service.ListCategories("tenant_1")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(categoriesResult) != 1 {
		t.Errorf("Expected 1 category, got: %d", len(categoriesResult))
	}

	expectedKeys := []string{"vegetarian"}
	for i, catItem := range categoriesResult {
		if catItem.Name != expectedKeys[i] {
			t.Errorf("Category[%d]: Expected name %s, got: %s", i, expectedKeys[i], catItem.Name)
		}
	}
}

// TestListCategories_EmptyResult tests successful retrieval with empty category list
func TestListCategories_EmptyResult(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	mockRepo.categories = []models.FoodCategory{}
	l := logger.NewLogger()

	service := NewFoodCategoryService(mockRepo, l, &MockCacheClient{})

	categoriesResult, err := service.ListCategories("tenant_1")

	if err != nil {
		t.Errorf("Expected no error with empty result, got: %v", err)
	}

	if len(categoriesResult) != 0 {
		t.Errorf("Expected empty slice, got: %d items", len(categoriesResult))
	}
}

// TestGetCategoryByID_Success tests successful retrieval of a single category by ID
func TestGetCategoryByID_Success(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	mockRepo.categories = []models.FoodCategory{categories[0]}
	l := logger.NewLogger()

	service := NewFoodCategoryService(mockRepo, l, &MockCacheClient{})

	category, err := service.GetCategoryByID("cat1", "tenant_1")

	if err != nil {
		t.Errorf("Expected no error for valid ID, got: %v", err)
	}

	if category == nil {
		t.Error("Expected non-nil category for valid ID")
	}

	if category.Name != "vegetarian" {
		t.Errorf("Expected name 'vegetarian', got: %s", category.Name)
	}
}

// TestGetCategoryByID_NotFound tests retrieval when category doesn't exist
func TestGetCategoryByID_NotFound(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	mockRepo.categories = []models.FoodCategory{categories[0]}
	l := logger.NewLogger()

	service := NewFoodCategoryService(mockRepo, l, &MockCacheClient{})

	category, _ := service.GetCategoryByID("non-existent-id", "tenant_1")

	// Service returns (nil, nil) when category not found - caller should check for this
	if category != nil {
		t.Errorf("Expected nil category for non-existent ID, got: %v", category)
	}
}
