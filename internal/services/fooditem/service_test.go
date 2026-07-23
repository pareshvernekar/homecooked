package fooditem

import (
	"context"
	"errors"
	"testing"
	"time"

	cache "github.com/pareshvernekar/homecooked/internal/cache"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	models "github.com/pareshvernekar/homecooked/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// Mock Definitions for Repository and Cache
// =============================================================================

// MockFoodItemRepository is a mock implementation of the FoodItemRepository interface
type MockFoodItemRepository struct {
	mock.Mock
}

func (m *MockFoodItemRepository) Create(ctx context.Context, item *models.FoodItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockFoodItemRepository) GetByID(ctx context.Context, tenantID string, id string) (*models.FoodItem, error) {
	args := m.Called(ctx, tenantID, id)
	// Handle nil case before type assertion
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FoodItem), args.Error(1)
}

func (m *MockFoodItemRepository) Update(ctx context.Context, item *models.FoodItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockFoodItemRepository) Delete(ctx context.Context, id string) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFoodItemRepository) ListByTenant(ctx context.Context, tenantID string, offset, limit int) ([]*models.FoodItem, int64, error) {
	args := m.Called(ctx, tenantID, offset, limit)
	return args.Get(0).([]*models.FoodItem), args.Get(1).(int64), args.Error(2)
}

// MockCacheClient mocks the TypedClient[models.FoodItem] interface
type MockCacheClient struct {
	mock.Mock
	CacheHits   []string
	CacheMisses []string
	CacheErrors []error
	CacheSets   int
	CacheKeys   map[string]string
	GetKey      string
}

func (m *MockCacheClient) Get(ctx context.Context, key string) (*models.FoodItem, error) {
	args := m.Called(key)
	if m.CacheHits != nil {
		m.CacheHits = append(m.CacheHits, key)
	}

	var hit *models.FoodItem
	currTime := time.Now().UTC().UnixMilli()
	switch key {
	case "food_item:valid123":
		hit = &models.FoodItem{
			ID:                 "valid123",
			TenantID:           "tenant-1",
			Name:               "Test Item",
			Description:        "Test Description",
			CategoryID:         "vegetarian",
			Price:              10.99,
			IsVegetarian:       true,
			AvailabilityStatus: "available",
			CreatedAt:          currTime,
			UpdatedAt:          currTime,
		}
	case "food_item:item-123":
		hit = &models.FoodItem{
			ID:                 "item-123",
			TenantID:           "tenant-1",
			Name:               "Delicious Pasta Bolognese",
			Description:        "Hearty Italian beef and tomato pasta dish",
			CategoryID:         "non_vegetarian",
			Price:              14.50,
			IsVegetarian:       false,
			AvailabilityStatus: "available",
			CreatedAt:          currTime,
			UpdatedAt:          currTime,
		}
	default:
		m.CacheMisses = append(m.CacheMisses, key)
		// Return nil on cache miss (non-fatal - fall through to DB), not an error
		return nil, nil
	}

	args = m.Called(key) // Re-call to get the item
	return hit, args.Error(1)
}

func (m *MockCacheClient) Set(ctx context.Context, key string, value *models.FoodItem, options ...cache.SetOption) error {
	if m.CacheSets > 0 {
		m.CacheKeys[key] = value.ID
	}
	params := []interface{}{key, value}
	for _, opt := range options {
		params = append(params, opt)
	}
	args := m.Called(params...)
	return args.Error(0)
}

func (m *MockCacheClient) Has(key string) bool {
	args := m.Called(key)
	if m.CacheHits != nil {
		m.CacheHits = append(m.CacheHits, key)
	}
	return !args.Bool(0)
}

func (m *MockCacheClient) PostInitialize(ctx context.Context, tenantID string, repos map[string]any) error {
	args := m.Called(tenantID, repos)
	return args.Error(0)
}

type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) GetCategoryByName(ctx context.Context, tenantID string, categoryName string) (*models.FoodCategory, error) {
	args := m.Called(ctx, tenantID, categoryName)
	return args.Get(0).(*models.FoodCategory), args.Error(1)
}

// =============================================================================
// Create Operation Tests
// =============================================================================

