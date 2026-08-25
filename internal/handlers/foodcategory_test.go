package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	middleware "github.com/pareshvernekar/homecooked/internal/middleware"
	models "github.com/pareshvernekar/homecooked/internal/models"
	foodcategory "github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockFoodCategoryService mocks the FoodCategoryService interface for testing
type MockFoodCategoryService struct {
	mock.Mock
}

func (m *MockFoodCategoryService) ListCategories(ctx context.Context, tenantID string) ([]*models.FoodCategory, error) {
	args := m.Called(tenantID)
	var result []*models.FoodCategory
	if args.Get(0) != nil {
		result = args.Get(0).([]*models.FoodCategory)
	}
	return result, args.Error(1)
}

func (m *MockFoodCategoryService) GetCategoryByID(ctx context.Context, categoryID string, tenantID string) (*models.FoodCategory, error) {
	args := m.Called(categoryID, tenantID)
	var result *models.FoodCategory
	if args.Get(0) != nil {
		result = args.Get(0).(*models.FoodCategory)
	}
	return result, args.Error(1)
}

func (m *MockFoodCategoryService) CreateCategory(ctx context.Context, category *models.FoodCategory) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockFoodCategoryService) UpdateCategory(ctx context.Context, category *models.FoodCategory) (int64, error) {
	args := m.Called(category)
	rowsAffected, _ := args.Get(0).(int64)
	return rowsAffected, nil
}

func (m *MockFoodCategoryService) DeleteCategory(ctx context.Context, tenantID string, id string) (int64, error) {
	m.Called(id)
	rowsAffected := int64(1)
	return rowsAffected, nil
}

// MockLogger mocks the Logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	m.Called(keysAndValues)
}

func (m *MockLogger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	m.Called(keysAndValues)
}

func (m *MockLogger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	m.Called(keysAndValues)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	m.Called(keysAndValues)
}

func (m *MockLogger) Fatal(ctx context.Context, msg string, keysAndValues ...interface{}) {
	m.Called(keysAndValues)
}

// NewFoodCategoryService creates a new FoodCategoryService instance with dependency injection
func NewFoodCategoryService(repo foodcategory.FoodCategoryRepository, l *logger.Logger) *foodcategory.FoodCategoryService {
	return foodcategory.NewFoodCategoryService(repo, l)
}

// MockFoodCategoryRepository mocks the FoodCategoryRepository interface for testing
type MockFoodCategoryRepository struct {
	categories []models.FoodCategory
}

func (m *MockFoodCategoryRepository) ListByTenant(ctx context.Context, tenantID string) ([]*models.FoodCategory, error) {
	var result []*models.FoodCategory
	for _, cat := range m.categories {
		if cat.TenantID == tenantID {
			result = append(result, &cat)
		}
	}
	return result, nil
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

// TestNewFoodCategoryHandler tests handler initialization with dependency injection
func TestNewFoodCategoryHandler(t *testing.T) {
	logger := &logger.Logger{}
	mockService := &MockFoodCategoryService{}

	handler := NewFoodCategoryHandler(mockService, logger)

	require.NotNil(t, handler, "Expected NewFoodCategoryHandler to return a valid handler")
	require.Equal(t, mockService, handler.service, "Expected handler to use injected service")
}

// TestListCategories_Success tests successful retrieval of all food categories
func TestListCategories_Success(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := foodcategory.NewFoodCategoryService(mockRepo, l)
	handler := NewFoodCategoryHandler(service, l)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant_1")
	mockRepo.categories = []models.FoodCategory{
		{ID: "cat1", TenantID: "tenant_1", Name: "vegetarian", Description: "plant-based", CreatedAt: time.Now().UTC().UnixMilli(), UpdatedAt: time.Now().UTC().UnixMilli()},
	}

	handler.ListCategories(c)

	require.Equal(t, http.StatusOK, w.Code)
}

// TestListCategories_EmptyResult tests successful retrieval with empty category list
func TestListCategories_EmptyResult(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := foodcategory.NewFoodCategoryService(mockRepo, l)
	handler := NewFoodCategoryHandler(service, l)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant_1")
	mockRepo.categories = []models.FoodCategory{}

	handler.ListCategories(c)

	require.Equal(t, http.StatusOK, w.Code)
}

// TestGetCategoryByID_Success tests successful retrieval of a single category by ID
func TestGetCategoryByID_Success(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := foodcategory.NewFoodCategoryService(mockRepo, l)
	handler := NewFoodCategoryHandler(service, l)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/cat1", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant_1")
	mockRepo.categories = []models.FoodCategory{
		{ID: "cat1", TenantID: "tenant_1", Name: "vegetarian", Description: "plant-based", CreatedAt: time.Now().UTC().UnixMilli(), UpdatedAt: time.Now().UTC().UnixMilli()},
	}

	handler.GetCategoryByID(c)

	require.Equal(t, http.StatusOK, w.Code)
}

