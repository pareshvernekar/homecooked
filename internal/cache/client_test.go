package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pareshvernekar/homecooked/internal/models"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	repo "github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/stretchr/testify/assert"
)

// MockLogger implements Logger interface for testing
type MockLogger struct {
	Messages []string
}

func (m *MockLogger) Info(ctx context.Context, msg string, keysAndVals ...interface{}) {
	fmt.Printf("%s: %s", msg, keysAndVals[0])
}

func (m *MockLogger) Debug(ctx context.Context, msg string, keysAndVals ...interface{}) {
	fmt.Printf("%s: %s", msg, keysAndVals[0])
}

func (m *MockLogger) Warn(ctx context.Context, msg string, keysAndVals ...interface{}) {
	fmt.Printf("%s: %s", msg, keysAndVals[0])
}

func (m *MockLogger) Error(ctx context.Context, msg string, keysAndVals ...interface{}) {
	fmt.Printf("%s: %s", msg, keysAndVals[0])
}

// MockFoodCategoryRepository is a mock implementation for testing
type MockFoodCategoryRepository struct {
	categories []models.FoodCategory
}

func (m *MockFoodCategoryRepository) ListByTenant(tenantID string) ([]models.FoodCategory, error) {
	return m.categories, nil
}

func (m *MockFoodCategoryRepository) PostInitialize(ctx context.Context, tenantID string, repository repo.FoodCategoryRepository) error {
	return nil
}

// TestNewCacheClient_SuccessfulCacheCreation tests cache client initialization with valid configuration
func TestNewCacheClient_SuccessfulCacheCreation(t *testing.T) {
	catName := "vegetarian"
	ttl := 30 * time.Minute
	config := CacheConfig{
		DefaultTTL:     ttl,
		MaxItems:       1000,
		TTLOverrides: map[string]time.Duration{
			fmt.Sprintf("food_category_%s", catName): ttl,
			"food_category_vegetarian":               ttl,
		},
	}

	logger := logger.NewLogger()
	client, err := NewCacheClient(config, logger, &MockFoodCategoryRepository{})
	assert.NoError(t, err, "Should create cache client successfully")

	key := fmt.Sprintf("food_category_%s", catName)
	val, exists := client.Get(context.Background(), key)
	assert.True(t, exists != nil, "Cache should return value for food category with entity_type:name key format")
	_ = val
}

// TestNewCacheClient_EmptyCategories tests cache client initialization with empty repository
func TestNewCacheClient_EmptyCategories(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:   30 * time.Minute,
		MaxItems:     1000,
	}
	mockRepo := &MockFoodCategoryRepository{
		categories: []models.FoodCategory{},
	}
	logger := logger.NewLogger()
	client, err := NewCacheClient(config, logger, mockRepo)
	assert.NoError(t, err, "Should create cache client successfully")

	ctx := context.Background()

	err = client.PostInitialize(ctx, "tenant_1", mockRepo)
	assert.NoError(t, err, "PostInitialize should succeed even with empty categories")
}

// TestNewCacheClient_MultipleCategories tests cache client initialization with multiple food categories
func TestNewCacheClient_MultipleCategories(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:   30 * time.Minute,
		MaxItems:     1000,
	}

	logger := logger.NewLogger()
	client, err := NewCacheClient(config, logger, &MockFoodCategoryRepository{})
	assert.NoError(t, err)

	ctx := context.Background()

	mockRepo := &MockFoodCategoryRepository{
		categories: []models.FoodCategory{
				{ID: "cat1", TenantID: "tenant_1", Name: "vegetarian", Description: func() *string { s := "plant-based"; return &s }(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: "cat2", TenantID: "tenant_1", Name: "non-vegetarian", Description: func() *string { s := "meat and dairy"; return &s }(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: "cat3", TenantID: "tenant_1", Name: "vegan", Description: func() *string { s := "no animal products"; return &s }(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
	}

	err = client.PostInitialize(ctx, "tenant_1", mockRepo)
	assert.NoError(t, err, "PostInitialize should succeed with multiple categories")
}