// TestCreateFoodItem_Success tests successful food item creation with all fields
func TestCreateFoodItem_Success(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"

	// Create mock repository
	mockRepo := &MockFoodItemRepository{}
	logger := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
		GetKey:      "",
	}

	// Setup mock to return success
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	cacheClient.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	categoryService := &MockCategoryService{}
	categoryService.On("GetCategoryByName", mock.Anything, mock.Anything, mock.Anything).Return(&models.FoodCategory{
		ID:   "category-1",
		Name: "vegetarian",
	}, nil)

	svc := NewFoodItemService(mockRepo, logger, cacheClient, categoryService)

	// Create request with all required fields
	createReq := &models.FoodItemCreateRequest{
		Name:               "Classic Margherita Pizza",
		Description:        "Traditional Italian pizza with tomato sauce, mozzarella cheese, and fresh basil",
		CategoryName:       "vegetarian", // Will be normalized to "vegetarian"
		Price:              12.99,
		IsVegetarian:       boolPtr(true),
		AvailabilityStatus: "available", // Will be normalized to "available"
	}

	result, err := svc.Create(ctx, createReq, tenantID)

	assert.NoError(t, err, "Create should succeed with valid request")
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, "Classic Margherita Pizza", result.Name, "Name should match")
	assert.Equal(t, "tenant-1", result.TenantID, "Tenant ID should match")
	assert.Equal(t, "category-1", result.CategoryID, "Category ID should be normalized")
	assert.WithinDuration(t, time.Now(), time.Unix(0, result.CreatedAt*int64(time.Millisecond)), 1*time.Second, "Created At should be set")
}

// TestCreateFoodItem_ValidationError tests validation errors on creation
func TestCreateFoodItem_ValidationError(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}

	categoryService := &MockCategoryService{}
	categoryService.On("GetCategoryByName", mock.Anything, mock.Anything, mock.Anything).Return(&models.FoodCategory{
		ID:   "category-1",
		Name: "vegetarian",
	}, nil)

	svc := NewFoodItemService(mockRepo, l, cacheClient, categoryService)

	// Test case 1: Empty name (should fail validation)
	invalidReq := &models.FoodItemCreateRequest{
		Name:               "",
		CategoryName:       "vegetarian",
		Price:              0,
		IsVegetarian:       nil,
		AvailabilityStatus: "available",
	}

	result, err := svc.Create(ctx, invalidReq, tenantID)
	assert.Error(t, err, "Create should fail with empty name")
	assert.Nil(t, result, "Result should be nil on validation error")
	assert.Contains(t, err.Error(), "Validation failed", "Error should indicate validation failure")
}

// TestCreateFoodItem_DatabaseError tests database error handling during creation
func TestCreateFoodItem_DatabaseError(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}

	categoryService := &MockCategoryService{}
	categoryService.On("GetCategoryByName", mock.Anything, mock.Anything, mock.Anything).Return(&models.FoodCategory{
		ID:   "category-1",
		Name: "vegetarian",
	}, nil)
	// Setup mock to return database error
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("duplicate key violation"))

	svc := NewFoodItemService(mockRepo, l, cacheClient, categoryService)

	createReq := &models.FoodItemCreateRequest{
		Name:               "Test Item",
		Description:        "Test Description",
		CategoryName:       "vegetarian",
		Price:              10.0,
		IsVegetarian:       nil,
		AvailabilityStatus: "available",
	}

	result, err := svc.Create(ctx, createReq, tenantID)

	assert.Error(t, err, "Create should fail with database error")
	assert.Nil(t, result, "Result should be nil on database error")
	assert.Contains(t, err.Error(), "Database error creating food item", "Error should indicate database failure")
}

// =============================================================================
// GetByID Operation Tests
// =============================================================================