// TestGetCategoryByID_NotFound tests retrieval when category doesn't exist
func TestGetCategoryByID_NotFound(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := foodcategory.NewFoodCategoryService(mockRepo, l)
	handler := NewFoodCategoryHandler(service, l)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/non-existent-id", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant_1")
	mockRepo.categories = []models.FoodCategory{
		{ID: "cat1", TenantID: "tenant_1", Name: "vegetarian", Description: "plant-based", CreatedAt: time.Now().UTC().UnixMilli(), UpdatedAt: time.Now().UTC().UnixMilli()},
	}

	handler.GetCategoryByID(c)

	require.Equal(t, http.StatusNotFound, w.Code)
}

// TestCreateCategory_Success tests successful creation of a new food category
func TestCreateCategory_Success(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := foodcategory.NewFoodCategoryService(mockRepo, l)
	handler := NewFoodCategoryHandler(service, l)

	jsonData := `{"category_name": "desserts", "description": "sweet treats"}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBufferString(jsonData))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant_1")

	handler.CreateCategory(c)

	require.Equal(t, http.StatusOK, w.Code)
}

// TestCreateCategory_InvalidBody tests creation with invalid request body
func TestCreateCategory_InvalidBody(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := foodcategory.NewFoodCategoryService(mockRepo, l)
	handler := NewFoodCategoryHandler(service, l)

	jsonData := `{"invalid": "data"}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBufferString(jsonData))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant_1")

	handler.CreateCategory(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUpdateCategory_Success tests successful update of an existing food category
func TestUpdateCategory_Success(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := foodcategory.NewFoodCategoryService(mockRepo, l)
	handler := NewFoodCategoryHandler(service, l)

	jsonData := `{"name": "updated-desserts", "description": "updated description"}`

	req := httptest.NewRequest(http.MethodPut, "/api/v1/categories/cat1", bytes.NewBufferString(jsonData))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant_1")
	mockRepo.categories = []models.FoodCategory{
		{ID: "cat1", TenantID: "tenant_1", Name: "vegetarian", Description: "plant-based", CreatedAt: time.Now().UTC().UnixMilli(), UpdatedAt: time.Now().UTC().UnixMilli()},
	}

	handler.UpdateCategory(c)

	require.Equal(t, http.StatusOK, w.Code)
}

// TestUpdateCategory_InvalidBody tests update with invalid request body
func TestUpdateCategory_InvalidBody(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := foodcategory.NewFoodCategoryService(mockRepo, l)
	handler := NewFoodCategoryHandler(service, l)

	jsonData := `{"invalid": "data"}`

	req := httptest.NewRequest(http.MethodPut, "/api/v1/categories/cat1", bytes.NewBufferString(jsonData))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant_1")
	mockRepo.categories = []models.FoodCategory{
		{ID: "cat1", TenantID: "tenant_1", Name: "vegetarian", Description: "plant-based", CreatedAt: time.Now().UTC().UnixMilli(), UpdatedAt: time.Now().UTC().UnixMilli()},
	}

	handler.UpdateCategory(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

// TestDeleteCategory_Success tests successful deletion of a food category
func TestDeleteCategory_Success(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := foodcategory.NewFoodCategoryService(mockRepo, l)
	handler := NewFoodCategoryHandler(service, l)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/categories/cat1", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant_1")
	mockRepo.categories = []models.FoodCategory{
		{ID: "cat1", TenantID: "tenant_1", Name: "vegetarian", Description: "plant-based", CreatedAt: time.Now().UTC().UnixMilli(), UpdatedAt: time.Now().UTC().UnixMilli()},
	}

	handler.DeleteCategory(c)

	require.Equal(t, http.StatusNoContent, w.Code)
}

// TestDeleteCategory_InvalidBody tests deletion with missing ID parameter
func TestDeleteCategory_MissingID(t *testing.T) {
	mockRepo := &MockFoodCategoryRepository{}
	l := logger.NewLogger()

	service := foodcategory.NewFoodCategoryService(mockRepo, l)
	handler := NewFoodCategoryHandler(service, l)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/categories", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant_1")
	mockRepo.categories = []models.FoodCategory{
		{ID: "cat1", TenantID: "tenant_1", Name: "vegetarian", Description: "plant-based", CreatedAt: time.Now().UTC().UnixMilli(), UpdatedAt: time.Now().UTC().UnixMilli()},
	}

	handler.DeleteCategory(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
