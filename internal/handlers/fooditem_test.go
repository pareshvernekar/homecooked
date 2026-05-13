package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/models"
)

type MockFoodItemRepository struct{}

func (m *MockFoodItemRepository) Create(foodItem *models.FoodItem) error { return nil }

func (m *MockFoodItemRepository) GetByID(id string) (*models.FoodItem, error) {
	return &models.FoodItem{}, nil
}

func (m *MockFoodItemRepository) Update(foodItem *models.FoodItem) error { return nil }

func (m *MockFoodItemRepository) Delete(id string) error { return nil }

func (m *MockFoodItemRepository) ListByTenant(tenantID string, offset, limit int) ([]models.FoodItem, int64, error) {
	return []models.FoodItem{{}}, 0, nil
}

func TestNewFoodItemHandler(t *testing.T) {
	mockRepo := &MockFoodItemRepository{}
	handler := NewFoodItemHandler(mockRepo)

	if handler == nil {
		t.Error("Expected NewFoodItemHandler to return a valid handler")
	}
}

func TestUpdateFoodItem_Success(t *testing.T) {
	mockRepo := &MockFoodItemRepository{}
	handler := NewFoodItemHandler(mockRepo)

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
	mockRepo := &MockFoodItemRepository{}
	handler := NewFoodItemHandler(mockRepo)

	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/food-items/550e8400-e29b-41d4-a716-446655440003", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant-1")

	handler.DeleteFoodItem(c)
}

func TestCreateFoodItem_InvalidRequest(t *testing.T) {
	mockRepo := &MockFoodItemRepository{}
	handler := NewFoodItemHandler(mockRepo)

	gin.SetMode(gin.TestMode)

	body := `{"name":"","price":0,"category":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/food-items", bytes.NewBufferString(body))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantIDKey, "tenant-1")

	handler.CreateFoodItem(c)
}