// TestGetFoodItemByDBCacheHit tests cache hit scenario for get by ID
func TestGetFoodItemByDBCacheHit(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}
	currTime := time.Now().UTC().UnixMilli()
	mockRepo.On("GetByID", mock.Anything, tenantID, "item-123").Return(
		&models.FoodItem{
			ID:                 "item-123",
			TenantID:           "tenant-1",
			Name:               "Delicious Pasta Bolognese",
			Description:        "Hearty Italian beef and tomato pasta dish",
			CategoryID:         "non_vegetarian",
			Price:              14.50,
			IsVegetarian:       false,
			AvailabilityStatus: "available",
			CreatedAt:          currTime,
			UpdatedAt:          currTime,
		},
		nil,
	)
	categoryService := &MockCategoryService{}
	categoryService.On("GetCategoryByName", mock.Anything, mock.Anything, mock.Anything).Return(&models.FoodCategory{
		ID:   "category-1",
		Name: "vegetarian",
	}, nil)
	cacheClient.On("Get", mock.Anything, mock.Anything).Return(&models.FoodItem{
		ID:                 "item-123",
		TenantID:           "tenant-1",
		Name:               "Delicious Pasta Bolognese",
		Description:        "Hearty Italian beef and tomato pasta dish",
		CategoryID:         "non_vegetarian",
		Price:              14.50,
		IsVegetarian:       false,
		AvailabilityStatus: "available",
		CreatedAt:          currTime,
		UpdatedAt:          currTime,
	}, nil)

	svc := NewFoodItemService(mockRepo, l, cacheClient, categoryService)

	result, err := svc.GetByID(ctx, "item-123", tenantID)

	assert.NoError(t, err, "GetByID should succeed with cache hit")
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, "item-123", result.ID, "ID should match requested item ID")
	assert.Equal(t, "Delicious Pasta Bolognese", result.Name, "Name should match")
}

// TestGetFoodItemByDBCacheMiss tests database fallback when cache misses
func TestGetFoodItemByDBCacheMiss(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}
	currTime := time.Now().UTC().UnixMilli()
	mockRepo.On("GetByID", mock.Anything, tenantID, "item-456").Return(
		&models.FoodItem{
			ID:                 "item-456",
			TenantID:           "tenant-1",
			Name:               "Fresh Garden Salad",
			Description:        "Mixed greens with seasonal vegetables",
			CategoryID:         "category-1",
			Price:              9.99,
			IsVegetarian:       true,
			AvailabilityStatus: "available",
			CreatedAt:          currTime,
			UpdatedAt:          currTime,
		},
		nil,
	)
	cacheClient.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("Cache miss"))
	cacheClient.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	categoryService := &MockCategoryService{}
	categoryService.On("GetCategoryByName", mock.Anything, mock.Anything, mock.Anything).Return(&models.FoodCategory{
		ID:   "category-1",
		Name: "vegetarian",
	}, nil)
	svc := NewFoodItemService(mockRepo, l, cacheClient, categoryService)

	result, err := svc.GetByID(ctx, "item-456", tenantID)

	assert.NoError(t, err, "GetByID should succeed with database lookup")
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, "Fresh Garden Salad", result.Name, "Name should match from DB")
}

// TestGetFoodItem_DBNotFound tests handling when item not found in database
func TestGetFoodItem_DBNotFound(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}

	// Mock returns nil from cache (cache miss), then database returns not found
	mockRepo.On("GetByID", mock.Anything, tenantID, "nonexistent-item").Return(nil, errors.New("no rows"))
	cacheClient.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("Cache miss"))
	categoryService := &MockCategoryService{}
	categoryService.On("GetCategoryByName", mock.Anything, mock.Anything, mock.Anything).Return(&models.FoodCategory{
		ID:   "category-1",
		Name: "vegetarian",
	}, nil)
	svc := NewFoodItemService(mockRepo, l, cacheClient, categoryService)

	result, err := svc.GetByID(ctx, "nonexistent-item", tenantID)

	assert.Error(t, err, "GetByID should fail when item not found")
	assert.Nil(t, result, "Result should be nil when item not found")
}

// =============================================================================
// List Operation Tests
// =============================================================================

// TestListFoodItems_Success tests successful list operation with pagination
func TestListFoodItems_Success(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}
	cacheClient.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	currTime := time.Now().UTC().UnixMilli()
	mockRepo.On("ListByTenant", mock.Anything, tenantID, 0, 10).Return(
		[]*models.FoodItem{
			{ID: "pizza-1", TenantID: tenantID, Name: "Pizza Margherita", Description: "Classic", CategoryID: "vegetarian", Price: 12.99, IsVegetarian: true, AvailabilityStatus: "available", CreatedAt: currTime, UpdatedAt: currTime},
			{ID: "pasta-1", TenantID: tenantID, Name: "Spaghetti Carbonara", Description: "Italian classic", CategoryID: "non_vegetarian", Price: 15.99, IsVegetarian: false, AvailabilityStatus: "available", CreatedAt: currTime, UpdatedAt: currTime},
		},
		int64(2), // total count
		nil,
	)

	svc := NewFoodItemService(mockRepo, l, cacheClient, nil)

	result, err := svc.List(ctx, tenantID, 0, 10)

	assert.NoError(t, err, "List should succeed")
	assert.Len(t, result, 2, "Should return 2 items")
}

