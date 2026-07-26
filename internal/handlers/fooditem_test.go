package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/models"
)

type MockFoodItemService struct{}

func (s *MockFoodItemService) Create(ctx context.Context, createReq *models.FoodItemCreateRequest, tenantID string) (*models.FoodItem, error) {
	return &models.FoodItem{

		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Name:        "Burger",
		Description: "A delicious burger",
		Price:       9.99,
		CategoryID:  "category-1234",
		TenantID:    tenantID,
	}, nil
}

func (s *MockFoodItemService) List(ctx context.Context, tenantID string, page int, limit int) ([]*models.FoodItem, error) {
	return []*models.FoodItem{
		{
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			Name:        "Burger",
			Description: "A delicious burger",
			Price:       9.99,
			CategoryID:  "category-1234",
			TenantID:    tenantID,
		},
	}, nil
}

func (s *MockFoodItemService) GetByID(ctx context.Context, id string, tenantID string) (*models.FoodItem, error) {
	return &models.FoodItem{
		ID:          id,
		Name:        "Burger",
		Description: "A delicious burger",
		Price:       9.99,
		CategoryID:  "category-1234",
		TenantID:    tenantID,
	}, nil
}

func (s *MockFoodItemService) Update(ctx context.Context, id string, updateReq *models.FoodItemUpdateRequest, tenantID string) error {
	return nil
}

func (s *MockFoodItemService) Delete(ctx context.Context, id string, tenantID string) (int64, error) {
	return 1, nil
}
func TestNewFoodItemHandler(t *testing.T) {

	mockFoodItemService := &MockFoodItemService{}

	l := logger.NewLogger()

	handler := NewFoodItemHandler(mockFoodItemService, l)

	if handler == nil {
		t.Error("Expected NewFoodItemHandler to return a valid handler")
	}
}

func TestUpdateFoodItem_Success(t *testing.T) {
	mockFoodItemService := &MockFoodItemService{}
	l := logger.NewLogger()
	handler := NewFoodItemHandler(mockFoodItemService, l)

	gin.SetMode(gin.TestMode)

	body := `{"name":"Updated Burger","description":"An updated description","price":15.99}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/food-items/550e8400-e29b-41d4-a716-446655440001", bytes.NewBufferString(body))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant-1")

	handler.UpdateFoodItem(c)
}

func TestDeleteFoodItem_Success(t *testing.T) {
	mockFoodItemService := &MockFoodItemService{}
	l := logger.NewLogger()
	handler := NewFoodItemHandler(mockFoodItemService, l)

	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/food-items/550e8400-e29b-41d4-a716-446655440003", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant-1")

	handler.DeleteFoodItem(c)
}

func TestCreateFoodItem_InvalidRequest(t *testing.T) {
	mockFoodItemService := &MockFoodItemService{}
	l := logger.NewLogger()
	handler := NewFoodItemHandler(mockFoodItemService, l)

	gin.SetMode(gin.TestMode)

	body := `{"name":"","price":0,"category":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/food-items", bytes.NewBufferString(body))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant-1")

	handler.CreateFoodItem(c)
}