// TestListFoodItems_Pagination tests pagination behavior
func TestListFoodItems_Pagination(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}
	cacheClient.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	currTime := time.Now().UTC().UnixMilli()
	// Test page 2 with limit 5
	mockRepo.On("ListByTenant", mock.Anything, tenantID, 5, 5).Return(
		[]*models.FoodItem{{ID: "item-3", TenantID: tenantID, Name: "Third Item", CategoryID: "vegan", Price: 8.99, IsVegetarian: true, AvailabilityStatus: "available", CreatedAt: currTime, UpdatedAt: currTime}},
		int64(10), // total count
		nil,
	)

	svc := NewFoodItemService(mockRepo, l, cacheClient, nil)

	result, err := svc.List(ctx, tenantID, 5, 5)

	assert.NoError(t, err, "List should succeed with pagination")
	assert.Len(t, result, 1, "Should return 1 item for page 2 with limit 5")
	assert.Equal(t, "item-3", result[0].ID, "Returned item ID should match expected")
}

// TestListFoodItems_EmptyResult tests listing when no items exist
func TestListFoodItems_EmptyResult(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}

	mockRepo.On("ListByTenant", mock.Anything, tenantID, 0, 10).Return(
		[]*models.FoodItem{},
		int64(0), // total count
		nil,
	)

	svc := NewFoodItemService(mockRepo, l, cacheClient, nil)

	result, err := svc.List(ctx, tenantID, 0, 10)

	assert.NoError(t, err, "List should succeed with empty result")
	assert.Len(t, result, 0, "Should return empty slice")
}

// =============================================================================
// Update Operation Tests
// =============================================================================

// TestUpdateFoodItem_Success tests successful food item update
func TestUpdateFoodItem_Success(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"
	itemID := "pizza-update-test"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}
	currTime := time.Now().UTC().UnixMilli()
	mockRepo.On("GetByID", mock.Anything, tenantID, itemID).Return(
		&models.FoodItem{
			ID:                 itemID,
			TenantID:           "tenant-1",
			Name:               "Old Name",
			Description:        "Old description",
			CategoryID:         "category-2",
			Price:              10.99,
			IsVegetarian:       true,
			AvailabilityStatus: "available",
			CreatedAt:          currTime,
			UpdatedAt:          currTime,
		},
		nil,
	)
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	categoryService := &MockCategoryService{}
	categoryService.On("GetCategoryByName", mock.Anything, mock.Anything, mock.Anything).Return(&models.FoodCategory{
		ID:   "category-2",
		Name: "vegan",
	}, nil)

	svc := NewFoodItemService(mockRepo, l, cacheClient, categoryService)

	updateReq := &models.FoodItemUpdateRequest{
		Name:               "Updated Pizza Name",
		Description:        "Updated description with new details",
		CategoryName:       "vegan",
		Price:              15.99,
		IsVegetarian:       boolPtr(false),
		AvailabilityStatus: "available",
	}

	err := svc.Update(ctx, itemID, updateReq, tenantID)

	assert.NoError(t, err, "Update should succeed")
}

// TestUpdateFoodItem_ValidationFailed tests validation failure on update
func TestUpdateFoodItem_ValidationFailed(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"
	itemID := "pizza-validation-fail"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}
	currTime := time.Now().UTC().UnixMilli()
	mockRepo.On("GetByID", mock.Anything, tenantID, itemID).Return(
		&models.FoodItem{
			ID:                 itemID,
			TenantID:           "tenant-1",
			Name:               "Test Pizza",
			Description:        "Old description",
			CategoryID:         "category-2",
			Price:              10.99,
			IsVegetarian:       true,
			AvailabilityStatus: "available",
			CreatedAt:          currTime,
			UpdatedAt:          currTime,
		},
		nil,
	)

	categoryService := &MockCategoryService{}
	categoryService.On("GetCategoryByName", mock.Anything, mock.Anything, mock.Anything).Return(&models.FoodCategory{
		ID:   "category-2",
		Name: "vegan",
	}, nil)

	svc := NewFoodItemService(mockRepo, l, cacheClient, categoryService)

	// Empty name validation
	updateReq := &models.FoodItemUpdateRequest{
		Name:               "", // Invalid - empty
		CategoryName:       "vegan",
		Price:              10.99,
		IsVegetarian:       nil,
		AvailabilityStatus: "available",
	}

	err := svc.Update(ctx, itemID, updateReq, tenantID)

	assert.Error(t, err, "Update should fail validation")
	assert.Contains(t, err.Error(), "Validation failed", "Error should indicate validation failure")
}

// TestUpdateFoodItem_ItemNotFound tests 404 error when updating non-existent item
func TestUpdateFoodItem_ItemNotFound(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"
	itemID := "nonexistent-item"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}

	mockRepo.On("GetByID", mock.Anything, tenantID, itemID).Return(nil, errors.New("no rows in result set"))

	categoryService := &MockCategoryService{}
	categoryService.On("GetCategoryByName", mock.Anything, mock.Anything, mock.Anything).Return(&models.FoodCategory{
		ID:   "category-1",
		Name: "vegetarian",
	}, nil)

	svc := NewFoodItemService(mockRepo, l, cacheClient, categoryService)

	updateReq := &models.FoodItemUpdateRequest{
		Name:               "New Name",
		Description:        "New description",
		CategoryName:       "vegetarian",
		Price:              10.99,
		IsVegetarian:       nil,
		AvailabilityStatus: "available",
	}

	err := svc.Update(ctx, itemID, updateReq, tenantID)

	assert.Error(t, err, "Update should fail when item not found")
	assert.Contains(t, err.Error(), "Failed to verify", "Error should indicate item not found")
}

// =============================================================================
// Delete Operation Tests
// =============================================================================

// TestDeleteFoodItem_Success tests successful food item deletion
func TestDeleteFoodItem_Success(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"
	itemID := "pizza-delete-test"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}
	currTime := time.Now().UTC().UnixMilli()
	mockRepo.On("GetByID", mock.Anything, tenantID, itemID).Return(
		&models.FoodItem{
			ID:                 itemID,
			TenantID:           "tenant-1",
			Name:               "Test Pizza",
			Description:        "Test description",
			CategoryID:         "vegetarian",
			Price:              10.99,
			IsVegetarian:       true,
			AvailabilityStatus: "available",
			CreatedAt:          currTime,
			UpdatedAt:          currTime,
		},
		nil,
	)
	mockRepo.On("Delete", mock.Anything, itemID).Return(int64(1), nil)

	svc := NewFoodItemService(mockRepo, l, cacheClient, nil)

	rowsAffected, err := svc.Delete(ctx, itemID, tenantID)

	assert.NoError(t, err, "Delete should succeed")
	assert.Equal(t, int64(1), rowsAffected, "Expected one row to be affected")
}

// TestDeleteFoodItem_NotFound tests 404 error when deleting non-existent item
func TestDeleteFoodItem_NotFound(t *testing.T) {
	ctx := t.Context()
	tenantID := "tenant-1"
	itemID := "nonexistent-item"

	mockRepo := &MockFoodItemRepository{}
	l := logger.NewLogger()
	cacheClient := &MockCacheClient{
		CacheHits:   make([]string, 0),
		CacheMisses: make([]string, 0),
		CacheErrors: make([]error, 0),
		CacheSets:   0,
		CacheKeys:   make(map[string]string),
	}

	mockRepo.On("GetByID", mock.Anything, tenantID, itemID).Return(nil, errors.New("no rows in result set"))

	svc := NewFoodItemService(mockRepo, l, cacheClient, nil)

	rowsAffected, err := svc.Delete(ctx, itemID, tenantID)

	assert.Error(t, err, "Delete should fail when item not found")
	assert.Contains(t, err.Error(), "Failed to verify ", "Error should indicate item not found")
	assert.Equal(t, int64(0), rowsAffected, "Expected no rows to be affected")
}

// =============================================================================
// Helper Functions for Tests
// =============================================================================
func boolPtr(b bool) *bool {
	return &b
}
